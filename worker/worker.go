package worker

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/telegram"
	"os/exec"
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
func (s *Service) RunBuild(project models.Project, commits []string) models.Build {
	build := models.Build{
		ProjectId: project.Id,
		StartedAt: time.Now().Format("2006-01-02 15:04:05"),
		Status:    models.StatusRunning,
	}
	build.Save()

	s.notifyAboutRun(project, build, commits)

	go s.process(project, build)

	return build
}

// Перенести в воркер
func (s *Service) process(project models.Project, build models.Build) {
	//dir, _ := os.Getwd()
	//
	//// Стандартный шаг с копированием репозитория
	//cloneStep := models.BuildStep{
	//	BuildId: build.Id,
	//	Name:    "Загрузка репозитория исходников",
	//	Status:  models.StepStatusRunning,
	//}
	//cloneStep.Save()
	//clone := exec.Command("bash", "./worker/scripts/upload.sh")
	//s.runStep(project, clone, &cloneStep)
	//cloneStep.Save()
	//
	//var buildStep models.BuildStep
	//if cloneStep.Error == "" {
	//	// Запускаем сборку
	//	buildStep = models.BuildStep{
	//		BuildId: build.Id,
	//		Name:    "Сборка проекта",
	//		Status:  models.StepStatusRunning,
	//	}
	//	buildStep.Save()
	//	reg := regexp.MustCompile("\r\n")
	//	plan := reg.ReplaceAllString(*project.BuildPlan, " ")
	//	buildCmd := exec.Command("bash", "-c", "docker run -u 1000:1000 -v "+dir+"/builds/project-"+strconv.Itoa(int(project.Id))+":/app "+project.Image().Name+" sh /build.sh '"+plan+"'")
	//	s.runStep(project, buildCmd, &buildStep)
	//	buildStep.Save()
	//}
	//
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
	//
	//cleanStep := models.BuildStep{
	//	BuildId: build.Id,
	//	Name:    "Удаление артефактов",
	//	Status:  models.StepStatusRunning,
	//}
	//cleanStep.Save()
	//s.runStep(project, exec.Command("bash", "./worker/scripts/clear.sh"), &cleanStep)
	//cleanStep.Save()
	//
	//if cloneStep.Status == models.StepStatusSuccess && buildStep.Status == models.StepStatusSuccess && cleanStep.Status == models.StepStatusSuccess && deployStep.Status == models.StepStatusSuccess {
	//	build.Status = models.StatusSuccess
	//} else {
	//	build.Status = models.StatusFailed
	//}
	//
	//endTime := time.Now().Format("2006-01-02 15:04:05")
	//build.EndedAt = &endTime
	//build.Save()
	//
	//s.notifyAboutResult(project, build)
}

// Выполнить этап билда
func (s *Service) runStep(project models.Project, cmd *exec.Cmd, result *models.BuildStep) {
	//var stdout, stderr bytes.Buffer
	//var env []string
	//env = append(env, "ID="+strconv.Itoa(int(project.Id)))
	//env = append(env, "NAME="+project.Name)
	//env = append(env, "DEPLOY_DIR="+strings.TrimSpace(*project.DeployDir))
	//env = append(env, "ARTIFACT_DIR="+strings.TrimSpace(*project.ArtifactDir))
	//env = append(env, "SSH_KEY="+*project.DeployPrivate)

	//if *project.ServerId != 0 && project.ServerId != nil {
	//	server := models.GetServerById(*project.ServerId)
	//	env = append(env, "USER="+server.Login)
	//	env = append(env, "HOST="+server.Host)
	//	env = append(env, "SSH_KEY_REMOTE="+server.DeployPrivate)
	//}
	//
	//cmd.Env = env
	//cmd.Stdout = &stdout
	//cmd.Stderr = &stderr
	//err := cmd.Run()
	//if err != nil {
	//	result.Error = err.Error()
	//	result.Status = models.StepStatusFailed
	//}
	//cmd.Wait()
	//result.StdOut = string(stdout.Bytes())
	//result.StdErr = string(stderr.Bytes())
	//
	//if result.Status == models.StepStatusRunning {
	//	result.Status = models.StepStatusSuccess
	//}

	return
}

// Уведомляет о начале сборки
func (s *Service) notifyAboutRun(project models.Project, build models.Build, commits []string) {
	buildUrl := s.host + "/builds/status?buildId=" + fmt.Sprintf("%v", build.Id)
	message := `[#` + strconv.Itoa(int(build.Id)) + ` Начата сборка проекта \"` + project.Name + `\".](` + buildUrl + `) \r\n`
	message = message + "\xE2\x8F\xB3 Приблизительное время сборки: " + project.GetAvgBuildTime() + " \r\n"

	if len(commits) > 0 {
		message = message + `*Комиты попавшие в сборку* \r\n`
		for _, commit := range commits {
			message = message + commit + "\r\n"
		}
	}

	s.telegram.SendMessage(message)
}

// Уведомляет о результате сборки
func (s *Service) notifyAboutResult(project models.Project, build models.Build) {
	buildUrl := s.host + "/builds/status?buildId=" + fmt.Sprintf("%v", build.Id)
	message := `[#` + strconv.Itoa(int(build.Id)) + ` сборка проекта \"` + project.Name + `\" завершилась.](` + buildUrl + `) \r\n`
	message = message + `*Статус шагов*: \r\n`

	steps := models.GetStepsByBuildId(build.Id)
	for _, step := range steps {
		status := "В процессе"

		if step.Status == models.StepStatusFailed {
			status = "Остановлен с ошибкой"
		} else if step.Status == models.StepStatusSuccess {
			status = "Успешно"
		}

		message = message + step.Name + ": " + status + "\r\n"
	}

	buildStatus := "В процессе"

	if build.Status == models.StatusFailed {
		buildStatus = "Остановлена с ошибкой"
	} else if build.Status == models.StatusSuccess {
		buildStatus = "Успешна"
	}

	message = message + `*Статус сборки*: ` + buildStatus + `\r\n`

	s.telegram.SendMessage(message)
}
