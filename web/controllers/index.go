package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
)

// Основной контроллер
type IndexController struct {
	auth *auth.Service
}

// Конструктор контроллера
func NewIndexController(auth *auth.Service) *IndexController {
	return &IndexController{auth: auth}
}

// Главная страница
type IndexPage struct {
	Builds []*models.Build
}

// Страница авторизации
type LoginPage struct {
	Message string
}

// Главная страница
func (c *IndexController) Index(w http.ResponseWriter, req *http.Request, user *models.User) {
	templates.Render(w, "web/templates/index.html", IndexPage{
		Builds: models.GetBuilds(50),
	}, user)
}

// Страница авторизации
func (c *IndexController) Login(w http.ResponseWriter, req *http.Request) {
	page := LoginPage{}

	if req.Method == http.MethodPost {
		login := req.FormValue("login")
		password := req.FormValue("password")

		result := c.auth.Auth(login, password, w, req)

		if result {
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}

		page.Message = "Неверный логин или пароль"
	}

	templates.Render(w, "web/templates/login.html", page, &models.User{})
}
