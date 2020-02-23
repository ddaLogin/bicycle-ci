package actions

import (
	"fmt"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/providers"
	"github.com/ddalogin/bicycle-ci/templates"
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

	providerData := models.GetProviderDataByUserAndType(user.Id, providerType)

	if (models.ProviderData{} == providerData) {
		providerData = models.ProviderData{UserId: user.Id, ProviderType: providerType}
	}

	providerData.ProviderAuthToken = providerToken
	provider.UpdateProviderData(&providerData)

	providerData.Save()

	http.Redirect(w, req, "/projects/choose?providerId="+fmt.Sprintf("%v", providerData.Id), http.StatusSeeOther)
}
