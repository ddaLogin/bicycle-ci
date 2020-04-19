package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
	"strconv"
)

// Образ контейнера для сборки
type Image struct {
	Id          int64
	Name        string
	Description string
	UserId      int
}

// Получить модель пользователя
func (img *Image) User() User {
	return GetUserById(strconv.Itoa(int(img.UserId)))
}

// Сохранить образ
func (img *Image) Save() bool {
	db := database.Db()
	defer db.Close()

	if img.Id == 0 {
		result, err := db.Exec("insert into images (`name`, description, user_id) values (?, ?, ?)",
			img.Name, img.Description, img.UserId)
		if err != nil {
			log.Println("Can't insert container. ", err, img)
			return false
		}

		img.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE images SET `name` = ?, description = ?, user_id = ? WHERE id = ?",
			img.Name, img.Description, img.UserId, img.Id)
		if err != nil {
			log.Println("Can't update container. ", err, img)
			return false
		}

		return true
	}

	return false
}

// Получить список всех образов
func GetImages() (images []Image) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM images")
	if err != nil {
		log.Println("Can't get Images like list. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		image := Image{}
		err := rows.Scan(
			&image.Id,
			&image.Name,
			&image.Description,
			&image.UserId,
		)
		if err != nil {
			log.Println("Can't scan Images like list. ", err)
			continue
		}

		images = append(images, image)
	}

	return
}

// Получить образ по ID
func GetImageById(imageId string) (image Image) {
	db := database.Db()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM images WHERE id = ?", imageId)
	if err != nil {
		log.Println("Can't get Images by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&image.Id,
			&image.Name,
			&image.Description,
			&image.UserId,
		)
		if err != nil {
			log.Println("Can't scan Images by id. ", err)
			continue
		}
	}

	return
}
