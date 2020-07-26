package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
	"strconv"
)

// Контроллер удаленных серверов
type ServerController struct {
	auth *auth.Service
	ssh  *ssh.Service
}

// Конструктор контроллера удаленных сверверов
func NewServerController(auth *auth.Service, ssh *ssh.Service) *ServerController {
	return &ServerController{auth: auth, ssh: ssh}
}

// Страница списка серверов
type ServersListPage struct {
	Servers []models.Server
	Message string
}

// Страница создание/редактирования сервера
type ServerCreatePage struct {
	Server  models.Server
	Message string
}

// Страница серверов
func (c *ServerController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "web/templates/servers/list.html", ServersListPage{
		Servers: models.GetAllServers(),
		Message: "",
	}, user)
}

// Страница создание сервера
func (c *ServerController) Create(w http.ResponseWriter, req *http.Request, user models.User) {
	serverId := req.URL.Query().Get("serverId")
	buf, _ := strconv.Atoi(serverId)
	server := models.GetServerById(buf)
	message := ""

	if http.MethodPost == req.Method {
		name := req.FormValue("name")
		login := req.FormValue("login")
		host := req.FormValue("host")
		isGenerate := req.FormValue("generate")

		if (models.Server{}) == server {
			isGenerate = "true"
		}

		server.Name = name
		server.Login = login
		server.Host = host

		// Автоматически генерируем SSH ключи
		if "true" == isGenerate {
			pair := c.ssh.GenerateKeyPair()
			server.DeployPublic = string(pair.Public)
			server.DeployPrivate = string(pair.Private)
		}

		if server.Save() {
			http.Redirect(w, req, "/servers/list", http.StatusSeeOther)
		} else {
			message = "Не удалось сохранить сервер. Пожалуйста попробуйте позже."
		}
	}

	templates.Render(w, "web/templates/servers/create.html", ServerCreatePage{
		Server:  server,
		Message: message,
	}, user)
}
