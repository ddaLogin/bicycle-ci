package models

import (
	"bicycle-ci/database"
	"log"
)

// Статусы проектов
const STATUS_NOT_ENABLED = 0 // Проект не активирован
const STATUS_ENABLED = 1     // Проект активирован

// Модель проекта
type Project struct {
	Id            int64   // Идентификатор проекта
	UserId        int     // Владелец проекта
	Name          string  // Имя проекта
	Provider      int64   // Идентификатор провайдера репозитория
	RepoId        int     // Идентификатор репозитория
	RepoName      *string // Имя репозитория
	RepoOwnerName *string // Логин владельца репозитория
	RepoOwnerId   *string // Идентификатор владельца репозитория
	Status        int     // Статус проекта
}

// Сохранить проект
func (pr Project) Save() bool {
	db := database.Db()
	defer db.Close()

	if pr.Id == 0 {
		result, err := db.Exec("insert into projects (user_id, `name`, provider, repo_id, repo_name, repo_owner_name, repo_owner_id, status) values (?, ?, ?, ?, ?, ?, ?, ?)",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.Status)
		if err != nil {
			log.Println("Can't insert Project. ", err, pr)
			return false
		}

		pr.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE projects SET user_id = ?, `name` = ?, provider = ?, repo_id = ?, repo_name = ?, repo_owner_name = ?, repo_owner_id = ?, status = ? WHERE id = ?",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.Status, pr.Id)
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
			&project.Status,
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
			&project.Status,
		)
		if err != nil {
			log.Println("Can't scan project by id. ", err)
			continue
		}
	}

	return
}
