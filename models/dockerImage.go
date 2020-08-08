package models

import (
	"database/sql"
	"github.com/ddalogin/bicycle-ci/database"
	"log"
	"strconv"
)

// Образ контейнера для сборки
type DockerImage struct {
	Id          int64
	Name        string
	Description string
	UserId      int64
}

// Создает модель докер контейнера по строке из базы
func scanDockerImage(row *sql.Row) (image DockerImage) {
	err := row.Scan(
		&image.Id,
		&image.Name,
		&image.Description,
		&image.UserId,
	)
	if err != nil {
		log.Println("Не удалось собрать модель докер контейнера", row)
	}

	return
}

// Создает массив моделей докер контейнера по строкам из базы
func scanDockerImages(rows *sql.Rows) (images []*DockerImage) {
	for rows.Next() {
		image := DockerImage{}
		err := rows.Scan(
			&image.Id,
			&image.Name,
			&image.Description,
			&image.UserId,
		)
		if err != nil {
			log.Println("Не удалось собрать модель докер контейнера из массива строк", err)
			return
		}

		images = append(images, &image)
	}

	return
}

// Получить модель пользователя
func (img *DockerImage) User() *User {
	return GetUserById(strconv.Itoa(int(img.UserId)))
}

// Сохранить докер образ
func (img *DockerImage) Save() bool {
	db := database.Db()
	defer db.Close()

	if img.Id == 0 {
		result, err := db.Exec(
			"insert into docker_images (`name`, description, user_id) values (?, ?, ?)",
			img.Name, img.Description, img.UserId,
		)

		if err != nil {
			log.Println("Не удалось сохранить новый докер образ", err, img)
			return false
		}

		img.Id, err = result.LastInsertId()
		if err != nil {
			log.Println("Не удалось получить ID нового докер образа", err, img)
			return false
		}
	} else {
		_, err := db.Exec(
			"UPDATE docker_images SET `name` = ?, description = ?, user_id = ? WHERE id = ?",
			img.Name, img.Description, img.UserId, img.Id,
		)

		if err != nil {
			log.Println("Не удалось обновить докер образа", err, img)
			return false
		}
	}

	return true
}

// Получить список всех образов
func GetAllDockerImages() []*DockerImage {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM docker_images")
	if err != nil {
		log.Println("Не удалось найти все докер образы")
		return nil
	}
	defer rows.Close()

	return scanDockerImages(rows)
}

// Получить докер образ по ID
func GetDockerImageById(imageId interface{}) *DockerImage {
	db := database.Db()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM docker_images WHERE id = ?", imageId)
	if row == nil {
		log.Println("Не удалось найти докер образ по ID", imageId)
		return nil
	}

	image := scanDockerImage(row)

	return &image
}
