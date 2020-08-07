package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель RemoteServer'а
type RemoteServer struct {
	Id            int64  // Идентификатор
	Name          string // Имя сервера
	Login         string // Логин пользователя
	Host          string // Хост сервера
	DeployPublic  string // Публичный ключ деплоя
	DeployPrivate string // Приватный ключ деплоя
}

// Получить все сервера
func GetAllServers() (servers []RemoteServer) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM remote_servers")
	if err != nil {
		log.Println("Can't get all servers. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		server := RemoteServer{}
		err := rows.Scan(
			&server.Id,
			&server.Name,
			&server.Login,
			&server.Host,
			&server.DeployPublic,
			&server.DeployPrivate,
		)
		if err != nil {
			log.Println("Can't scan servers. ", err)
			continue
		}

		servers = append(servers, server)
	}

	return
}

// Получить сервер по идентификатору
func GetServerById(id int) (server RemoteServer) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM remote_servers WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get server by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&server.Id,
			&server.Name,
			&server.Login,
			&server.Host,
			&server.DeployPublic,
			&server.DeployPrivate,
		)
		if err != nil {
			log.Println("Can't scan server by id. ", err)
			continue
		}
	}

	return
}

// Сохранить сервер
func (sr RemoteServer) Save() bool {
	db := database.Db()
	defer db.Close()

	if sr.Id == 0 {
		result, err := db.Exec("insert into remote_servers (`name`, login, host, deploy_public, deploy_private) values (?, ?, ?, ?, ?)",
			sr.Name, sr.Login, sr.Host, sr.DeployPublic, sr.DeployPrivate)
		if err != nil {
			log.Println("Can't insert RemoteServer. ", err, sr)
			return false
		}

		sr.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE remote_servers SET `name` = ?, login = ?, host = ?, deploy_public = ?, deploy_private = ? WHERE id = ?",
			sr.Name, sr.Login, sr.Host, sr.DeployPrivate, sr.DeployPublic, sr.Id)
		if err != nil {
			log.Println("Can't update RemoteServer. ", err, sr)
			return false
		}

		return true
	}

	return false
}
