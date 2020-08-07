package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Статусы проектов
const StatusNotEnabled = 0  // Проект не активирован
const StatusNotClonable = 1 // Нехватает ключей скачивания репозитория
const StatusReady = 2       // Готов к сборке

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
}

// Создает модель проекта по строке из базы
func scanProject(row *sql.Row) (project Project) {
	err := row.Scan(
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
	)
	if err != nil {
		log.Println("Не удалось собрать модель проекта", row)
	}

	return
}

// Создает массив моделей проекта по строкам из базы
func scanProjects(rows *sql.Rows) (projects []*Project) {
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
		)
		if err != nil {
			log.Println("Не удалось собрать модель проекта из массива строк", err)
			return
		}

		projects = append(projects, &project)
	}

	return
}

// Получить модель Docker образа
func (pr *Project) Image() *DockerImage {
	return GetDockerImageById(1)
}

// Получить статус проекта
func (pr Project) Status() int {
	if pr.Id == 0 {
		return StatusNotEnabled
	}

	if pr.DeployKeyId == nil || *pr.DeployKeyId == 0 || pr.DeployPrivate == nil || *pr.DeployPrivate == "" {
		return StatusNotClonable
	}

	return StatusReady
}

// Хелпер для рендера названия статуса
func (pr Project) StatusTitle() string {
	switch pr.Status() {
	case StatusNotEnabled:
		return "Не активен"
	case StatusNotClonable:
		return "Установите VCS ключи"
	case StatusReady:
		return "Готов к сборке"
	}

	return ""
}

// Хелпер для рендера статуса нужным цветом
func (pr Project) StatusColor() string {
	switch pr.Status() {
	case StatusNotEnabled:
		return "secondary"
	case StatusNotClonable:
		return "warning"
	case StatusReady:
		return "primary"
	}

	return ""
}

// Сохранить проект
func (pr *Project) Save() bool {
	db := database.Db()
	defer db.Close()

	if pr.Id == 0 {
		result, err := db.Exec(
			"insert into projects (user_id, `name`, provider, repo_id, repo_name, repo_owner_name, repo_owner_id, deploy_key_id, deploy_private) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый проекта", err, pr)
			return false
		}

		pr.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового проекта", err, pr)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE projects SET user_id = ?, `name` = ?, provider = ?, repo_id = ?, repo_name = ?, repo_owner_name = ?, repo_owner_id = ?, deploy_key_id = ?, deploy_private = ? WHERE id = ?",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate, pr.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить проект", err, pr)
			return false
		}
	}

	return true
}

// Получить проекты пользователя
func GetProjectsByUserId(userId interface{}) []*Project {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM projects WHERE user_id = ?", userId)
	if err != nil {
		log.Println("Не удалось найти все проекты пользователя")
		return nil
	}
	defer rows.Close()

	return scanProjects(rows)
}

// Получить проект по идентификатору
func GetProjectById(id interface{}) *Project {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM projects WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти план сборки по ID", id)
		return nil
	}

	project := scanProject(row)

	return &project
}

// Получить проект по идентификатору
func (pr Project) GetAvgBuildTime() string {
	var time string
	db := database.Db()
	defer db.Close()

	err := db.QueryRow("select COALESCE(RIGHT(SEC_TO_TIME(ROUND(AVG(TIMESTAMPDIFF(SECOND , started_at, ended_at)))), 5), '') from builds WHERE project_id = ?", pr.Id).Scan(&time)
	switch {
	case err == sql.ErrNoRows:
		log.Println("Не удалось получить среднее время сборки проекта", err)
		time = ""
	case err != nil:
		log.Println("Ошибка при получение среднего времени сборки", err)
		time = ""
	}

	return time
}
