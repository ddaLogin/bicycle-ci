package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
	"os"
	"text/template"
	"time"
)

const StatusRunning = 0 // Сборка в процессе
const StatusSuccess = 1 // Сборка прошла успешно
const StatusFailed = 2  // Сборка завершилась с ошибкой

// Модель сборки проекта
type Build struct {
	Id                 int64
	ProjectBuildPlanId int64
	UserId             int64
	Status             int
	StartedAt          string
	EndedAt            *string
	projectBuildPlan   *ProjectBuildPlan
	user               *User
}

// Модель сообщения о начале сборки
type BuildStartMessage struct {
	Build    *Build
	BuildUrl string
	Commits  []string
}

// Модель сообщения о завершение сборки
type BuildCompleteMessage struct {
	Build    *Build
	Steps    []*BuildStep
	BuildUrl string
}

// Создает модель реультата сборки по строке из базы
func scanBuild(row *sql.Row) (build Build) {
	err := row.Scan(
		&build.Id,
		&build.ProjectBuildPlanId,
		&build.UserId,
		&build.Status,
		&build.StartedAt,
		&build.EndedAt,
	)
	if err != nil {
		log.Println("Не удалось собрать модель реультата сборки", row)
	}

	return
}

// Создает массив моделей реультата сборки по строкам из базы
func scanBuilds(rows *sql.Rows) (builds []*Build) {
	for rows.Next() {
		build := Build{}
		err := rows.Scan(
			&build.Id,
			&build.ProjectBuildPlanId,
			&build.UserId,
			&build.Status,
			&build.StartedAt,
			&build.EndedAt,
		)
		if err != nil {
			log.Println("Не удалось собрать модель реультата сборки из массива строк", err)
			return
		}

		builds = append(builds, &build)
	}

	return
}

// Возвращает имя и путь для артефактов сборки
func (bld *Build) GetArtifactName() string {
	return fmt.Sprintf("artifact_%d_%d_%d.zip", bld.GetProjectBuildPlan().ProjectId, bld.ProjectBuildPlanId, bld.Id)
}

// Проверяет существует ли файл артефактов сборки
func (bld *Build) IsArtifactExists() bool {
	info, err := os.Stat("builds/" + bld.GetArtifactName())
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Получить план сборки
func (bld *Build) GetProcessTime() string {
	if bld.EndedAt == nil {
		return ""
	}

	st, err := time.Parse("2006-01-02 15:04:05", bld.StartedAt)
	if err != nil {
		log.Fatal("Не удалось распарсить время начала сборки")
	}

	en, err := time.Parse("2006-01-02 15:04:05", *bld.EndedAt)
	if err != nil {
		log.Fatal("Не удалось распарсить время завершения сборки")
	}

	return en.Sub(st).String()
}

// Получить план сборки
func (bld *Build) GetProjectBuildPlan() *ProjectBuildPlan {
	if bld.projectBuildPlan == nil {
		bld.projectBuildPlan = GetProjectBuildPlanById(bld.ProjectBuildPlanId)
	}

	return bld.projectBuildPlan
}

// Получить пользователя запустившего сборку
func (bld *Build) GetUser() *User {
	if bld.user == nil {
		bld.user = GetUserById(bld.UserId)
	}

	return bld.user
}

// Хелпер для рендера названия статуса
func (bld *Build) GetStatusTitle() string {
	switch bld.Status {
	case StatusRunning:
		return "В процессе"
	case StatusSuccess:
		return "Успешно"
	case StatusFailed:
		return "Ошибка"
	}

	return ""
}

// Хелпер для рендера статуса нужным цветом
func (bld *Build) GetStatusColor() string {
	switch bld.Status {
	case StatusRunning:
		return "info"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "danger"
	}

	return ""
}

// Собирает сообщение о начале сборки
func (bld *Build) GetStartMessage(host string, commits []string) string {
	var buffer bytes.Buffer
	messageTemplate, err := template.ParseFiles("templates/build_start")

	if err != nil {
		log.Fatal("Не удалось прочитать шаблон сообщения о начале сборки", err)

		return ""
	}

	err = messageTemplate.Execute(&buffer, BuildStartMessage{
		Build:    bld,
		BuildUrl: fmt.Sprintf("%s/builds/status?buildId=%d", host, bld.Id),
		Commits:  commits,
	})

	if err != nil {
		log.Fatal("Не удалось собрать сообщние о начале сборки по шаблону", err)

		return ""
	}

	return buffer.String()
}

// Собирает сообщение о завершение сборки
func (bld *Build) GetCompleteMessage(host string) string {
	var buffer bytes.Buffer
	messageTemplate, err := template.ParseFiles("templates/build_complete")

	if err != nil {
		log.Fatal("Не удалось прочитать шаблон сообщения о завершение сборки", err)

		return ""
	}

	err = messageTemplate.Execute(&buffer, BuildCompleteMessage{
		Build:    bld,
		Steps:    GetStepsByBuildId(bld.Id),
		BuildUrl: fmt.Sprintf("%s/builds/status?buildId=%d", host, bld.Id),
	})

	if err != nil {
		log.Fatal("Не удалось собрать сообщние о завершение сборки по шаблону", err)

		return ""
	}

	return buffer.String()
}

// Сохранить реультат сборки
func (bld *Build) Save() bool {
	db := database.Db()
	defer db.Close()

	if bld.Id == 0 {
		result, err := db.Exec(
			"insert into builds (project_build_plan_id, user_id, status, started_at, ended_at) values (?, ?, ?, ?, ?)",
			bld.ProjectBuildPlanId, bld.UserId, bld.Status, bld.StartedAt, bld.EndedAt,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый реультат сборки", err, bld)
			return false
		}

		bld.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового реультата сборки", err, bld)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE builds SET project_build_plan_id = ?, user_id = ?, status = ?, started_at = ?, ended_at = ? WHERE id = ?",
			bld.ProjectBuildPlanId, bld.UserId, bld.Status, bld.StartedAt, bld.EndedAt, bld.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить реультат сборки", err, bld)
			return false
		}
	}

	return true
}

// Получить результат сборки по идентификатору
func GetBuildById(id interface{}) *Build {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM builds WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти результат сборки по ID", id)
		return nil
	}

	build := scanBuild(row)

	return &build
}

// Получить список результатов всех сборок
func GetAllBuilds() []*Build {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM builds ORDER BY started_at DESC")
	if err != nil {
		log.Println("Не удалось найти результаты всех сборок")
		return nil
	}
	defer rows.Close()

	return scanBuilds(rows)
}

// Получить список результатов всех сборок одного проекта
func GetAllBuildsByProjectId(projectId interface{}) []*Build {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT bld.* FROM builds as bld JOIN project_build_plans pbp on bld.project_build_plan_id = pbp.id WHERE pbp.project_id = ? ORDER BY bld.started_at DESC", projectId)
	if err != nil {
		log.Println("Не удалось найти результаты всех сборок проекта")
		return nil
	}
	defer rows.Close()

	return scanBuilds(rows)
}
