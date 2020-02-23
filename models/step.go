package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

const STEP_STATUS_FAILED = 0  // Шаг заверишлся ошибкой
const STEP_STATUS_SUCCESS = 1 // Шаг прошел успешно

// Один шаг из сборки
type Step struct {
	Id      int64
	BuildId int64
	Name    string
	StdOut  string
	StdErr  string
	Error   string
	Status  int
}

// Сохранить Шаг
func (st Step) Save() bool {
	db := database.Db()
	defer db.Close()

	if st.Id == 0 {
		result, err := db.Exec("insert into steps (build_id, `name`, std_out, std_err, error, status) values (?, ?, ?, ?, ?, ?)",
			st.BuildId, st.Name, st.StdOut, st.StdErr, st.Error, st.Status)
		if err != nil {
			log.Println("Can't insert step. ", err, st)
			return false
		}

		st.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE steps SET build_id = ?, `name` = ?, std_out = ?, std_err = ?, error = ?, status = ? WHERE id = ?",
			st.BuildId, st.Name, st.StdOut, st.StdErr, st.Error, st.Status, st.Id)
		if err != nil {
			log.Println("Can't update step. ", err, st)
			return false
		}

		return true
	}

	return false
}

// Получить шаги билда
func GetStepsByBuildId(buildId int64) (steps []Step) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM steps WHERE build_id = ?", buildId)
	if err != nil {
		log.Println("Can't get steps by build id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		step := Step{}
		err := rows.Scan(
			&step.Id,
			&step.BuildId,
			&step.Name,
			&step.StdOut,
			&step.StdErr,
			&step.Error,
			&step.Status,
		)
		if err != nil {
			log.Println("Can't scan steps by build id. ", err)
			continue
		}

		steps = append(steps, step)
	}

	return
}
