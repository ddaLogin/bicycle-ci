package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"bytes"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

// Шаблон страницы выполнения билда
type RunPage struct {
	Project models.Project
	Output  []StepResult
}

// Результаты запуска шага
type StepResult struct {
	StepName string
	Error    error
	StdOut   string
	StdErr   string
}

// Регистрация роутов для сборок
func BuildsRoutes() {
	http.Handle("/builds/run", auth.RequireAuthentication(run))
}

// Запуск сборки
func run(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	page := RunPage{
		Project: project,
	}

	// Стандартный шаг с копированием репозитория
	upload := exec.Command("bash", "./scripts/upload.sh")
	uploadResult := runStep(project, upload)
	uploadResult.StepName = "Cloning repository"
	page.Output = append(page.Output, uploadResult)

	if uploadResult.Error == nil {
		//Запускаем сборку
		build := exec.Command("bash", "./scripts/build.sh")
		buildResult := runStep(project, build)
		buildResult.StepName = "Build project"
		page.Output = append(page.Output, buildResult)
	}

	cleanResult := runStep(project, exec.Command("bash", "./scripts/clear.sh"))
	cleanResult.StepName = "Cleaning up"
	page.Output = append(page.Output, cleanResult)

	templates.Render(w, "templates/run.html", page, user)
}

// Выполнить этап билда
func runStep(project models.Project, cmd *exec.Cmd) (result StepResult) {
	var stdout, stderr bytes.Buffer
	var env []string
	env = append(env, "ID="+strconv.Itoa(int(project.Id)))
	env = append(env, "NAME="+project.Name)
	env = append(env, "PLAN="+strings.TrimSpace(*project.Plan))
	env = append(env, "SSHKEY="+*project.DeployPrivate)

	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		result.Error = err
	}
	cmd.Wait()
	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())

	return
}

//// Выполнение комманды
//func buildProject(result *StepResult) {
//	cmd := exec.StepName("bash", "-c", result.StepName)
//	var stdout, stderr bytes.Buffer
//	cmd.Stdout = &stdout
//	cmd.Stderr = &stderr
//	err := cmd.Run()
//	if err != nil {
//		result.Error = err
//	}
//	cmd.Wait()
//	result.StdOut = string(stdout.Bytes())
//	result.StdErr = string(stderr.Bytes())
//}
