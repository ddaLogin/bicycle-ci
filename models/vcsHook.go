package models

import (
	"database/sql"
	"fmt"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

var Host string

// Модель VcsHook'а
type VcsHook struct {
	Id                 int64  // Идентификатор хука
	ProjectId          int64  // Проект за которым закреплен хук
	ProjectBuildPlanId int64  // Проект за которым закреплен хук
	UserId             int    // Создатель хука
	HookId             string // Идентификатор хука на стороне провайдера
	Event              string // Тип события при котором хук срабатывает
	Branch             string // Целевая ветка хука
}

// Создает модель VcsHook'а по строке из базы
func scanVcsHook(row *sql.Row) (hook VcsHook) {
	err := row.Scan(
		&hook.Id,
		&hook.ProjectId,
		&hook.ProjectBuildPlanId,
		&hook.UserId,
		&hook.HookId,
		&hook.Event,
		&hook.Branch,
	)
	if err != nil {
		log.Println("Не удалось собрать модель VcsHook'а", row)
	}

	return
}

// Создает массив моделей VcsHook'а по строкам из базы
func scanVcsHooks(rows *sql.Rows) (hooks []*VcsHook) {
	for rows.Next() {
		hook := VcsHook{}
		err := rows.Scan(
			&hook.Id,
			&hook.ProjectId,
			&hook.ProjectBuildPlanId,
			&hook.UserId,
			&hook.HookId,
			&hook.Event,
			&hook.Branch,
		)
		if err != nil {
			log.Println("Не удалось собрать модель VcsHook'а из массива строк", err)
			return
		}

		hooks = append(hooks, &hook)
	}

	return
}

// Хелпер для генерации урла по которому хук будет трегериться
func (vh *VcsHook) GetTriggerUrl() string {
	return fmt.Sprintf("%s/hooks/trigger?hookId=%d", Host, vh.Id)
}

// Возвращает план сборки за которой закреплен хук
func (vh *VcsHook) GetProjectBuildPlan() *ProjectBuildPlan {
	return GetProjectBuildPlanById(vh.ProjectBuildPlanId)
}

// Сохранить VcsHook
func (vh *VcsHook) Save() bool {
	db := database.Db()
	defer db.Close()

	if vh.Id == 0 {
		result, err := db.Exec(
			"INSERT INTO vcs_hooks (project_id, project_build_plan_id, user_id, hook_id, event, branch) VALUES (?, ?, ?, ?, ?, ?)",
			vh.ProjectId, vh.ProjectBuildPlanId, vh.UserId, vh.HookId, vh.Event, vh.Branch,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый VcsHook", err, vh)
			return false
		}

		vh.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового VcsHook'а", err, vh)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE vcs_hooks SET project_id = ?, project_build_plan_id = ?, user_id = ?, hook_id = ?, event = ?, branch = ? WHERE id = ?",
			vh.ProjectId, vh.ProjectBuildPlanId, vh.UserId, vh.HookId, vh.Event, vh.Branch, vh.Id,
		)
		if err != nil {
			log.Println("Не удалось обновить VcsHook", err, vh)
			return false
		}
	}

	return true
}

// Удалить VcsHook
func (vh *VcsHook) Delete() bool {
	db := database.Db()
	defer db.Close()

	_, err := db.Exec("DELETE FROM vcs_hooks WHERE id = ?", vh.Id)
	if err != nil {
		log.Println("Не удалось удалить VcsHook", err, vh)
		return false
	}

	return true
}

// Получить все хуки проекта
func GetVcsHooksByProjectId(projectId interface{}) []*VcsHook {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM vcs_hooks WHERE project_id = ?", projectId)
	if err != nil {
		log.Println("Не удалось найти все VcsHook'и проекта")
		return nil
	}
	defer rows.Close()

	return scanVcsHooks(rows)
}

// Получить hook по идентификатору
func GetVcsHookById(id interface{}) *VcsHook {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM vcs_hooks WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти VcsHook по ID", id)
		return nil
	}

	hook := scanVcsHook(row)

	return &hook
}
