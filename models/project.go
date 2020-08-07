package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Статусы проектов
const StatusNotEnabled = 0    // Проект не активирован
const StatusNotClonable = 1   // Нехватает ключей скачивания репозитория
const StatusNotConfigured = 2 // Нехватает конфигурации
const StatusNotDeployable = 3 // Нехватает конфигурации деплоя
const StatusReady = 4         // Готов к сборке
const StatusBuildProcess = 5  // Сборка в процессе
const StatusBuildSuccess = 6  // Проект успешно собран
const StatusBuildFailed = 7   // Во время сборки произошла ошибка

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

//
//// Получить модель Docker образа
//func (pr *Project) Image() DockerImage {
//	return GetImageById(strconv.Itoa(*pr.BuildImage))
//}

// Получить статус проекта
//func (pr Project) Status() int {
//	if pr.Id == 0 {
//		return StatusNotEnabled
//	}
//
//	if pr.DeployKeyId == nil || *pr.DeployKeyId == 0 || pr.DeployPrivate == nil || *pr.DeployPrivate == "" {
//		return StatusNotClonable
//	}
//
//	if pr.BuildPlan == nil || *pr.BuildPlan == "" || pr.BuildImage == nil || *pr.BuildImage == 0 {
//		return StatusNotConfigured
//	}
//
//	if pr.DeployDir == nil || *pr.DeployDir == "" || pr.ArtifactDir == nil || *pr.ArtifactDir == "" {
//		return StatusNotDeployable
//	}
//
//	return StatusReady
//}

//// Хелпер для рендера названия статуса
//func (pr Project) StatusTitle() string {
//	switch pr.Status() {
//	case StatusNotEnabled:
//		return "Не активен"
//	case StatusNotClonable:
//		return "Установите VCS ключи"
//	case StatusNotConfigured:
//		return "Создайте план сборки"
//	case StatusNotDeployable:
//		return "Укажите директорию развертывания"
//	case StatusReady:
//		return "Готов к сборке"
//	case StatusBuildProcess:
//		return "Сборка в процессе"
//	case StatusBuildSuccess:
//		return "Успешная сборка"
//	case StatusBuildFailed:
//		return "Ошибка в сборке"
//	}
//
//	return ""
//}
//
//// Хелпер для рендера статуса нужным цветом
//func (pr Project) StatusColor() string {
//	switch pr.Status() {
//	case StatusNotEnabled:
//		return "secondary"
//	case StatusNotClonable:
//		return "warning"
//	case StatusNotConfigured:
//		return "warning"
//	case StatusNotDeployable:
//		return "warning"
//	case StatusReady:
//		return "primary"
//	case StatusBuildProcess:
//		return "info"
//	case StatusBuildSuccess:
//		return "success"
//	case StatusBuildFailed:
//		return "danger"
//	}
//
//	return ""
//}

// Сохранить проект
func (pr *Project) Save() bool {
	db := database.Db()
	defer db.Close()

	if pr.Id == 0 {
		result, err := db.Exec("insert into projects (user_id, `name`, provider, repo_id, repo_name, repo_owner_name, repo_owner_id, deploy_key_id, deploy_private) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate)
		if err != nil {
			log.Println("Can't insert Project. ", err, pr)
			return false
		}

		pr.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE projects SET user_id = ?, `name` = ?, provider = ?, repo_id = ?, repo_name = ?, repo_owner_name = ?, repo_owner_id = ?, deploy_key_id = ?, deploy_private = ? WHERE id = ?",
			pr.UserId, pr.Name, pr.Provider, pr.RepoId, pr.RepoName, pr.RepoOwnerName, pr.RepoOwnerId, pr.DeployKeyId, pr.DeployPrivate, pr.Id)
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
func GetProjectById(id interface{}) (project Project) {
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
		)
		if err != nil {
			log.Println("Can't scan project by id. ", err)
			continue
		}
	}

	return
}

// Получить проект по идентификатору
func (pr Project) GetAvgBuildTime() string {
	var time string
	db := database.Db()
	defer db.Close()

	err := db.QueryRow("select COALESCE(RIGHT(SEC_TO_TIME(ROUND(AVG(TIMESTAMPDIFF(SECOND , started_at, ended_at)))), 5), '') from builds WHERE project_id = ?", pr.Id).Scan(&time)
	switch {
	case err == sql.ErrNoRows:
		log.Println("Can't get avg time. ", err)
		time = ""
	case err != nil:
		log.Println("Error while get avg time. ", err)
		time = ""
	}

	return time
}
