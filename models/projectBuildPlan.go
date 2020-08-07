package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель плана сборки проекта
type ProjectBuildPlan struct {
	Id               int64  // Идентификатор плана сборки
	ProjectId        int    // Идентификатор собираемого проекта
	Title            string // Заголовок сборки
	DockerImage      int    // Docker образ в котором будет выполняться сборка
	BuildInstruction string // Инструкции сборки проекта
	Artifact         string // Цель будущего деплоймента (файл или папка после сборки)
}

// Создает модель плана сборки по строке из базы
func scanBuildPlan(row *sql.Row) (plan ProjectBuildPlan) {
	err := row.Scan(
		&plan.Id,
		&plan.ProjectId,
		&plan.Title,
		&plan.DockerImage,
		&plan.BuildInstruction,
		&plan.Artifact,
	)
	if err != nil {
		log.Println("Не удалось собрать модель плана сборки", row)
	}

	return
}

// Создает массив моделей плана сборки по строкам из базы
func scanBuildPlans(rows *sql.Rows) (plans []*ProjectBuildPlan) {
	for rows.Next() {
		plan := ProjectBuildPlan{}
		err := rows.Scan(
			&plan.Id,
			&plan.ProjectId,
			&plan.Title,
			&plan.DockerImage,
			&plan.BuildInstruction,
			&plan.Artifact,
		)
		if err != nil {
			log.Println("Не удалось собрать модель плана сборки из массива строк", err)
			return
		}

		plans = append(plans, &plan)
	}

	return
}

// Сохранить план сборки
func (pl *ProjectBuildPlan) Save() bool {
	db := database.Db()
	defer db.Close()

	if pl.Id == 0 {
		result, err := db.Exec(
			"insert into project_build_plans (project_id, title, docker_image_id, build_instruction, artifact) values (?, ?, ?, ?, ?)",
			pl.ProjectId, pl.Title, pl.DockerImage, pl.BuildInstruction, pl.Artifact,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый план сборки", err, pl)
			return false
		}

		pl.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового плана сборки", err, pl)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE project_build_plans SET project_id = ?, title = ?, docker_image_id = ?, build_instruction = ?, artifact = ? WHERE id = ?",
			pl.ProjectId, pl.Title, pl.DockerImage, pl.BuildInstruction, pl.Artifact, pl.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить план сборки", err, pl)
			return false
		}
	}

	return true
}

// Получить план сборки по id
func GetProjectBuildPlanById(id interface{}) *ProjectBuildPlan {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM project_build_plans WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти план сборки по ID", id)
		return nil
	}

	plan := scanBuildPlan(row)

	return &plan
}

// Получить планы сборки проекта
func GetProjectBuildPlansByProjectId(projectId interface{}) []*ProjectBuildPlan {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM project_build_plans WHERE project_id = ?", projectId)
	if err != nil {
		log.Println("Не удалось найти все планы сборок проекта")
		return nil
	}
	defer rows.Close()

	return scanBuildPlans(rows)
}
