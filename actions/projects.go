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
	ProjectsToEnable map[int]models.Project
}

// Регистрация роутов по проектам
func ProjectRoutes() {
	http.Handle("/projects/list", auth.RequireAuthentication(projectsList))
	http.Handle("/projects/choose", auth.RequireAuthentication(projectsChoose))
}

// Страница проектов пользователя
func projectsList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница подключения проекта
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
