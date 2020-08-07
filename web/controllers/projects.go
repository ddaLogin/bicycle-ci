package controllers

import (
	"fmt"
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

// Страница деталей проекта
type ProjectDetailPage struct {
	Project     models.Project
	BuildPlans  []*models.ProjectBuildPlan
	DeployPlans []*models.ProjectDeployPlan
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

// Страница редактирования плана сборки
type ProjectBuildPlanPage struct {
	Project   models.Project
	BuildPlan *models.ProjectBuildPlan
	Images    []models.DockerImage
	Message   string
}

// Страница редактирования плана релиза
type ProjectDeployPlanPage struct {
	Project    models.Project
	DeployPlan *models.ProjectDeployPlan
	Servers    []models.RemoteServer
	Message    string
}

// Страница проектов пользователя
func (c *ProjectController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "web/templates/projects/list.html", ProjectListPage{
		Projects: models.GetProjectsByUserId(user.Id),
	}, user)
}

// Страница проекта
func (c *ProjectController) Detail(w http.ResponseWriter, req *http.Request, user models.User) {
	project := models.GetProjectById(req.URL.Query().Get("id"))

	if (models.Project{}) == project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	templates.Render(w, "web/templates/projects/detail.html", ProjectDetailPage{
		Project:     project,
		BuildPlans:  models.GetProjectBuildPlansByProjectId(project.Id),
		DeployPlans: models.GetProjectDeployPlansByProjectId(project.Id),
	}, user)
}

// Страница выбора репозитория для нового проекта
func (c *ProjectController) Repos(w http.ResponseWriter, req *http.Request, user models.User) {
	providerData := models.GetProviderDataById(req.URL.Query().Get("providerId"))

	if (models.VcsProviderData{}) == providerData && providerData.UserId != user.Id {
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

	if (models.VcsProviderData{}) == providerData || providerData.UserId != user.Id {
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

	if (models.Project{}) == project || project.UserId != user.Id {
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

		if provider == nil || providerData == (models.VcsProviderData{}) {
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

// Редактирование/Создание плана сборки
func (c *ProjectController) PlanBuild(w http.ResponseWriter, req *http.Request, user models.User) {
	project := models.GetProjectById(req.URL.Query().Get("projectId"))
	images := models.GetImages()
	buildPlan := &models.ProjectBuildPlan{}
	message := ""

	if (models.Project{}) == project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	buildPlanId := req.URL.Query().Get("id")

	if buildPlanId != "" {
		buildPlan = models.GetProjectBuildPlanById(buildPlanId)

		if buildPlan == nil || *buildPlan == (models.ProjectBuildPlan{}) {
			http.NotFound(w, req)
			return
		}
	}

	if http.MethodPost == req.Method {
		imageId, _ := strconv.Atoi(req.FormValue("docker_image"))

		buildPlan.Title = req.FormValue("title")
		buildPlan.ProjectId = int(project.Id)
		buildPlan.BuildInstruction = req.FormValue("plan")
		buildPlan.Artifact = req.FormValue("artifact")
		buildPlan.DockerImage = imageId

		if buildPlan.Save() {
			http.Redirect(w, req, fmt.Sprintf("/projects/detail?id=%d", project.Id), http.StatusSeeOther)
		} else {
			message = "Не удалось сохранить план сборки. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/projects/plan-build.html", ProjectBuildPlanPage{
		Project:   project,
		BuildPlan: buildPlan,
		Images:    images,
		Message:   message,
	}, user)
}

// Редактирование/Создание деплоймент плана
func (c *ProjectController) PlanDeploy(w http.ResponseWriter, req *http.Request, user models.User) {
	project := models.GetProjectById(req.URL.Query().Get("projectId"))
	servers := models.GetAllServers()
	deployPlan := &models.ProjectDeployPlan{}
	message := ""

	if (models.Project{}) == project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	deployPlanId := req.URL.Query().Get("id")

	if deployPlanId != "" {
		deployPlan = models.GetProjectDeployPlanById(deployPlanId)

		if deployPlan == nil || *deployPlan == (models.ProjectDeployPlan{}) {
			http.NotFound(w, req)
			return
		}
	}

	if http.MethodPost == req.Method {
		serverId, _ := strconv.Atoi(req.FormValue("remote_server"))

		deployPlan.Title = req.FormValue("title")
		deployPlan.ProjectId = int(project.Id)
		deployPlan.DeploymentDirectory = req.FormValue("deployment_directory")

		if serverId > 0 {
			deployPlan.RemoteServerId = &serverId
		} else {
			deployPlan.RemoteServerId = nil
		}

		if deployPlan.Save() {
			http.Redirect(w, req, fmt.Sprintf("/projects/detail?id=%d", project.Id), http.StatusSeeOther)
		} else {
			message = "Не удалось сохранить план релиза. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/projects/plan-deploy.html", ProjectDeployPlanPage{
		Project:    project,
		DeployPlan: deployPlan,
		Servers:    servers,
		Message:    message,
	}, user)
}
