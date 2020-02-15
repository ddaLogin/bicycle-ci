package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/templates"

	"net/http"
)

type IndexPage struct {
	Base    templates.BaseData
	Message string
}

type LoginPage struct {
	Base    templates.BaseData
	Message string
}

func IndexRoutes() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
}

// Главная страница
func index(w http.ResponseWriter, req *http.Request) {
	user := auth.GetCurrentUser(req)
	if (auth.User{}) == user {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	templates.Render(w, "templates/index.html", IndexPage{
		Base: templates.BaseData{User: user},
	})
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

	templates.Render(w, "templates/login.html", data)
}
