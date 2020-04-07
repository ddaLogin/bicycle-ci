package actions

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/templates"
	"net/http"
	"strconv"
)

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

// Регистрация роутов по серверам
func ServerRoutes() {
	http.Handle("/servers/list", auth.RequireAuthentication(serverList))
	http.Handle("/servers/create", auth.RequireAuthentication(serverCreate))
}

// Страница серверов
func serverList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/servers/list.html", ServersListPage{
		Servers: models.GetAllServers(),
		Message: "",
	}, user)
}

// Страница создание сервера
func serverCreate(w http.ResponseWriter, req *http.Request, user models.User) {
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
			pair := ssh.GenerateKeyPair()
			server.DeployPublic = string(pair.Public)
			server.DeployPrivate = string(pair.Private)
		}

		if server.Save() {
			http.Redirect(w, req, "/servers/list", http.StatusSeeOther)
		} else {
			message = "Can't save server. Please try again"
		}
	}

	templates.Render(w, "templates/servers/create.html", ServerCreatePage{
		Server:  server,
		Message: message,
	}, user)
}
