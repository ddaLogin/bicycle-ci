package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель релиз плана
type ProjectDeployPlan struct {
	Id                  int64  // Идентификатор плана релиза
	ProjectId           int    // Идентификатор релизного проекта
	Title               string // Заголовок релиз плана (test/production/staging)
	RemoteServerId      *int   // Сервер для удаленного деплоя, null = local
	DeploymentDirectory string // Папка на удаленном сервере, куда будет развернут проект после деплоймента
}

// Создает модель релиз плана по строке из базы
func scanDeployPlan(row *sql.Row) (plan ProjectDeployPlan) {
	err := row.Scan(
		&plan.Id,
		&plan.ProjectId,
		&plan.Title,
		&plan.RemoteServerId,
		&plan.DeploymentDirectory,
	)
	if err != nil {
		log.Println("Не удалось собрать модель плана релиза", row)
	}

	return
}

// Создает массив моделей релиз планов по строкам из базы
func scanDeployPlans(rows *sql.Rows) (plans []*ProjectDeployPlan) {
	for rows.Next() {
		plan := ProjectDeployPlan{}
		err := rows.Scan(
			&plan.Id,
			&plan.ProjectId,
			&plan.Title,
			&plan.RemoteServerId,
			&plan.DeploymentDirectory,
		)
		if err != nil {
			log.Println("Не удалось собрать модель плана релиза из массива строк", err)
			return
		}

		plans = append(plans, &plan)
	}

	return
}

// Сохранить план релиза
func (pl *ProjectDeployPlan) Save() bool {
	db := database.Db()
	defer db.Close()

	if pl.Id == 0 {
		result, err := db.Exec(
			"insert into project_deploy_plans (project_id, title, remote_server_id, deployment_directory) values (?, ?, ?, ?)",
			pl.ProjectId, pl.Title, pl.RemoteServerId, pl.DeploymentDirectory,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый план релиза", err, pl)
			return false
		}

		pl.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового плана релиза", err, pl)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE project_deploy_plans SET project_id = ?, title = ?, remote_server_id = ?, deployment_directory = ? WHERE id = ?",
			pl.ProjectId, pl.Title, pl.RemoteServerId, pl.DeploymentDirectory, pl.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить план релиза", err, pl)
			return false
		}
	}

	return true
}

// Получить план релиза по id
func GetProjectDeployPlanById(id interface{}) *ProjectDeployPlan {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM project_deploy_plans WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти план релиза по ID", id)
		return nil
	}

	plan := scanDeployPlan(row)

	return &plan
}

// Получить релиз планы проекта
func GetProjectDeployPlansByProjectId(projectId interface{}) []*ProjectDeployPlan {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM project_deploy_plans WHERE project_id = ?", projectId)
	if err != nil {
		log.Println("Не удалось найти все релиз планы проекта")
		return nil
	}
	defer rows.Close()

	return scanDeployPlans(rows)
}
