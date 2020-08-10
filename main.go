package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/ddalogin/bicycle-ci/auth"
	database2 "github.com/ddalogin/bicycle-ci/database"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/ssh"
	"github.com/ddalogin/bicycle-ci/telegram"
	"github.com/ddalogin/bicycle-ci/vcs/github"
	"github.com/ddalogin/bicycle-ci/web"
	"github.com/ddalogin/bicycle-ci/worker"
	"io"
	"log"
	"os"
)

type Config struct {
	SessionName         string
	SessionSecretKey    string
	MaxArtifactsPerPlan int
	Web                 web.Config
	Db                  database2.Config
	Github              github.Config
	Telegram            telegram.Config
}

func main() {
	configPath := flag.String("config", "config.toml", "Конфиг с настройками")
	logPath := flag.String("log", "errors.log", "Лог файл")
	flag.Parse()

	f, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Не удалось открыть лог файл", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	var config Config
	if _, err := toml.DecodeFile(*configPath, &config); err != nil {
		log.Fatal("Не удалось прочитать конфиг файл", err)
	}

	github.SetConfig(config.Github)
	database2.SetConfig(config.Db)
	models.Host = "http://e916cdad6447.ngrok.io"
	authService := auth.NewService(config.SessionName, config.SessionSecretKey, "/login")
	sshService := ssh.NewService()
	telegramService := telegram.NewService(config.Telegram)
	workerService := worker.NewService(telegramService, config.Web.Host, config.Web.Port, config.MaxArtifactsPerPlan)

	server := web.NewServer(config.Web, authService, sshService, workerService)
	server.Run()
}
