package templates

import (
	"github.com/ddalogin/bicycle-ci/models"
	"html/template"
	"net/http"
)

// Базовый темплейт
type baseTemplate struct {
	User    *models.User
	Content interface{}
}

// Выполнение указанного шаблона
func Render(w http.ResponseWriter, templateFile string, data interface{}, user *models.User) {
	base := baseTemplate{
		User:    user,
		Content: data,
	}

	view, _ := template.New("").Funcs(template.FuncMap{
		"RefEq": func(i *int64, j int64) bool {
			if i == nil {
				return false
			}

			return *i == j
		},
	}).ParseFiles(templateFile, "web/templates/base.html")

	view.ExecuteTemplate(w, "base", base)
}
