package actions

import (
	"encoding/json"
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/providers"
	"github.com/ddalogin/bicycle-ci/templates"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Мдель хука получаемого от гитхаба
type HookPayload struct {
	Ref string
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

// Регистрация роутов по хукам
func HookRoutes() {
	http.Handle("/hooks/list", auth.RequireAuthentication(hookList))
	http.Handle("/hooks/create", auth.RequireAuthentication(hookCreate))
	http.Handle("/hooks/trigger", http.HandlerFunc(hookTrigger))
}

// Страница хуков проекта
func hookList(w http.ResponseWriter, req *http.Request, user models.User) {
	projectId := req.URL.Query().Get("projectId")
	project := models.GetProjectById(projectId)

	if (models.Project{}) == project && project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	templates.Render(w, "templates/hooks/list.html", HookListPage{
		Project: project,
		Hooks:   models.GetHooksByProjectId(projectId),
	}, user)
}

// Страница создания/редактирования хука
func hookCreate(w http.ResponseWriter, req *http.Request, user models.User) {
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
		provider := providers.GetProviderByType(providerData.ProviderType)

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
				message = "Can't create or save web hook. Please try again"
			} else {
				http.Redirect(w, req, "/hooks/list?projectId="+strconv.Itoa(int(project.Id)), http.StatusSeeOther)
			}
		} else {
			message = "Can't save web hook. Please try again"
		}
	}

	templates.Render(w, "templates/hooks/create.html", HookCreatePage{
		Project: project,
		Message: message,
	}, user)
}

// Роут тригера хука
func hookTrigger(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost == req.Method {
		hookId := req.URL.Query().Get("hookId")
		hook := models.GetHookById(hookId)

		if (models.WebHook{}) == hook {
			http.NotFound(w, req)
			return
		}

		decoder := json.NewDecoder(req.Body)
		var payload HookPayload
		err := decoder.Decode(&payload)
		if err != nil {
			http.NotFound(w, req)
			return
		}

		if strings.Contains(payload.Ref, "/"+hook.Branch) {
			project := models.GetProjectById(strconv.Itoa(int(hook.ProjectId)))

			if (models.Project{}) == project {
				http.NotFound(w, req)
				return
			}

			build := models.Build{
				ProjectId: project.Id,
				StartedAt: time.Now().Format("2006-01-02 15:04:05"),
				Status:    models.STATUS_RUNNING,
			}
			build.Save()

			go Process(project, build)

			w.WriteHeader(200)
		}
	}

	http.NotFound(w, req)
	return
}
