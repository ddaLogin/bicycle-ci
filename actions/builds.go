package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"bicycle-ci/worker"
	"fmt"
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
	clone := exec.Command("bash", "./worker/scripts/upload.sh")
	cloneStep := worker.RunStep(project, clone)
	cloneStep.Name = "Cloning repository"
	cloneStep.BuildId = build.Id
	cloneStep.Save()

	var buildStep models.Step
	if cloneStep.Error == "" {
		//Запускаем сборку
		buildCmd := exec.Command("bash", "./worker/scripts/build.sh")
		buildStep = worker.RunStep(project, buildCmd)
		buildStep.Name = "Build project"
		buildStep.BuildId = build.Id
		buildStep.Save()
	}

	cleanStep := worker.RunStep(project, exec.Command("bash", "./worker/scripts/clear.sh"))
	cleanStep.Name = "Cleaning up"
	cleanStep.BuildId = build.Id
	cleanStep.Save()

	if cloneStep.Status == models.STEP_STATUS_SUCCESS && buildStep.Status == models.STEP_STATUS_SUCCESS && cleanStep.Status == models.STEP_STATUS_SUCCESS {
		build.Status = models.STATUS_SUCCESS
	} else {
		build.Status = models.STATUS_FAILED
	}

	build.Save()
}
