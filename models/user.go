package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель пользователя
type User struct {
	Id       int64  // Идентификатор
	Login    string // Логин
	Password string // Пароль, хеш sha256
	Salt     string // Соль для криптинга jwt токена
}

// Создает модель пользователя по строке из базы
func scanUser(row *sql.Row) (user User) {
	err := row.Scan(
		&user.Id,
		&user.Login,
		&user.Password,
		&user.Salt,
	)
	if err != nil {
		log.Println("Не удалось собрать модель пользователя", row)
	}

	return
}

// Создает массив моделей пользователя по строкам из базы
func scanUsers(rows *sql.Rows) (users []*User) {
	for rows.Next() {
		user := User{}
		err := rows.Scan(
			&user.Id,
			&user.Login,
			&user.Password,
			&user.Salt,
		)
		if err != nil {
			log.Println("Не удалось собрать модель пользователя из массива строк", err)
			return
		}

		users = append(users, &user)
	}

	return
}

// Получить пользователя по логину и паролю
func GetUserByLoginAndPassword(login string, password string) *User {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM users WHERE login = ? AND password = ?", login, password)
	if row == nil {
		log.Println("Не удалось найти пользователя по логину и паролю", login, password)
		return nil
	}

	user := scanUser(row)

	return &user
}

// Получить пользователя по ID
func GetUserById(id interface{}) *User {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	if row == nil {
		log.Println("Не удалось найти пользователя по ID", id)
		return nil
	}

	user := scanUser(row)

	return &user
}
