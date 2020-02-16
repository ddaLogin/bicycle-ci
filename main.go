package main

import (
	"bicycle-ci/actions"
	"bicycle-ci/database"
	"bicycle-ci/providers/github"
	"github.com/BurntSushi/toml"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Url    string
	Db     database.Config
	Github github.Config
}

func init() {
	initLogger()
}

func main() {
	cfg := loadConfig()

	database.SetConfig(cfg.Db)
	github.SetConfig(cfg.Github)

	startServer(cfg)
}

// Инициализация логов
func initLogger() {
	f, err := os.OpenFile("./errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	actions.ProviderRoutes()
	actions.BuildsRoutes()

	http.ListenAndServe(":8090", nil)
}
