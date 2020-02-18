package providers

import (
	"bicycle-ci/models"
	"bicycle-ci/providers/github"
	"net/http"
)

// GitHub провайдер
const GITHUB_TYPE = 1

// Интерфейс VCS провадеров
type ProviderInterface interface {
	SetProviderData(providerData models.ProviderData)       // Устанавливает данные провайдера
	GetTitle() string                                       // Название провайдера
	GetImageUrl() string                                    // Ссылка на картинку
	GetAuthLink() string                                    // Генерация ссылки для OAuth авторизации
	OAuthCallback(req *http.Request) string                 // Обработка oAuth авторизации
	UpdateProviderData(provider *models.ProviderData)       // Запрос на основную информацию аккаунта
	LoadProjects() (projects map[int]*models.Project)       // Загрузить список репозиториев
	LoadProjectByName(name string) (project models.Project) // Загрузить репозиторий
}

// Список всех доступных VCS провайдеров
func GetAvailableProviders() (list []ProviderInterface) {
	list = append(list, &github.GitHub{})
	return
}

// Фабричный метод для создания нужного провайдера по типу
func GetProviderByType(providerType int) (provider ProviderInterface) {

	if providerType == GITHUB_TYPE {
		provider = &github.GitHub{}
	}

	return
}
