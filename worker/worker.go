package worker

import (
	"bytes"
	"fmt"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/telegram"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// Сервик запуска сборок
type Service struct {
	telegram *telegram.Service
	host     string
}

// Конструктор сервиса запуска сборок
func NewService(telegram *telegram.Service, host string, port string) *Service {
	return &Service{telegram: telegram, host: host + ":" + port}
}

// Запускает сборку проекта
func (s *Service) RunBuild(buildPlan *models.ProjectBuildPlan, user *models.User, commits []string) models.Build {
	build := models.Build{
		ProjectBuildPlanId: buildPlan.Id,
		UserId:             user.Id,
		StartedAt:          time.Now().Format("2006-01-02 15:04:05"),
		Status:             models.StatusRunning,
	}
	build.Save()

	s.telegram.SendMessage(build.GetStartMessage(s.host, commits))
	go s.process(build)

	return build
}

// Перенести в воркер
func (s *Service) process(build models.Build) {
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
	s.runStep(clone, cloneStep)
	cloneStep.Save()

	var buildStep models.BuildStep
	if cloneStep.Error == "" {
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
		buildCmd := exec.Command("bash", "-c", "docker run -u 1000:1000 -v "+dir+"/builds/project-"+strconv.Itoa(int(plan.ProjectId))+":/app "+plan.GetDockerImage().Name+" sh /build.sh '"+instructions+"'")
		s.runStep(buildCmd, &buildStep)
		buildStep.Save()
	}

	//var deployStep models.BuildStep
	//if buildStep.Error == "" {
	//	deployStep = models.BuildStep{
	//		BuildId: build.Id,
	//		Name:    "Развертывание проекта",
	//		Status:  models.StepStatusRunning,
	//	}
	//	deployStep.Save()
	//
	//	if *project.ServerId == 0 || project.ServerId == nil {
	//		s.runStep(project, exec.Command("bash", "./worker/scripts/deploy.sh"), &deployStep)
	//	} else {
	//		s.runStep(project, exec.Command("bash", "./worker/scripts/deploy_remote.sh"), &deployStep)
	//	}
	//
	//	deployStep.Save()
	//}

	cleanStep := models.BuildStep{
		BuildId: build.Id,
		Name:    "Удаление артефактов",
		Status:  models.StepStatusRunning,
	}
	cleanStep.SetBuild(&build)
	cleanStep.Save()
	s.runStep(exec.Command("bash", "./worker/scripts/clear.sh"), &cleanStep)
	cleanStep.Save()

	if cloneStep.Status == models.StepStatusSuccess && buildStep.Status == models.StepStatusSuccess && cleanStep.Status == models.StepStatusSuccess {
		build.Status = models.StatusSuccess
	} else {
		build.Status = models.StatusFailed
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	build.EndedAt = &endTime
	build.Save()

	s.telegram.SendMessage(build.GetCompleteMessage(s.host))
}

// Выполнить этап билда
func (s *Service) runStep(cmd *exec.Cmd, result *models.BuildStep) {
	project := result.GetBuild().GetProjectBuildPlan().GetProject()
	var stdout, stderr bytes.Buffer
	var env []string
	env = append(env, fmt.Sprintf("ID=%d", project.Id))
	env = append(env, fmt.Sprintf("NAME=%s", project.Name))
	env = append(env, fmt.Sprintf("SSH_KEY=%s", *project.DeployPrivate))
	env = append(env, fmt.Sprintf("ARTIFACT_DIR=%s", result.GetBuild().GetProjectBuildPlan().Artifact))

	//env = append(env, "DEPLOY_DIR="+strings.TrimSpace(*project.DeployDir))

	//if *project.ServerId != 0 && project.ServerId != nil {
	//	server := models.GetServerById(*project.ServerId)
	//	env = append(env, "USER="+server.Login)
	//	env = append(env, "HOST="+server.Host)
	//	env = append(env, "SSH_KEY_REMOTE="+server.DeployPrivate)
	//}

	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		result.Error = err.Error()
		result.Status = models.StepStatusFailed
	}
	cmd.Wait()
	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())

	if result.Status == models.StepStatusRunning {
		result.Status = models.StepStatusSuccess
	}

	return
}
