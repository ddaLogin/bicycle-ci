package main

import (
	"github.com/BurntSushi/toml"
	"github.com/ddalogin/bicycle-ci/actions"
	"github.com/ddalogin/bicycle-ci/database"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/providers/github"
	"github.com/ddalogin/bicycle-ci/telegram"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Url      string
	Db       database.Config
	Github   github.Config
	Telegram telegram.Config
}

func init() {
	initLogger()
}

func main() {
	cfg := loadConfig()
	actions.Host = cfg.Url
	models.Host = cfg.Url

	database.SetConfig(cfg.Db)
	github.SetConfig(cfg.Github)
	telegram.SetConfig(cfg.Telegram)

	startServer(cfg)
}

// Инициализация логов
func initLogger() {
	f, err := os.OpenFile("errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
}

// Чтение настроек
func loadConfig() Config {
	var config Config
	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		log.Fatal(err)
	}

	return config
}

// Подготовка, настройка и запуск сервера
func startServer(cfg Config) {
	actions.IndexRoutes()
	actions.ProjectRoutes()
	actions.HookRoutes()
	actions.ProviderRoutes()
	actions.BuildsRoutes()
	actions.ServerRoutes()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8090", nil)
}
