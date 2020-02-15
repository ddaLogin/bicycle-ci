package actions

import (
	"bicycle-ci/providers/github"
	"html/template"
	"net/http"
)

type listPage struct {
	GithubAuthLink string
	Token          string
}

var token github.AccessToken

func ProviderRoutes() {
	http.HandleFunc("/providers/list", list)
	http.HandleFunc("/providers/github/callback", gitHubCallback)
}

// Страница провайдеров
func list(w http.ResponseWriter, req *http.Request) {
	view, _ := template.ParseFiles("templates/providers/list.html")

	view.Execute(w, listPage{
		GithubAuthLink: github.GetOAuthLink(),
		Token:          token.Token,
	})
}

// Callback роут после авторизации на гитхабе
func gitHubCallback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")

	token = github.GetAccessToken(code)
}
