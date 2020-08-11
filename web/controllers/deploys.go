package controllers

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"github.com/ddalogin/bicycle-ci/worker"
	"net/http"
)

// Контроллер релизов
type DeploysController struct {
	auth          *auth.Service
	workerService *worker.Service
}

// Конструктор контроллера сборок
func NewDeploysController(auth *auth.Service, workerService *worker.Service) *DeploysController {
	return &DeploysController{auth: auth, workerService: workerService}
}

// Шаблон страницы статуса релиза
type DeployStatusPage struct {
	Deploy *models.Deploy
}

// Запуск развертывания (деплоймента)
func (c *DeploysController) Release(w http.ResponseWriter, req *http.Request, user *models.User) {
	deployPlan := models.GetProjectDeployPlanById(req.URL.Query().Get("id"))

	if deployPlan == nil || (models.ProjectDeployPlan{}) == *deployPlan {
		http.NotFound(w, req)
		return
	}

	build := models.GetBuildById(req.URL.Query().Get("buildId"))

	if build == nil || (models.Build{}) == *build {
		http.NotFound(w, req)
		return
	}

	if build.IsArtifactExists() == false {
		http.NotFound(w, req)
		return
	}

	project := deployPlan.GetProject()

	if project == nil || (models.Project{}) == *project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if build.GetProjectBuildPlan().ProjectId != project.Id {
		http.NotFound(w, req)
		return
	}

	deploy := c.workerService.RunDeployment(deployPlan, build, user)

	http.Redirect(w, req, fmt.Sprintf("/deployments/status?id=%d", deploy.Id), http.StatusSeeOther)
}

// Страница релиза
func (c *DeploysController) Status(w http.ResponseWriter, req *http.Request, user *models.User) {
	deploy := models.GetDeployById(req.URL.Query().Get("id"))

	if deploy == nil || (models.Deploy{}) == *deploy {
		http.NotFound(w, req)
		return
	}

	project := deploy.GetProjectDeployPlan().GetProject()

	if project == nil || (models.Project{}) == *project {
		http.NotFound(w, req)
		return
	}

	templates.Render(w, "web/templates/status_deploy.html", DeployStatusPage{
		Deploy: deploy,
	}, user)
}
