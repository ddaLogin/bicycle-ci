package worker

import (
	"bytes"
	"fmt"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/telegram"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Сервик запуска сборок
type Service struct {
	telegram        *telegram.Service
	host            string
	artifactPerPlan int // Максимальное кол-во сохраняемый архивов на 1 билдплан
}

// Конструктор сервиса запуска сборок
func NewService(telegram *telegram.Service, host string, port string, artifactPerPlan int) *Service {
	return &Service{telegram: telegram, host: host + ":" + port, artifactPerPlan: artifactPerPlan}
}

// Запускает сборку проекта
func (s *Service) RunBuild(buildPlan *models.ProjectBuildPlan, user *models.User, branch string, commits []string) models.Build {
	build := models.Build{
		ProjectBuildPlanId: buildPlan.Id,
		UserId:             user.Id,
		Branch:             branch,
		StartedAt:          time.Now().Format("2006-01-02 15:04:05"),
		Status:             models.BuildStatusRunning,
	}
	build.Save()

	s.telegram.SendMessage(build.GetStartMessage(s.host, commits))
	go s.processBuild(build)

	s.clearOldArtifacts(s.artifactPerPlan, &build)

	return build
}

// Перенести в воркер
func (s *Service) processBuild(build models.Build) {
	plan := build.GetProjectBuildPlan()
	dir, _ := os.Getwd()

	// Стандартный шаг с копированием репозитория
	cloneStep := &models.BuildStep{
		BuildId: build.Id,
		Name:    "Загрузка исходников",
		Status:  models.StepStatusRunning,
	}
	cloneStep.SetBuild(&build)
	cloneStep.Save()
	clone := exec.Command("bash", "./worker/scripts/upload.sh")
	s.runBuildStep(clone, cloneStep)
	cloneStep.Save()

	var buildStep models.BuildStep
	if cloneStep.Status == models.StepStatusSuccess {
		// Запускаем сборку
		buildStep = models.BuildStep{
			BuildId: build.Id,
			Name:    "Сборка проекта",
			Status:  models.StepStatusRunning,
		}
		buildStep.SetBuild(&build)
		buildStep.Save()
		reg := regexp.MustCompile("\r\n")
		instructions := reg.ReplaceAllString(plan.BuildInstruction, " ")
		buildCmd := exec.Command("bash", "-c", "docker run -v "+dir+"/builds/project-"+strconv.Itoa(int(plan.ProjectId))+":/app "+plan.GetDockerImage().Name+" sh /build.sh '"+instructions+"'")
		s.runBuildStep(buildCmd, &buildStep)
		buildStep.Save()
	}

	var artifactStep models.BuildStep
	if buildStep.Status == models.StepStatusSuccess {
		artifactStep = models.BuildStep{
			BuildId: build.Id,
			Name:    "Упаковка артефактов",
			Status:  models.StepStatusRunning,
		}
		artifactStep.SetBuild(&build)
		artifactStep.Save()

		s.runBuildStep(exec.Command("bash", "./worker/scripts/packaging.sh"), &artifactStep)
		artifactStep.Save()
	}

	cleanStep := models.BuildStep{
		BuildId: build.Id,
		Name:    "Очистка",
		Status:  models.StepStatusRunning,
	}
	cleanStep.SetBuild(&build)
	cleanStep.Save()
	s.runBuildStep(exec.Command("bash", "./worker/scripts/clear.sh"), &cleanStep)
	cleanStep.Save()

	if cloneStep.Status == models.StepStatusSuccess && buildStep.Status == models.StepStatusSuccess && cleanStep.Status == models.StepStatusSuccess && artifactStep.Status == models.StepStatusSuccess {
		build.Status = models.BuildStatusSuccess
	} else {
		build.Status = models.BuildStatusFailed
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	build.EndedAt = &endTime
	build.Save()

	s.telegram.SendMessage(build.GetCompleteMessage(s.host))
}

// Выполнить этап билда
func (s *Service) runBuildStep(cmd *exec.Cmd, result *models.BuildStep) {
	buildPlan := result.GetBuild().GetProjectBuildPlan()
	project := buildPlan.GetProject()
	var stdout, stderr bytes.Buffer
	var env []string
	env = append(env, fmt.Sprintf("ID=%d", project.Id))
	env = append(env, fmt.Sprintf("NAME=%s", project.Name))
	env = append(env, fmt.Sprintf("BRANCH=%s", result.GetBuild().Branch))
	env = append(env, fmt.Sprintf("SSH_KEY=%s", *project.DeployPrivate))
	env = append(env, fmt.Sprintf("ARTIFACT_DIR=%s", buildPlan.Artifact))
	env = append(env, fmt.Sprintf("ARTIFACT_ZIP_NAME=%s", result.GetBuild().GetArtifactName()))

	cmd.Env = env
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stdout, &stderr)
	err := cmd.Run()
	if err != nil {
		result.Error = err.Error()
		result.Status = models.StepStatusFailed
	}

	err = cmd.Wait()
	if err != nil {
		log.Println("Выполнение шага завершилось с ошибкой", err, cmd)
	}

	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())

	if result.Status == models.StepStatusRunning {
		result.Status = models.StepStatusSuccess
	}

	return
}

