package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"net/http"
	"os/exec"
)

// Шаблон страницы выполнения билда
type RunPage struct {
	Output string
}

// Регистрация роутов для сборок
func BuildsRoutes() {
	http.Handle("/builds/run", auth.RequireAuthentication(run))
}

// Запуск сборки
func run(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	page := RunPage{}

	cmd := exec.Command("sh", "-c", *project.Plan)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		page.Output = err.Error()
	}
	page.Output = string(stdoutStderr)
	templates.Render(w, "templates/Run.html", page, user)
}
