package controllers

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"github.com/ddalogin/bicycle-ci/worker"
	"net/http"
)

// Контроллер сборок
type BuildsController struct {
	auth          *auth.Service
	workerService *worker.Service
}

// Конструктор контроллера сборок
func NewBuildsController(auth *auth.Service, workerService *worker.Service) *BuildsController {
	return &BuildsController{auth: auth, workerService: workerService}
}

// Шаблон страницы сборки
type StatusPage struct {
	Project     *models.Project
	Steps       []*models.BuildStep
	Build       *models.Build
	DeployPlans []*models.ProjectDeployPlan
}

// Запуск сборки
func (c *BuildsController) Run(w http.ResponseWriter, req *http.Request, user *models.User) {
	buildPlan := models.GetProjectBuildPlanById(req.URL.Query().Get("id"))

	if buildPlan == nil || (models.ProjectBuildPlan{}) == *buildPlan {
		http.NotFound(w, req)
		return
	}

	project := buildPlan.GetProject()

	if project == nil || (models.Project{}) == *project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	build := c.workerService.RunBuild(buildPlan, user, "master", nil)

	http.Redirect(w, req, fmt.Sprintf("/builds/status?buildId=%d", build.Id), http.StatusSeeOther)
}

// Страница сборки
func (c *BuildsController) Status(w http.ResponseWriter, req *http.Request, user *models.User) {
	build := models.GetBuildById(req.URL.Query().Get("buildId"))

	if build == nil || (models.Build{}) == *build {
		http.NotFound(w, req)
		return
	}

	project := build.GetProjectBuildPlan().GetProject()

	if project == nil || (models.Project{}) == *project {
		http.NotFound(w, req)
		return
	}

	steps := models.GetStepsByBuildId(build.Id)

	templates.Render(w, "web/templates/status.html", StatusPage{
		Project:     project,
		Steps:       steps,
		Build:       build,
		DeployPlans: models.GetProjectDeployPlansByProjectId(project.Id),
	}, user)
}

// Метод для скачивания архива артефактов сборки
func (c *BuildsController) Artifact(w http.ResponseWriter, req *http.Request, user *models.User) {
	build := models.GetBuildById(req.URL.Query().Get("id"))

	if build == nil || (models.Build{}) == *build || build.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if build.IsArtifactExists() == false {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, req, "builds/"+build.GetArtifactName())
}
