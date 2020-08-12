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

// Сохранить пользователя
func (u *User) Save() bool {
	db := database.Db()
	defer db.Close()

	if u.Id == 0 {
		result, err := db.Exec(
			"insert into users (login, password, salt) values (?, ?, ?)",
			u.Login, u.Password, u.Salt,
		)

		if err != nil {
			log.Println("Не удалось сохранить нового пользователя", err, u)
			return false
		}

		u.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового пользователя", err, u)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE users SET login = ?, password = ?, salt = ? WHERE id = ?",
			u.Login, u.Password, u.Salt, u.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить пользователя", err, u)
			return false
		}
	}

	return true
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

// Получить пользователя по логину
func GetUserByLogin(login interface{}) *User {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM users WHERE login = ?", login)
	if row == nil {
		log.Println("Не удалось найти пользователя по login", login)
		return nil
	}

	user := scanUser(row)

	return &user
}

// Получить всех пользователей
func GetAllUsers() []*User {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Println("Не удалось найти всех пользователей")
		return nil
	}
	defer rows.Close()

	return scanUsers(rows)
}

//Получить пользователя по vcs id
func GetUserByVcsId(vcsId interface{}) *User {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT usr.* FROM users as usr JOIN vcs_providers vp on usr.id = vp.user_id WHERE vp.provider_account_id = ?", vcsId)
	if row == nil {
		log.Println("Не удалось найти пользователя по vcsId", vcsId)
		return nil
	}

	user := scanUser(row)

	return &user
}
