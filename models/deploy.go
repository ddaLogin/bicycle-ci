package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
	"text/template"
	"time"
)

const DeployStatusRunning = 0 // Сборка в процессе
const DeployStatusSuccess = 1 // Сборка прошла успешно
const DeployStatusFailed = 2  // Сборка завершилась с ошибкой

// Модель релиза проекта
type Deploy struct {
	Id                  int64
	ProjectDeployPlanId int64
	UserId              int64
	Status              int
	StdOut              string
	StdErr              string
	Error               string
	StartedAt           string
	EndedAt             *string
	user                *User
	projectDeployPlan   *ProjectDeployPlan
}

// Модель сообщения о начале релиза
type DeployStartMessage struct {
	Deploy    *Deploy
	DeployUrl string
}

// Модель сообщения о завершение релиза
type DeployCompleteMessage struct {
	Deploy    *Deploy
	DeployUrl string
}

// Создает модель реультата сборки по строке из базы
func scanDeploy(row *sql.Row) (deploy Deploy) {
	err := row.Scan(
		&deploy.Id,
		&deploy.ProjectDeployPlanId,
		&deploy.UserId,
		&deploy.Status,
		&deploy.StdOut,
		&deploy.StdErr,
		&deploy.Error,
		&deploy.StartedAt,
		&deploy.EndedAt,
	)
	if err != nil {
		log.Println("Не удалось собрать модель реультата релиза", row)
	}

	return
}

// Создает массив моделей реультата сборки по строкам из базы
func scanDeploys(rows *sql.Rows) (deploys []*Deploy) {
	for rows.Next() {
		deploy := Deploy{}
		err := rows.Scan(
			&deploy.Id,
			&deploy.ProjectDeployPlanId,
			&deploy.UserId,
			&deploy.Status,
			&deploy.StdOut,
			&deploy.StdErr,
			&deploy.Error,
			&deploy.StartedAt,
			&deploy.EndedAt,
		)
		if err != nil {
			log.Println("Не удалось собрать модель реультата релиза из массива строк", err)
			return
		}

		deploys = append(deploys, &deploy)
	}

	return
}

// Получить план релиза
func (dpl *Deploy) GetProjectDeployPlan() *ProjectDeployPlan {
	if dpl.projectDeployPlan == nil {
		dpl.projectDeployPlan = GetProjectDeployPlanById(dpl.ProjectDeployPlanId)
	}

	return dpl.projectDeployPlan
}

// Получить пользователя запустившего релиз
func (dpl *Deploy) GetUser() *User {
	if dpl.user == nil {
		dpl.user = GetUserById(dpl.UserId)
	}

	return dpl.user
}

// Хелпер для рендера названия статуса
func (dpl *Deploy) GetStatusTitle() string {
	switch dpl.Status {
	case BuildStatusRunning:
		return "В процессе"
	case BuildStatusSuccess:
		return "Успешно"
	case BuildStatusFailed:
		return "Ошибка"
	}

	return ""
}

// Хелпер для рендера статуса нужным цветом
func (dpl *Deploy) GetStatusColor() string {
	switch dpl.Status {
	case BuildStatusRunning:
		return "info"
	case BuildStatusSuccess:
		return "success"
	case BuildStatusFailed:
		return "danger"
	}

	return ""
}

// Собирает сообщение о начале сборки
func (dpl *Deploy) GetStartMessage(host string) string {
	var buffer bytes.Buffer
	messageTemplate, err := template.ParseFiles("templates/deploy_start")

	if err != nil {
		log.Fatal("Не удалось прочитать шаблон сообщения о начале релиза", err)

		return ""
	}

	err = messageTemplate.Execute(&buffer, DeployStartMessage{
		Deploy:    dpl,
		DeployUrl: fmt.Sprintf("%s/deployments/status?id=%d", host, dpl.Id),
	})

	if err != nil {
		log.Fatal("Не удалось собрать сообщние о начале релиза по шаблону", err)

		return ""
	}

	return buffer.String()
}

// Собирает сообщение о завершение релиза
func (dpl *Deploy) GetCompleteMessage(host string) string {
	var buffer bytes.Buffer
	messageTemplate, err := template.ParseFiles("templates/deploy_complete")

	if err != nil {
		log.Fatal("Не удалось прочитать шаблон сообщения о завершение релиза", err)

		return ""
	}

	err = messageTemplate.Execute(&buffer, DeployStartMessage{
		Deploy:    dpl,
		DeployUrl: fmt.Sprintf("%s/deployments/status?id=%d", host, dpl.Id),
	})

	if err != nil {
		log.Fatal("Не удалось собрать сообщние о завершение релиза по шаблону", err)

		return ""
	}

	return buffer.String()
}

// Получить продолжительность релиза
func (dpl *Deploy) GetProcessTime() string {
	if dpl.EndedAt == nil {
		return ""
	}

	st, err := time.Parse("2006-01-02 15:04:05", dpl.StartedAt)
	if err != nil {
		log.Fatal("Не удалось распарсить время начала релиза")
	}

	en, err := time.Parse("2006-01-02 15:04:05", *dpl.EndedAt)
	if err != nil {
		log.Fatal("Не удалось распарсить время завершения релиза")
	}

	return en.Sub(st).String()
}

// Сохранить реультат сборки
func (dpl *Deploy) Save() bool {
	db := database.Db()
	defer db.Close()

	if dpl.Id == 0 {
		result, err := db.Exec(
			"insert into deployments (project_deploy_plan_id, user_id, status, std_out, std_err, error, started_at, ended_at) values (?, ?, ?, ?, ?, ?, ?, ?)",
			dpl.ProjectDeployPlanId, dpl.UserId, dpl.Status, dpl.StdOut, dpl.StdErr, dpl.Error, dpl.StartedAt, dpl.EndedAt,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый реультат релиза", err, dpl)
			return false
		}

		dpl.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового реультата релиза", err, dpl)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE deployments SET project_deploy_plan_id = ?, user_id = ?, status = ?, std_out = ?, std_err = ?, error = ?, started_at = ?, ended_at = ? WHERE id = ?",
			dpl.ProjectDeployPlanId, dpl.UserId, dpl.Status, dpl.StdOut, dpl.StdErr, dpl.Error, dpl.StartedAt, dpl.EndedAt, dpl.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить реультат релиза", err, dpl)
			return false
		}
	}

	return true
}

// Получить результат релиза по идентификатору
func GetDeployById(id interface{}) *Deploy {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM deployments WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти результат релиза по ID", id)
		return nil
	}

	deploy := scanDeploy(row)

	return &deploy
}

// Получить список результатов всех релизов одного проекта
func GetAllDeploysByProjectId(projectId interface{}) []*Deploy {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT dpl.* FROM deployments as dpl JOIN project_deploy_plans pdp on dpl.project_deploy_plan_id = pdp.id WHERE pdp.project_id = ? ORDER BY dpl.started_at DESC", projectId)
	if err != nil {
		log.Println("Не удалось найти результаты всех релизов проекта")
		return nil
	}
	defer rows.Close()

	return scanDeploys(rows)
}
