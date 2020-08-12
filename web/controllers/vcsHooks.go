package controllers

import (
	"encoding/json"
	"fmt"
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
	Ref    string
	Sender struct {
		Id int
	}
	Commits []struct {
		Message string
	}
}

// Страница создания/редактирования хука
type HookCreatePage struct {
	Project    *models.Project
	BuildPlans []*models.ProjectBuildPlan
	VcsHook    *models.VcsHook
	Message    string
}

// Страница создания/редактирования хука
func (c *VcsHooksController) Create(w http.ResponseWriter, req *http.Request, user *models.User) {
	project := models.GetProjectById(req.URL.Query().Get("projectId"))
	message := ""
	vcsHook := &models.VcsHook{}

	if project == nil || (models.Project{}) == *project || project.UserId != user.Id {
		http.NotFound(w, req)
		return
	}

	vcsHookId := req.URL.Query().Get("id")

	if vcsHookId != "" {
		vcsHook = models.GetVcsHookById(vcsHookId)

		if vcsHook == nil || *vcsHook == (models.VcsHook{}) {
			http.NotFound(w, req)
			return
		}
	}

	if http.MethodPost == req.Method {
		providerData := models.GetProviderDataById(project.Provider)
		provider := vcs.GetProviderByType(providerData.ProviderType)
		buildPlan := models.GetProjectBuildPlanById(req.FormValue("build_plan_id"))

		if provider == nil || providerData == (models.VcsProviderData{}) {
			http.NotFound(w, req)
			return
		}

		if buildPlan == nil || (models.ProjectBuildPlan{}) == *buildPlan {
			http.NotFound(w, req)
			return
		}

		provider.SetProviderData(providerData)

		vcsHook.Branch = req.FormValue("branch")
		vcsHook.Event = `push`
		vcsHook.UserId = user.Id
		vcsHook.ProjectId = project.Id
		vcsHook.ProjectBuildPlanId = buildPlan.Id
		vcsHook.HookId = "0"

		// Сохраняем хук первым, для получения его ID, т.к это id нужно указать в callback url хука
		if vcsHook.Save() {
			vcsHook.HookId = provider.CreateWebHook(vcsHook, project)

			if vcsHook.HookId == "0" || vcsHook.HookId == "" {
				message = "Не удалось создать триггер. Пожалуйста попробуйте позже."
				vcsHook.Delete()
			} else {
				vcsHook.Save()

				http.Redirect(w, req, fmt.Sprintf("/hooks/list?projectId=%d", project.Id), http.StatusSeeOther)
			}
		} else {
			message = "Не удалось сохранить триггер. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/hooks/create.html", HookCreatePage{
		Project:    project,
		VcsHook:    vcsHook,
		BuildPlans: models.GetProjectBuildPlansByProjectId(project.Id),
		Message:    message,
	}, user)
}

// Роут тригера хука
func (c *VcsHooksController) Trigger(w http.ResponseWriter, req *http.Request) {
	if http.MethodPost == req.Method {
		hookId := req.URL.Query().Get("hookId")
		vcsHook := models.GetVcsHookById(hookId)

		if vcsHook == nil || *vcsHook == (models.VcsHook{}) {
			http.NotFound(w, req)
			return
		}

		var payload HookPayload
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&payload)
		if err != nil {
			log.Print("Не удалось прочитать vcs web vcsHook. ", err, payload)
			http.NotFound(w, req)
			return
		}

		if strings.Contains(payload.Ref, "/"+vcsHook.Branch) {
			project := models.GetProjectById(strconv.Itoa(int(vcsHook.ProjectId)))

			if project == nil || (models.Project{}) == *project {
				http.NotFound(w, req)
				return
			}

			var commits []string

			if payload.Ref != "" {
				for _, commit := range payload.Commits {
					commits = append(commits, commit.Message)
				}
			}

			user := models.GetUserByVcsId(payload.Sender.Id)

			if user == nil || *user == (models.User{}) {
				user = models.GetUserByLogin("Vcs-trigger")
			}

			c.workerService.RunBuild(vcsHook.GetProjectBuildPlan(), user, vcsHook.Branch, commits)

			w.WriteHeader(200)
			return
		}
	}

	http.NotFound(w, req)
	return
}
