package actions

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/templates"
	"github.com/ddalogin/bicycle-ci/worker"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

// Шаблон страницы билда
type WatchPage struct {
	Project models.Project
	Output  []models.Step
	Build   models.Build
}

// Регистрация роутов для сборок
func BuildsRoutes() {
	http.Handle("/builds/run", auth.RequireAuthentication(run))
	http.Handle("/builds/watch", auth.RequireAuthentication(watch))
}

// Запуск сборки
func run(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	build := models.Build{
		ProjectId: project.Id,
		StartedAt: time.Now().Format("2006-01-02 15:04:05"),
		Status:    models.STATUS_RUNNING,
	}
	build.Save()

	go process(project, build)

	http.Redirect(w, req, "/builds/watch?buildId="+fmt.Sprintf("%v", build.Id), http.StatusSeeOther)
}

// Страница сборки
func watch(w http.ResponseWriter, req *http.Request, user models.User) {
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

	templates.Render(w, "templates/watch.html", WatchPage{
		Project: project,
		Output:  steps,
		Build:   build,
	}, user)
}

// Перенести в воркер
func process(project models.Project, build models.Build) {
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
		buildCmd := exec.Command("bash", "./worker/scripts/build.sh")
		worker.RunStep(project, buildCmd, &buildStep)
		buildStep.Save()
	}

	cleanStep := models.Step{
		BuildId: build.Id,
		Name:    "Cleaning up",
		Status:  models.STEP_STATUS_RUNING,
	}
	cleanStep.Save()
	worker.RunStep(project, exec.Command("bash", "./worker/scripts/clear.sh"), &cleanStep)
	cleanStep.Save()

	if cloneStep.Status == models.STEP_STATUS_SUCCESS && buildStep.Status == models.STEP_STATUS_SUCCESS && cleanStep.Status == models.STEP_STATUS_SUCCESS {
		build.Status = models.STATUS_SUCCESS
	} else {
		build.Status = models.STATUS_FAILED
	}

	build.Save()
}
