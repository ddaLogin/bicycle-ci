package actions

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/providers"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/templates"
	"net/http"
	"strconv"
)

// Страница списка проектов
type ProjectListPage struct {
	Projects []models.Project
}

// Страница активации проектов
type ProjectEnablePage struct {
	ProjectsToEnable map[int]*models.Project
}

// Страница настройки ключей деплоя
type ProjectDeployPage struct {
	Project models.Project
	Message string
}

// Страница редактирования плана
type ProjectPlanPage struct {
	Project models.Project
	Message string
}

// Регистрация роутов по проектам
func ProjectRoutes() {
	http.Handle("/projects/list", auth.RequireAuthentication(projectsList))
	http.Handle("/projects/choose", auth.RequireAuthentication(projectsChoose))
	http.Handle("/projects/enable", auth.RequireAuthentication(projectsEnable))
	http.Handle("/projects/deploy", auth.RequireAuthentication(projectsDeploy))
	http.Handle("/projects/plan", auth.RequireAuthentication(projectsPlan))
}

// Страница проектов пользователя
func projectsList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница выбора репозитория для нового проекта
func projectsChoose(w http.ResponseWriter, req *http.Request, user models.User) {
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := providers.GetProviderByType(providerData.ProviderType)

	if provider == nil {
		http.NotFound(w, req)
		return
	}

	provider.SetProviderData(providerData)
	projectsToEnable := provider.LoadProjects()

	for _, value := range models.GetProjectsByUserId(user.Id) {
		if _, ok := projectsToEnable[value.RepoId]; ok {
			projectsToEnable[value.RepoId] = &value
		}
	}

	templates.Render(w, "templates/projects/choose.html", ProjectEnablePage{
		ProjectsToEnable: projectsToEnable,
	}, user)
}

// Активация проекта на основе репозитория
func projectsEnable(w http.ResponseWriter, req *http.Request, user models.User) {
	repoName := req.URL.Query().Get("repoName")
	repoOwner := req.URL.Query().Get("repoOwner")
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := providers.GetProviderByType(providerData.ProviderType)

	if provider == nil {
		http.NotFound(w, req)
		return
	}

	provider.SetProviderData(providerData)
	project := provider.LoadProjectToEnable(repoOwner, repoName)

	project.Save()

	http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
}

// Настройка ключей деплоя
func projectsDeploy(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	message := ""

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if http.MethodPost == req.Method {
		isGenerate := req.FormValue("isNeedGenerated")
		publicKey := req.FormValue("public_key")
		privateKey := req.FormValue("private_key")
		titleKey := req.FormValue("title_key")
		providerData := models.GetProviderDataById(strconv.Itoa(int(project.Provider)))
		provider := providers.GetProviderByType(providerData.ProviderType)

		if provider == nil || providerData == (models.ProviderData{}) {
			http.NotFound(w, req)
			return
		}

		provider.SetProviderData(providerData)

		// Автоматически генерируем SSH ключи
		if "true" == isGenerate {
			pair := ssh.GenerateKeyPair()
			publicKey = string(pair.Public)
			privateKey = string(pair.Private)
		}

		keyId := provider.UploadProjectDeployKey(titleKey, publicKey, project)

		if 0 != keyId {
			project.DeployKeyId = &keyId
			project.DeployPrivate = &privateKey

			if project.Save() {
				http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
			} else {
				message = "Can't save project with deploy key. Please try again"
			}
		} else {
			message = "Can't upload deployment key. Please try again"
		}
	}

	templates.Render(w, "templates/projects/deploy.html", ProjectDeployPage{
		Project: project,
		Message: message,
	}, user)
}

// Редактирование плана сборки
func projectsPlan(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	message := ""

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if http.MethodPost == req.Method {
		plan := req.FormValue("plan")
		project.Plan = &plan

		if project.Save() {
			http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
		} else {
			message = "Can't save build plan. Please try again"
		}
	}

	templates.Render(w, "templates/projects/plan.html", ProjectPlanPage{
		Project: project,
		Message: message,
	}, user)
}
