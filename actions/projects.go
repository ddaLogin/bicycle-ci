package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/providers"
	"bicycle-ci/templates"
	"net/http"
)

// Страница списка проектов
type ProjectListPage struct {
	Projects []models.Project
}

// Страница активации проектов
type ProjectEnablePage struct {
	ProjectsToEnable map[int]*models.Project
}

// Страница редактирования плана
type ProjectPlanPage struct {
	Project models.Project
	Message string
}

// Регистрация роутов по проектам
func ProjectRoutes() {
	http.Handle("/projects/list", auth.RequireAuthentication(projectsList))
	http.Handle("/projects/choose", auth.RequireAuthentication(projectsChoose))
	http.Handle("/projects/enable", auth.RequireAuthentication(projectsEnable))
	http.Handle("/projects/plan", auth.RequireAuthentication(projectsPlan))
}

// Страница проектов пользователя
func projectsList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница выбора репозитория для нового проекта
func projectsChoose(w http.ResponseWriter, req *http.Request, user models.User) {
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := providers.GetProviderByType(providerData.ProviderType)

	if provider == nil {
		http.NotFound(w, req)
		return
	}

	provider.SetProviderData(providerData)
	projectsToEnable := provider.LoadProjects()

	for _, value := range models.GetProjectsByUserId(user.Id) {
		if val, ok := projectsToEnable[value.RepoId]; ok {
			val.Status = models.STATUS_ENABLED
		}
	}

	templates.Render(w, "templates/projects/choose.html", ProjectEnablePage{
		ProjectsToEnable: projectsToEnable,
	}, user)
}

// Активация проекта на основе репозитория
func projectsEnable(w http.ResponseWriter, req *http.Request, user models.User) {
	repoName := req.URL.Query().Get("repoName")
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := providers.GetProviderByType(providerData.ProviderType)

	if provider == nil {
		http.NotFound(w, req)
		return
	}

	provider.SetProviderData(providerData)
	project := provider.LoadProjectByName(repoName)

	project.Status = models.STATUS_ENABLED
	project.Save()

	http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
}

// Редактирование плана сборки
func projectsPlan(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	message := ""

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if http.MethodPost == req.Method {
		plan := req.FormValue("plan")
		project.Plan = &plan

		if project.Save() {
			http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
		} else {
			message = "Can't save build plan. Please try again"
		}
	}

	templates.Render(w, "templates/projects/plan.html", ProjectPlanPage{
		Project: project,
		Message: message,
	}, user)
}
