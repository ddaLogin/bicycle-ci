package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
)

// Контроллер пользователей
type UsersController struct {
	auth *auth.Service
}

// Конструктор контроллера пользователей
func NewUsersController(auth *auth.Service) *UsersController {
	return &UsersController{auth: auth}
}

// Шаблон страницы регистрации
type RegistrationPage struct {
	Message string
}

// Шаблон страницы иска пользователей
type UsersListPage struct {
	Users []*models.User
}

// Список пользователей
func (c *UsersController) List(w http.ResponseWriter, req *http.Request, user *models.User) {
	templates.Render(w, "web/templates/users/list.html", UsersListPage{
		Users: models.GetAllUsers(),
	}, user)
}

// Регистрация
func (c *UsersController) Registration(w http.ResponseWriter, req *http.Request, user *models.User) {
	var message string

	if req.Method == http.MethodPost {
		login := req.FormValue("login")
		password := req.FormValue("password")
		user := models.GetUserByLogin(login)

		if user != nil && *user != (models.User{}) {
			message = "Указанный логин уже зарегестрирован"
		} else {
			user = &models.User{
				Login:    login,
				Password: c.auth.HashPassword(password),
				Salt:     c.auth.GenerateSalt(10),
			}

			if user.Save() {
				http.Redirect(w, req, "/users/list", http.StatusSeeOther)
				return
			} else {
				message = "Не удалось сохранить пользователя"
			}
		}
	}

	templates.Render(w, "web/templates/users/registration.html", RegistrationPage{
		Message: message,
	}, user)
}
