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
	"strings"
	"time"
)

// Начинает сборку прокета
func RunProcess(project models.Project, payload HookPayload) models.Build {
	build := models.Build{
		ProjectId: project.Id,
		StartedAt: time.Now().Format("2006-01-02 15:04:05"),
		Status:    models.STATUS_RUNNING,
	}
	build.Save()

	notifyStartBuild(project, build, payload)

	go process(project, build)

	return build
}

// Перенести в воркер
func process(project models.Project, build models.Build) {
	dir, _ := os.Getwd()

	// Стандартный шаг с копированием репозитория
	cloneStep := models.Step{
		BuildId: build.Id,
		Name:    "Cloning repository",
		Status:  models.STEP_STATUS_RUNING,
	}
	cloneStep.Save()
	clone := exec.Command("bash", "./worker/scripts/upload.sh")
	worker.RunStep(project, clone, &cloneStep)
	cloneStep.Save()

	var buildStep models.Step
	if cloneStep.Error == "" {
		//Запускаем сборку
		buildStep = models.Step{
			BuildId: build.Id,
			Name:    "Build project",
			Status:  models.STEP_STATUS_RUNING,
		}
		buildStep.Save()
		reg := regexp.MustCompile("\r\n")
		plan := reg.ReplaceAllString(*project.BuildPlan, " ")
		buildCmd := exec.Command("bash", "-c", "docker run -u 1000:1000 -v "+dir+"/builds/project-"+strconv.Itoa(int(project.Id))+":/app "+project.Image().Name+" sh /build.sh '"+plan+"'")
		worker.RunStep(project, buildCmd, &buildStep)
		buildStep.Save()
	}

	var deployStep models.Step
	if buildStep.Error == "" {
		deployStep = models.Step{
			BuildId: build.Id,
			Name:    "Deploy project",
			Status:  models.STEP_STATUS_RUNING,
		}
		deployStep.Save()

		if *project.ServerId == 0 || project.ServerId == nil {
			worker.RunStep(project, exec.Command("bash", "./worker/scripts/deploy.sh"), &deployStep)
		} else {
			worker.RunStep(project, exec.Command("bash", "./worker/scripts/deploy_remote.sh"), &deployStep)
		}

		deployStep.Save()
	}

	cleanStep := models.Step{
		BuildId: build.Id,
		Name:    "Cleaning up",
		Status:  models.STEP_STATUS_RUNING,
	}
	cleanStep.Save()
	worker.RunStep(project, exec.Command("bash", "./worker/scripts/clear.sh"), &cleanStep)
	cleanStep.Save()

	if cloneStep.Status == models.STEP_STATUS_SUCCESS && buildStep.Status == models.STEP_STATUS_SUCCESS && cleanStep.Status == models.STEP_STATUS_SUCCESS && deployStep.Status == models.STEP_STATUS_SUCCESS {
		build.Status = models.STATUS_SUCCESS
	} else {
		build.Status = models.STATUS_FAILED
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	build.EndedAt = &endTime
	build.Save()

	notifyResultBuild(project, build)
}

// Уведомляет о начале сборки
func notifyStartBuild(project models.Project, build models.Build, payload HookPayload) {
	buildUrl := Host + "/builds/status?buildId=" + fmt.Sprintf("%v", build.Id)
	message := `[#` + strconv.Itoa(int(build.Id)) + ` Начата сборка проекта \"` + project.Name + `\".](` + buildUrl + `) \r\n`
	message = message + "\xE2\x8F\xB3 Приблизительное время сборки: " + project.GetAvgBuildTime() + " \r\n"

	if payload.Ref != "" {
		message = message + `*Комиты попавшие в сборку* \r\n`
		for _, commit := range payload.Commits {
			message = message + commit.Message + "\r\n"
		}
	}

	telegram.SendMessage(message)
}

// Уведомляет о результате сборки
func notifyResultBuild(project models.Project, build models.Build) {
	buildUrl := Host + "/builds/status?buildId=" + fmt.Sprintf("%v", build.Id)
	message := `[#` + strconv.Itoa(int(build.Id)) + ` сборка проекта \"` + project.Name + `\" завершилась.](` + buildUrl + `) \r\n`
	message = message + `*Статус шагов*: \r\n`

	steps := models.GetStepsByBuildId(build.Id)
	for _, step := range steps {
		status := "Running"

		if step.Status == models.STEP_STATUS_FAILED {
			status = "Failed"
		} else if step.Status == models.STEP_STATUS_SUCCESS {
			status = "Success"
		}

		message = message + step.Name + ": " + status + "\r\n"
	}

	buildStatus := "Running"

	if build.Status == models.STATUS_FAILED {
		buildStatus = "Failed"
	} else if build.Status == models.STATUS_SUCCESS {
		buildStatus = "Success"
	}

	message = message + `*Статус сборки*: ` + buildStatus + `\r\n`

	telegram.SendMessage(message)
}

// Выполнить этап билда
func RunStep(project models.Project, cmd *exec.Cmd, result *models.Step) {
	var stdout, stderr bytes.Buffer
	var env []string
	env = append(env, "ID="+strconv.Itoa(int(project.Id)))
	env = append(env, "NAME="+project.Name)
	env = append(env, "DEPLOY_DIR="+strings.TrimSpace(*project.DeployDir))
	env = append(env, "ARTIFACT_DIR="+strings.TrimSpace(*project.ArtifactDir))
	env = append(env, "SSH_KEY="+*project.DeployPrivate)

	if *project.ServerId != 0 && project.ServerId != nil {
		server := models.GetServerById(*project.ServerId)
		env = append(env, "USER="+server.Login)
		env = append(env, "HOST="+server.Host)
		env = append(env, "SSH_KEY_REMOTE="+server.DeployPrivate)
	}

	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		result.Error = err.Error()
		result.Status = models.STEP_STATUS_FAILED
	}
	cmd.Wait()
	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())

	if result.Status == models.STEP_STATUS_RUNING {
		result.Status = models.STEP_STATUS_SUCCESS
	}

	return
}
