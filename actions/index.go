package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/templates"
	"net/http"
)

// Главная страница
type IndexPage struct {
	Builds []models.Build
}

// Страница авторизации
type LoginPage struct {
	Message string
}

// Регистрация основных роутов
func IndexRoutes() {
	http.Handle("/", auth.RequireAuthentication(index))
	http.HandleFunc("/login", login)
}

// Главная страница
func index(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/index.html", IndexPage{
		Builds: models.GetBuildsWithProjects(),
	}, user)
}

// Страница авторизации
func login(w http.ResponseWriter, req *http.Request) {
	data := LoginPage{}

	if req.Method == http.MethodPost {
		result, _ := auth.Auth(w, req)

		if result {
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}

		data.Message = "Wrong login or password"
	}

	templates.Render(w, "templates/login.html", data, models.User{})
}
