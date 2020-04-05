package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
	"strconv"
)

// Модель WebHook'а
type WebHook struct {
	Id        int64   // Идентификатор хука
	ProjectId int64   // Проект за которым закреплен хук
	UserId    int     // Создатель хука
	HookId    *string // Идентификатор хука на стороне провайдера
	Event     string  // Тип события при котором хук срабатывает
	Branch    string  // Целевая ветка хука
}

// Хелпер для генерации урла по которому хук будет трегериться
func (wh WebHook) GetTriggerUrl() string {
	return "https://ee859862.ngrok.io/hooks/trigger?hookId=" + strconv.Itoa(int(wh.Id))
}

// Сохранить WebHook
func (wh *WebHook) Save() bool {
	db := database.Db()
	defer db.Close()

	if wh.Id == 0 {
		result, err := db.Exec("INSERT INTO hooks (project_id, user_id, event, branch) VALUES (?, ?, ?, ?)",
			wh.ProjectId, wh.UserId, wh.Event, wh.Branch)
		if err != nil {
			log.Println("Can't insert Hook. ", err, wh)
			return false
		}

		wh.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE hooks SET project_id = ?, user_id = ?, hook_id = ?, event = ?, branch = ? WHERE id = ?",
			wh.ProjectId, wh.UserId, wh.HookId, wh.Event, wh.Branch, wh.Id)
		if err != nil {
			log.Println("Can't update Hook. ", err, wh)
			return false
		}

		return true
	}

	return false
}

// Удалить WebHook
func (wh WebHook) Delete() bool {
	db := database.Db()
	defer db.Close()

	if wh.Id == 0 {
		_, err := db.Exec("DELETE FROM hooks WHERE id = ?", wh.Id)
		if err != nil {
			log.Println("Can't delete Hook. ", err, wh)
			return false
		}

		return true
	} else {
		return true
	}

	return false
}

// Получить все хуки проекта
func GetHooksByProjectId(projectId string) (hooks []WebHook) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM hooks WHERE project_id = ?", projectId)
	if err != nil {
		log.Println("Can't get hooks by project id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		hook := WebHook{}
		err := rows.Scan(
			&hook.Id,
			&hook.ProjectId,
			&hook.UserId,
			&hook.HookId,
			&hook.Event,
			&hook.Branch,
		)
		if err != nil {
			log.Println("Can't scan hooks by project id. ", err)
			continue
		}

		hooks = append(hooks, hook)
	}

	return
}

// Получить hook по идентификатору
func GetHookById(id string) (hook WebHook) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM hooks WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get hook by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&hook.Id,
			&hook.ProjectId,
			&hook.UserId,
			&hook.HookId,
			&hook.Event,
			&hook.Branch,
		)
		if err != nil {
			log.Println("Can't scan hook by id. ", err)
			continue
		}
	}

	return
}
