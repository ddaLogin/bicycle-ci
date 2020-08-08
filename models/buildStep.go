package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

const StepStatusRunning = 0 // Шаг в процессе
const StepStatusFailed = 1  // Шаг заверишлся ошибкой
const StepStatusSuccess = 2 // Шаг прошел успешно

// Один шаг из сборки
type BuildStep struct {
	Id      int64
	BuildId int64
	Name    string
	StdOut  string
	StdErr  string
	Error   string
	Status  int
	build   *Build
}

// Возвращает модель процесса всей сборки
func (st *BuildStep) GetBuild() *Build {
	if st.build == nil {
		st.SetBuild(GetBuildById(st.BuildId))
	}

	return st.build
}

// Устанавливает модель процесса всей сборки
func (st *BuildStep) SetBuild(build *Build) {
	st.build = build
}

// Возвращает человекопонятное описание статуса шага
func (st *BuildStep) GetStatusTitle() string {
	if st.Status == StepStatusFailed {
		return "Остановлен с ошибкой"
	} else if st.Status == StepStatusSuccess {
		return "Успешно"
	}

	return "В процессе"
}

// Создает модель шага сборки по строке из базы
func scanBuildStep(row *sql.Row) (step BuildStep) {
	err := row.Scan(
		&step.Id,
		&step.BuildId,
		&step.Name,
		&step.StdOut,
		&step.StdErr,
		&step.Error,
		&step.Status,
	)
	if err != nil {
		log.Println("Не удалось собрать модель шага сборки", row)
	}

	return
}

// Создает массив моделей шага сборки по строкам из базы
func scanBuildSteps(rows *sql.Rows) (steps []*BuildStep) {
	for rows.Next() {
		step := BuildStep{}
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
			log.Println("Не удалось собрать модель шага сборки из массива строк", err)
			return
		}

		steps = append(steps, &step)
	}

	return
}

// Сохранить шаг сборки
func (st *BuildStep) Save() bool {
	db := database.Db()
	defer db.Close()

	if st.Id == 0 {
		result, err := db.Exec(
			"insert into build_steps (build_id, `name`, std_out, std_err, error, status) values (?, ?, ?, ?, ?, ?)",
			st.BuildId, st.Name, st.StdOut, st.StdErr, st.Error, st.Status,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый шаг сборки", err, st)
			return false
		}

		st.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового шага сборки", err, st)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE build_steps SET build_id = ?, `name` = ?, std_out = ?, std_err = ?, error = ?, status = ? WHERE id = ?",
			st.BuildId, st.Name, st.StdOut, st.StdErr, st.Error, st.Status, st.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить шаг сборки", err, st)
			return false
		}
	}

	return true
}

// Получить шаги билда
func GetStepsByBuildId(buildId interface{}) []*BuildStep {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM build_steps WHERE build_id = ?", buildId)
	if err != nil {
		log.Println("Не удалось найти все докер образы")
		return nil
	}
	defer rows.Close()

	return scanBuildSteps(rows)
}
