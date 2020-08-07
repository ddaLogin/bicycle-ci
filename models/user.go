package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель пользователя
type User struct {
	Id       int    // Идентификатор
	Login    string // Логин
	Password string // Пароль, хеш sha256
	Salt     string // Соль для криптинга jwt токена
}

// Получить пользователя по логину и паролю
func GetUserByLoginAndPassword(login string, password string) (user User) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM users WHERE login = ? AND password = ?", login, password)
	if err != nil {
		log.Println("Can't get user by login and password. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Login, &user.Password, &user.Salt)
		if err != nil {
			log.Println("Can't scan user by login and password. ", err)
			return
		}
	}

	return
}

// Получить пользователя по ID
func GetUserById(id string) (user User) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get user by ID. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Login, &user.Password, &user.Salt)
		if err != nil {
			log.Println("Can't scan user by ID. ", err)
			return
		}
	}

	return
}
