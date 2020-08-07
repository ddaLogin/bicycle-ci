package vcs

import (
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/vcs/github"
	"net/http"
)

// GitHub провайдер
const GithubType = 1

// Интерфейс VCS провадеров
type ProviderInterface interface {
	SetProviderData(providerData models.VcsProviderData)                            // Устанавливает данные провайдера
	GetTitle() string                                                               // Название провайдера
	GetImageUrl() string                                                            // Ссылка на картинку
	GetAuthLink() string                                                            // Генерация ссылки для OAuth авторизации
	OAuthCallback(req *http.Request) string                                         // Обработка oAuth авторизации
	UpdateProviderData(provider *models.VcsProviderData)                            // Запрос на основную информацию аккаунта
	LoadProjects() (projects map[int]*models.Project)                               // Загрузить список репозиториев
	LoadProjectToEnable(ownerName string, repoName string) (project models.Project) // Загрузить репозиторий
	UploadProjectDeployKey(keyName string, key string, project *models.Project) int // Загружает на сервер VCS деплой ключ
	CreateWebHook(webHook *models.VcsHook, project *models.Project) string          // Создает Web Hook в репозитории
}

// Список всех доступных VCS провайдеров
func GetAvailableProviders() (list []ProviderInterface) {
	list = append(list, &github.GitHub{})
	return
}

// Фабричный метод для создания нужного провайдера по типу
func GetProviderByType(providerType int) (provider ProviderInterface) {

	if providerType == GithubType {
		provider = &github.GitHub{}
	}

	return
}
