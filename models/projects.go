package models

import (
	"bicycle-ci/database"
	"log"
)

// Статусы проектов
const STATUS_NOT_ENABLED = 0    // Проект не активирован
const STATUS_NOT_DEPLOYABLE = 1 // Нехватает ключей деплоя
const STATUS_NOT_CONFIGURED = 2 // Нехватает конфигурации
const STATUS_READY = 3          // Готов к сборке
const STATUS_BUILD_PROCESS = 4  // Сборка в процессе
const STATUS_BUILD_SUCCESS = 5  // Проект успешно собран
const STATUS_BUILD_FAILED = 6   // Во время сборки произошла ошибка

// Модель проекта
type Project struct {
	Id            int64   // Идентификатор проекта
	UserId        int     // Владелец проекта
	Name          string  // Имя проекта
	Provider      int64   // Идентификатор провайдера репозитория
	RepoId        int     // Идентификатор репозитория
	RepoName      string  // Имя репозитория
	RepoOwnerName string  // Логин владельца репозитория
	RepoOwnerId   string  // Идентификатор владельца репозитория
	DeployKeyId   *int    // Идентификатор ключа деплоя
	DeployPrivate *string // Приватный SSH ключ
	Status        int     // Статус проекта
	Plan          *string // Build plan проекта
}

// Хелпер для рендера названия статуса
func (pr Project) StatusTitle() string {
	switch pr.Status {
	case STATUS_NOT_ENABLED:
		return "Not enabled"
	case STATUS_NOT_DEPLOYABLE:
		return "Not deployable"
	case STATUS_NOT_CONFIGURED:
		return "Not configured"
	case STATUS_READY:
		return "Ready"
	case STATUS_BUILD_PROCESS:
		return "Build in progress"
	case STATUS_BUILD_SUCCESS:
		return "Success"
	case STATUS_BUILD_FAILED:
		return "Build failed"
	}

	return ""
}

// Хелпер для рендера статуса нужным цветом
func (pr Project) StatusColor() string {
	switch pr.Status {
	case STATUS_NOT_ENABLED:
		return "secondary"
	case STATUS_NOT_DEPLOYABLE:
		return "warning"
	case STATUS_NOT_CONFIGURED:
		return "warning"
	case STATUS_READY:
		return "info"
	case STATUS_BUILD_PROCESS:
		return "primary"
	case STATUS_BUILD_SUCCESS:
		return "success"
	case STATUS_BUILD_FAILED:
		return "danger"
	}

	return ""
}

// Сохранить проект
func (pr Project) Save() bool {
	db := database.Db()
	defer db.Close()

	if pr.Id == 0 {
		result, err := db.Exec("insert into projects (user_id, `name`, provider, repo_id, repo_name, repo_owner_name, repo_owner_id, deploy_key_id, deploy_private, status, plan) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate, pr.Status, pr.Plan)
		if err != nil {
			log.Println("Can't insert Project. ", err, pr)
			return false
		}

		pr.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE projects SET user_id = ?, `name` = ?, provider = ?, repo_id = ?, repo_name = ?, repo_owner_name = ?, repo_owner_id = ?, deploy_key_id = ?, deploy_private = ?, status = ?, plan = ? WHERE id = ?",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate, pr.Status, pr.Plan, pr.Id)
		if err != nil {
			log.Println("Can't update Project. ", err, pr)
			return false
		}

		return true
	}

	return false
}

// Получить проекты пользователя
func GetProjectsByUserId(userId int) (projects []Project) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM projects WHERE user_id = ?", userId)
	if err != nil {
		log.Println("Can't get projects by user id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		project := Project{}
		err := rows.Scan(
			&project.Id,
			&project.UserId,
			&project.Name,
			&project.Provider,
			&project.RepoId,
			&project.RepoName,
			&project.RepoOwnerName,
			&project.RepoOwnerId,
			&project.DeployKeyId,
			&project.DeployPrivate,
			&project.Status,
			&project.Plan,
		)
		if err != nil {
			log.Println("Can't scan projects by user id. ", err)
			continue
		}

		projects = append(projects, project)
	}

	return
}

// Получить проект по идентификатору
func GetProjectById(id string) (project Project) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM projects WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get project by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&project.Id,
			&project.UserId,
			&project.Name,
			&project.Provider,
			&project.RepoId,
			&project.RepoName,
			&project.RepoOwnerName,
			&project.RepoOwnerId,
			&project.DeployKeyId,
			&project.DeployPrivate,
			&project.Status,
			&project.Plan,
		)
		if err != nil {
			log.Println("Can't scan project by id. ", err)
			continue
		}
	}

	return
}
