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
	Id            int     // Идентификатор проекта
	UserId        int     // Владелец проекта
	Name          string  // Имя проекта
	Provider      int     // Идентификатор провайдера репозитория
	RepoId        int     // Идентификатор репозитория
	RepoName      *string // Имя репозитория
	RepoOwnerName *string // Логин владельца репозитория
	RepoOwnerId   *string // Идентификатор владельца репозитория
	Status        int     // Статус проекта
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
		)
		if err != nil {
			log.Println("Can't scan projects by user id. ", err)
			continue
		}

		projects = append(projects, project)
	}

	return
}
