package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"net/http"
)

// Страница списка проектов
type ProjectListPage struct {
	Projects []models.Project
}

// Регистрация роутов по проектам
func ProjectRoutes() {
	http.Handle("/projects/list", auth.RequireAuthentication(projectsList))
	http.Handle("/projects/enable", auth.RequireAuthentication(projectsEnable))
}

// Страница проектов пользователя
func projectsList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница подключения проекта
func projectsEnable(w http.ResponseWriter, req *http.Request, user models.User) {
	//owner := req.URL.Query().Get("owner")
	//repo := req.URL.Query().Get("repo")
	////hook := github.CreatePushHook(owner, repo)
	//
	//fmt.Printf("%+v", hook)
	//
	//http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
}
