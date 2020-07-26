package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/vcs"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
	"strconv"
)

// Контроллер проектов
type ProjectController struct {
	auth *auth.Service
	ssh  *ssh.Service
}

//Констрктор контроллера проектов
func NewProjectController(auth *auth.Service, ssh *ssh.Service) *ProjectController {
	return &ProjectController{auth: auth, ssh: ssh}
}

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
	Servers []models.Server
	Images  []models.Image
	Message string
}

// Страница проектов пользователя
func (c *ProjectController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "web/templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница выбора репозитория для нового проекта
func (c *ProjectController) Repos(w http.ResponseWriter, req *http.Request, user models.User) {
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := vcs.GetProviderByType(providerData.ProviderType)

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

	templates.Render(w, "web/templates/projects/repos.html", ProjectEnablePage{
		ProjectsToEnable: projectsToEnable,
	}, user)
}

// Активация проекта на основе репозитория
func (c *ProjectController) Create(w http.ResponseWriter, req *http.Request, user models.User) {
	repoName := req.URL.Query().Get("repoName")
	repoOwner := req.URL.Query().Get("repoOwner")
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.ProviderData{}) == providerData && providerData.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	provider := vcs.GetProviderByType(providerData.ProviderType)

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
func (c *ProjectController) Deploy(w http.ResponseWriter, req *http.Request, user models.User) {
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
		provider := vcs.GetProviderByType(providerData.ProviderType)

		if provider == nil || providerData == (models.ProviderData{}) {
			http.NotFound(w, req)
			return
		}

		provider.SetProviderData(providerData)

		// Автоматически генерируем SSH ключи
		if "true" == isGenerate {
			pair := c.ssh.GenerateKeyPair()
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
				message = "Не удалось сохранить деплой ключи. Пожалуйста попробуй позже."
			}
		} else {
			message = "Не удалось загрузить деплой ключи. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/projects/deploy.html", ProjectDeployPage{
		Project: project,
		Message: message,
	}, user)
}

// Редактирование плана сборки
func (c *ProjectController) Plan(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	servers := models.GetAllServers()
	images := models.GetImages()
	message := ""

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if http.MethodPost == req.Method {
		plan := req.FormValue("plan")
		deployDir := req.FormValue("deploy_dir")
		artifactDir := req.FormValue("artifact_dir")
		imageId := req.FormValue("build_image")
		serverId := req.FormValue("server_id")

		buff, _ := strconv.Atoi(imageId)
		project.BuildImage = &buff
		project.BuildPlan = &plan
		project.DeployDir = &deployDir
		project.ArtifactDir = &artifactDir

		buff2, _ := strconv.Atoi(serverId)
		project.ServerId = &buff2

		if project.Save() {
			http.Redirect(w, req, "/projects/list", http.StatusSeeOther)
		} else {
			message = "Не удалось сохранить план сборки. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/projects/plan.html", ProjectPlanPage{
		Project: project,
		Servers: servers,
		Images:  images,
		Message: message,
	}, user)
}
