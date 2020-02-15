package actions

import (
	"bicycle-ci/providers/github"
	"net/http"
)

func ProjectRoutes() {
	http.HandleFunc("/projects/list", list)
}

// Страница провайдеров
func list(w http.ResponseWriter, req *http.Request) {
	view, _ := template.ParseFiles("templates/providers/list.html")

	view.Execute(w, listPage{
		GithubAuthLink: github.GetOAuthLink(),
		Token:          token.Token,
	})
}
