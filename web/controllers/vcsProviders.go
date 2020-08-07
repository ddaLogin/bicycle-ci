package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/vcs"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
	"strconv"
)

// Контроллер систем конроля версий (Github/Bitbucket ...)
type VcsProvidersController struct {
	auth *auth.Service
}

// Конструктор систем контроля версий
func NewVcsController(auth *auth.Service) *VcsProvidersController {
	return &VcsProvidersController{auth: auth}
}

// Страница списка доступных провайдеров
type ProvidersListPage struct {
	Providers []vcs.ProviderInterface
	Message   string
}

// Страница VCS провайдеров
func (c *VcsProvidersController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	message := req.URL.Query().Get("message")

	templates.Render(w, "web/templates/vcs/list.html", ProvidersListPage{
		Providers: vcs.GetAvailableProviders(),
		Message:   message,
	}, user)
}

// Callback роут после oauth авторизации у провайдера
func (c *VcsProvidersController) OAuthCallback(w http.ResponseWriter, req *http.Request, user models.User) {
	providerType, _ := strconv.Atoi(req.URL.Query().Get("providerType"))
	provider := vcs.GetProviderByType(providerType)

	if provider == nil {
		http.Redirect(w, req, "/vcs/list?message=Неизвестный провайдер", http.StatusSeeOther)
		return
	}

	providerToken := provider.OAuthCallback(req)

	if "" == providerToken {
		http.Redirect(w, req, "/vcs/list?message=Не удалось авторизоваться, попробуйте еще раз.", http.StatusSeeOther)
		return
	}

	providerData := models.GetProviderDataByUserAndType(user.Id, providerType)

	if (models.VcsProviderData{} == providerData) {
		providerData = models.VcsProviderData{UserId: user.Id, ProviderType: providerType}
	}

	providerData.ProviderAuthToken = providerToken
	provider.UpdateProviderData(&providerData)

	providerData.Save()

	http.Redirect(w, req, "/projects/repos?providerId="+strconv.Itoa(int(providerData.Id)), http.StatusSeeOther)
}
