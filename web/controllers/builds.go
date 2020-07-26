package controllers

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/telegram"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"github.com/ddalogin/bicycle-ci/worker"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// Контроллер сборок
type BuildsController struct {
	auth *auth.Service
}

// Конструктор контроллера сборок
func NewBuildsController(auth *auth.Service) *BuildsController {
	return &BuildsController{auth: auth}
}

// Шаблон страницы сборки
type StatusPage struct {
	Project models.Project
	Output  []models.Step
	Build   models.Build
}

// Запуск сборки
func (c *BuildsController) Run(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	build := RunProcess(project, HookPayload{})

	http.Redirect(w, req, "/builds/status?buildId="+fmt.Sprintf("%v", build.Id), http.StatusSeeOther)
}

// Страница сборки
func (c *BuildsController) Status(w http.ResponseWriter, req *http.Request, user models.User) {
	buildId := req.URL.Query().Get("buildId")
	build := models.GetBuildById(buildId)

	if (models.Build{}) == build {
		http.NotFound(w, req)
		return
	}

	project := models.GetProjectById(strconv.Itoa(int(build.ProjectId)))

	if (models.Project{}) == project {
		http.NotFound(w, req)
		return
	}

	steps := models.GetStepsByBuildId(build.Id)

	templates.Render(w, "web/templates/status.html", StatusPage{
		Project: project,
		Output:  steps,
		Build:   build,
	}, user)
}