// Удаляет старые артефакты билд плана, если их больше максимума
func (s *Service) clearOldArtifacts(maxCount int, build *models.Build) {
	pattern := fmt.Sprintf("builds/artifact_%d_%d_*.zip", build.GetProjectBuildPlan().ProjectId, build.ProjectBuildPlanId)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println(err)
	}

	if len(matches) > maxCount {
		for _, file := range matches[:len(matches)-maxCount] {
			err := os.Remove(file)
			if err != nil {
				log.Fatal("Не удалось удалить старый артефакт сборки", err)
			}
		}
	}
}

// Запуск деплоймента
func (s *Service) RunDeployment(plan *models.ProjectDeployPlan, build *models.Build, user *models.User) models.Deploy {
	deploy := models.Deploy{
		ProjectDeployPlanId: plan.Id,
		BuildId:             build.Id,
		UserId:              user.Id,
		StartedAt:           time.Now().Format("2006-01-02 15:04:05"),
		Status:              models.DeployStatusRunning,
	}
	deploy.Save()

	s.telegram.SendMessage(deploy.GetStartMessage(s.host))

	go s.processDeploy(plan, &deploy, build)

	return deploy
}

// Выполнение деплоймента
func (s *Service) processDeploy(plan *models.ProjectDeployPlan, deploy *models.Deploy, build *models.Build) {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd
	var env []string

	env = append(env, fmt.Sprintf("ID=%d", plan.GetProject().Id))
	env = append(env, fmt.Sprintf("DEPLOY_DIR=%s", strings.TrimSpace(plan.DeploymentDirectory)))
	env = append(env, fmt.Sprintf("ARTIFACT_ZIP=builds/%s", build.GetArtifactName()))

	if plan.RemoteServerId == nil || *plan.RemoteServerId == 0 {
		cmd = exec.Command("bash", "./worker/scripts/deploy.sh")
	} else {
		cmd = exec.Command("bash", "./worker/scripts/deploy_remote.sh")

		server := models.GetServerById(*plan.RemoteServerId)
		env = append(env, "USER="+server.Login)
		env = append(env, "HOST="+server.Host)
		env = append(env, "SSH_KEY_REMOTE="+server.DeployPrivate)
	}

	cmd.Env = env
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stdout, &stderr)
	err := cmd.Run()
	if err != nil {
		deploy.Error = err.Error()
		deploy.Status = models.DeployStatusFailed
	}

	err = cmd.Wait()
	if err != nil {
		log.Println("Выполнение релиза завершилось с ошибкой", err, cmd)
	}

	deploy.StdOut = string(stdout.Bytes())
	deploy.StdErr = string(stderr.Bytes())

	if deploy.Status == models.DeployStatusRunning {
		deploy.Status = models.DeployStatusSuccess
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	deploy.EndedAt = &endTime
	deploy.Save()

	s.telegram.SendMessage(deploy.GetCompleteMessage(s.host))
}
