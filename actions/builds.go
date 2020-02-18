package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"bytes"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// Шаблон страницы выполнения билда
type RunPage struct {
	Project models.Project
	Output  []StepResult
}

// Результаты запуска шага
type StepResult struct {
	Command string
	Error   error
	StdOut  string
	StdErr  string
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
	commands := strings.Split(*project.Plan, "\r\n")

	for _, command := range commands {
		result := StepResult{
			Command: command,
		}

		makeStep(&result)
		page.Output = append(page.Output, result)
		if result.Error != nil {
			log.Println("Error while make step. ", result.Error)
			break
		}
	}

	templates.Render(w, "templates/Run.html", page, user)
}

// Выполнение комманды
func makeStep(result *StepResult) {
	cmd := exec.Command("sh", "-c", result.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		result.Error = err
	}
	cmd.Wait()
	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())
}
