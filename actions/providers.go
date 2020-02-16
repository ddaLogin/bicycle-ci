package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"bicycle-ci/providers"
	"bicycle-ci/templates"
	"net/http"
	"strconv"
)

// Страница списка доступных провайдеров
type ProvidersListPage struct {
	Providers []providers.ProviderInterface
	Message   string
}

//type ReposPage struct {
//	Repos []github.Repo
//}

// Регистрация роутов по провайдерам
func ProviderRoutes() {
	http.Handle("/providers/list", auth.RequireAuthentication(providersList))
	http.Handle("/providers/callback", auth.RequireAuthentication(oAuthCallback))

	//http.Handle("/providers/github/repos", auth.RequireAuthentication(gitHubRepos))
}

// Страница провайдеров
func providersList(w http.ResponseWriter, req *http.Request, user models.User) {
	message := req.URL.Query().Get("message")
	templates.Render(w, "templates/providers/list.html", ProvidersListPage{
		Providers: providers.GetAvailableProviders(),
		Message:   message,
	}, user)
}

// Callback роут после oauth авторизации у провайдера
func oAuthCallback(w http.ResponseWriter, req *http.Request, user models.User) {
	providerType, _ := strconv.Atoi(req.URL.Query().Get("providerType"))
	provider := providers.GetProviderByType(providerType)

	if provider == nil {
		http.Redirect(w, req, "/providers/list?message=Unknown provider", http.StatusSeeOther)
		return
	}

	providerToken := provider.OAuthCallback(req)

	if "" == providerToken {
		http.Redirect(w, req, "/providers/list?message=Failed to login, try again", http.StatusSeeOther)
		return
	}

	providerModel := models.ProviderData{UserId: user.Id, ProviderType: providerType, ProviderAuthToken: providerToken}
	provider.GetProviderData(&providerModel)

	providerModel.Save()

	http.Redirect(w, req, "/providers/repos?providerType="+req.URL.Query().Get("providerType"), http.StatusSeeOther)
}

//// Страница репозиториев с гитхаба
//func gitHubRepos(w http.ResponseWriter, req *http.Request, user models.User) {
//	templates.Render(w, "templates/providers/repos.html", ReposPage{
//		Repos: github.GetRepos(),
//	}, user)
//}
