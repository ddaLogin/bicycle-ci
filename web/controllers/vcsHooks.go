package controllers

import (
	"encoding/json"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/vcs"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"github.com/ddalogin/bicycle-ci/worker"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Контроллер vcs web хуков
type VcsHooksController struct {
	auth          *auth.Service
	workerService *worker.Service
}

// Конструктор контроллера vcs web хуков
func NewHookController(auth *auth.Service, workerService *worker.Service) *VcsHooksController {
	return &VcsHooksController{auth: auth, workerService: workerService}
}

// Мдель хука получаемого от гитхаба
type HookPayload struct {
	Ref     string
	Commits []struct {
		Message string
	}
}

// Страница списка хуков проекта
type HookListPage struct {
	Project models.Project
	Hooks   []models.VcsHook
}

// Страница создания/редактирования хука
type HookCreatePage struct {
	Project models.Project
	Message string
}

// Страница хуков проекта
func (c *VcsHooksController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	templates.Render(w, "web/templates/hooks/list.html", HookListPage{
		Project: project,
		Hooks:   models.GetHooksByProjectId(projectId),
	}, user)
}

// Страница создания/редактирования хука
func (c *VcsHooksController) Create(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)
	message := ""

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	if http.MethodPost == req.Method {
		branch := req.FormValue("branch")
		providerData := models.GetProviderDataById(strconv.Itoa(int(project.Provider)))
		provider := vcs.GetProviderByType(providerData.ProviderType)

		if provider == nil || providerData == (models.VcsProviderData{}) {
			http.NotFound(w, req)
			return
		}

		provider.SetProviderData(providerData)

		webHook := models.VcsHook{}
		webHook.Branch = branch
		webHook.Event = `push`
		webHook.UserId = user.Id
		webHook.ProjectId = project.Id

		if webHook.Save() {
			hookId := provider.CreateWebHook(webHook, project)
			webHook.HookId = &hookId

			if *webHook.HookId == "0" || *webHook.HookId == "" || webHook.Save() == false {
				// TODO: REMOVE HOOK ON PROVIDER
				webHook.Delete()
				message = "Не удалось создать триггер. Пожалуйста попробуйте позже."
			} else {
				http.Redirect(w, req, "/hooks/list?projectId="+strconv.Itoa(int(project.Id)), http.StatusSeeOther)
			}
		} else {
			message = "Не удалось сохранить триггер. Пожаолуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/hooks/create.html", HookCreatePage{
		Project: project,
		Message: message,
	}, user)
}

// Роут тригера хука
func (c *VcsHooksController) Trigger(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost == req.Method {
		hookId := req.URL.Query().Get("hookId")
		hook := models.GetHookById(hookId)

		if (models.VcsHook{}) == hook {
			http.NotFound(w, req)
			return
		}

		var payload HookPayload
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&payload)
		if err != nil {
			log.Print("Не удалось прочитать vcs web hook. ", err, payload)
			http.NotFound(w, req)
			return
		}

		if strings.Contains(payload.Ref, "/"+hook.Branch) {
			project := models.GetProjectById(strconv.Itoa(int(hook.ProjectId)))

			if (models.Project{}) == project {
				http.NotFound(w, req)
				return
			}

			var commits []string

			if payload.Ref != "" {
				for _, commit := range payload.Commits {
					commits = append(commits, commit.Message)
				}
			}

			c.workerService.RunBuild(project, commits)

			w.WriteHeader(200)
			return
		}
	}

	http.NotFound(w, req)
	return
}
