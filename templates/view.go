package templates

import (
	"bicycle-ci/auth"
	"html/template"
	"net/http"
)

type BaseData struct {
	User auth.User
}

// Выполнение указанного шаблона
func Render(w http.ResponseWriter, templateFile string, data interface{}) {
	view, _ := template.New("").ParseFiles(templateFile, "templates/base.html")

	view.ExecuteTemplate(w, "base", data)
}
