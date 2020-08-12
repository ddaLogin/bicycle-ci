package web

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/web/controllers"
	"github.com/ddalogin/bicycle-ci/worker"
	"log"
	"net/http"
)

// Конфиг http сервера
type Config struct {
	Host string
	Port string
}

// Http сервер
type Server struct {
	config        Config
	authService   *auth.Service
	sshService    *ssh.Service
	workerService *worker.Service
}

// Конструктор http сервера
func NewServer(config Config, authService *auth.Service, sshService *ssh.Service, workerService *worker.Service) *Server {
	return &Server{config: config, authService: authService, sshService: sshService, workerService: workerService}
}

// Запуск сервера
func (s *Server) Run() {
	s.route()

	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	err := http.ListenAndServe(s.config.Host+":"+s.config.Port, nil)
	if err != nil {
		log.Fatal("Не удалось поднять http сервер", err)
	}
}

// Настройка всех доступных роутов
func (s *Server) route() {
	indexC := controllers.NewIndexController(s.authService)
	http.Handle("/", s.authService.AuthMiddleware(indexC.Index))
	http.HandleFunc("/login", indexC.Login)

	imagesC := controllers.NewImagesController(s.authService)
	http.Handle("/docker/images/list", s.authService.AuthMiddleware(imagesC.List))
	http.Handle("/docker/images/create", s.authService.AuthMiddleware(imagesC.Create))

	projectC := controllers.NewProjectController(s.authService, s.sshService)
	http.Handle("/projects/list", s.authService.AuthMiddleware(projectC.List))
	http.Handle("/projects/detail", s.authService.AuthMiddleware(projectC.Detail))
	http.Handle("/projects/repos", s.authService.AuthMiddleware(projectC.Repos))
	http.Handle("/projects/create", s.authService.AuthMiddleware(projectC.Create))
	http.Handle("/projects/deploy", s.authService.AuthMiddleware(projectC.Deploy))
	http.Handle("/projects/plan/build", s.authService.AuthMiddleware(projectC.PlanBuild))
	http.Handle("/projects/plan/deploy", s.authService.AuthMiddleware(projectC.PlanDeploy))

	serverC := controllers.NewServerController(s.authService, s.sshService)
	http.Handle("/servers/list", s.authService.AuthMiddleware(serverC.List))
	http.Handle("/servers/create", s.authService.AuthMiddleware(serverC.Create))

	vscC := controllers.NewVcsController(s.authService)
	http.Handle("/vcs/list", s.authService.AuthMiddleware(vscC.List))
	http.Handle("/vcs/callback", s.authService.AuthMiddleware(vscC.OAuthCallback))

	hookC := controllers.NewHookController(s.authService, s.workerService)
	http.Handle("/hooks/create", s.authService.AuthMiddleware(hookC.Create))
	http.Handle("/hooks/trigger", http.HandlerFunc(hookC.Trigger))

	buildC := controllers.NewBuildsController(s.authService, s.workerService)
	http.Handle("/builds/run", s.authService.AuthMiddleware(buildC.Run))
	http.Handle("/builds/status", s.authService.AuthMiddleware(buildC.Status))
	http.Handle("/builds/artifact", s.authService.AuthMiddleware(buildC.Artifact))

	deployC := controllers.NewDeploysController(s.authService, s.workerService)
	http.Handle("/deployments/run", s.authService.AuthMiddleware(deployC.Release))
	http.Handle("/deployments/status", s.authService.AuthMiddleware(deployC.Status))

	userC := controllers.NewUsersController(s.authService)
	http.Handle("/users/registration", s.authService.AuthMiddleware(userC.Registration))
	http.Handle("/users/list", s.authService.AuthMiddleware(userC.List))
}
