package models

import (
	"bicycle-ci/database"
	"log"
)

const STATUS_RUNNING = 0 // Сборка в процессе
const STATUS_SUCCESS = 1 // Сборка прошла успешно
const STATUS_FAILED = 2  // Сборка завершилась с ошибкой

// Модель процесса сборки проекта
type Build struct {
	Id        int64
	ProjectId int64
	Status    int
	StartedAt string
	EndedAt   *string
}

// Хелпер для рендера названия статуса
func (b Build) StatusTitle() string {
	switch b.Status {
	case STATUS_RUNNING:
		return "In progress"
	case STATUS_SUCCESS:
		return "Success build"
	case STATUS_FAILED:
		return "Build failed"
	}

	return ""
}

// Хелпер для рендера статуса нужным цветом
func (b Build) StatusColor() string {
	switch b.Status {
	case STATUS_RUNNING:
		return "info"
	case STATUS_SUCCESS:
		return "success"
	case STATUS_FAILED:
		return "danger"
	}

	return ""
}

// Сохранить билд
func (bld *Build) Save() bool {
	db := database.Db()
	defer db.Close()

	if bld.Id == 0 {
		result, err := db.Exec("insert into builds (project_id, status, started_at, ended_at) values (?, ?, ?, ?)",
			bld.ProjectId, bld.Status, bld.StartedAt, bld.EndedAt)
		if err != nil {
			log.Fatal("Can't insert Build. ", err, bld)
			return false
		}

		bld.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE builds SET project_id = ?, status = ?, started_at = ?, ended_at = ? WHERE id = ?",
			bld.ProjectId, bld.Status, bld.StartedAt, bld.EndedAt, bld.Id)
		if err != nil {
			log.Println("Can't update Build. ", err, bld)
			return false
		}

		return true
	}

	return false
}

// Получить билд по идентификатору
func GetBuildById(id string) (build Build) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM builds WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get Build by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&build.Id,
			&build.ProjectId,
			&build.Status,
			&build.StartedAt,
			&build.EndedAt,
		)
		if err != nil {
			log.Println("Can't scan Build by id. ", err)
			continue
		}
	}

	return
}

// Получить список всех билдов
func GetBuilds() (builds []Build) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM builds ORDER BY started_at DESC")
	if err != nil {
		log.Println("Can't get Build like list. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		build := Build{}
		err := rows.Scan(
			&build.Id,
			&build.ProjectId,
			&build.Status,
			&build.StartedAt,
			&build.EndedAt,
		)
		if err != nil {
			log.Println("Can't scan Build like list. ", err)
			continue
		}

		builds = append(builds, build)
	}

	return
}
