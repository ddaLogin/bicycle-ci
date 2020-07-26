package controllers

import (
	"encoding/json"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/vcs"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Контроллер vcs web хуков
type HookController struct {
	auth *auth.Service
}

// Конструктор контроллера vcs web хуков
func NewHookController(auth *auth.Service) *HookController {
	return &HookController{auth: auth}
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
	Hooks   []models.WebHook
}

// Страница создания/редактирования хука
type HookCreatePage struct {
	Project models.Project
	Message string
}

// Страница хуков проекта
func (c *HookController) List(w http.ResponseWriter, req *http.Request, user models.User) {
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
func (c *HookController) Create(w http.ResponseWriter, req *http.Request, user models.User) {
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

		if provider == nil || providerData == (models.ProviderData{}) {
			http.NotFound(w, req)
			return
		}

		provider.SetProviderData(providerData)

		webHook := models.WebHook{}
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
func (c *HookController) Trigger(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost == req.Method {
		hookId := req.URL.Query().Get("hookId")
		hook := models.GetHookById(hookId)

		if (models.WebHook{}) == hook {
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

			//RunProcess(project, payload)

			w.WriteHeader(200)
			return
		}
	}

	http.NotFound(w, req)
	return
}
