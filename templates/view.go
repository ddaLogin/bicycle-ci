package templates

import (
	"bicycle-ci/models"
	"html/template"
	"net/http"
)

// Базовый темплейт
type baseTemplate struct {
	User    models.User
	Content interface{}
}

// Выполнение указанного шаблона
func Render(w http.ResponseWriter, templateFile string, data interface{}, user models.User) {
	base := baseTemplate{
		User:    user,
		Content: data,
	}

	view, _ := template.New("").ParseFiles(templateFile, "templates/base.html")

	view.ExecuteTemplate(w, "base", base)
}
