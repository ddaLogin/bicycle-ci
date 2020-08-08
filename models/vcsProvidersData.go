package models

import (
	"github.com/ddalogin/bicycle-ci/database"
	"log"
)

// Модель подлюченного провайдера
type VcsProviderData struct {
	Id                   int64  // Идентификатор провайдера
	UserId               int64  // Идентификатор пользователя
	ProviderType         int    // Тип провайдера
	ProviderAuthToken    string // Авторизационый токен для АПИ
	ProviderAccountId    int    // Идентификатор аккаунта на стороне провайдера
	ProviderAccountLogin string // Логин аккаунта на стороне провайдера
}

// Сохранить данные о провайдере
func (pd *VcsProviderData) Save() bool {
	db := database.Db()
	defer db.Close()

	if pd.Id == 0 {
		result, err := db.Exec("insert into vcs_providers (user_id, provider_type, provider_auth_token, provider_account_id, provider_account_login) values (?, ?, ?, ?, ?)",
			pd.UserId, pd.ProviderType, pd.ProviderAuthToken, pd.ProviderAccountId, pd.ProviderAccountLogin)
		if err != nil {
			log.Println("Can't insert VcsProviderData. ", err, pd)
			return false
		}

		pd.Id, _ = result.LastInsertId()

		return true
	} else {
		_, err := db.Exec("UPDATE vcs_providers SET user_id = ?, provider_type = ?, provider_auth_token = ?, provider_account_id = ?, provider_account_login = ? WHERE id = ?",
			pd.UserId, pd.ProviderType, pd.ProviderAuthToken, pd.ProviderAccountId, pd.ProviderAccountLogin, pd.Id)
		if err != nil {
			log.Println("Can't update VcsProviderData. ", err, pd)
			return false
		}

		return true
	}

	return false
}

// Получение подключенного провайдера пользователя по типу
func GetProviderDataByUserAndType(userId interface{}, providerType int) (provider VcsProviderData) {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM vcs_providers WHERE user_id = ? AND provider_type = ?", userId, providerType)
	if err != nil {
		log.Println("Can't get provider by user id and type. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&provider.Id,
			&provider.UserId,
			&provider.ProviderType,
			&provider.ProviderAuthToken,
			&provider.ProviderAccountId,
			&provider.ProviderAccountLogin,
		)
		if err != nil {
			log.Println("Can't scan provider by user id and type. ", err)
			return
		}
	}

	return
}

// Получение провайдера по ID
func GetProviderDataById(id interface{}) (provider VcsProviderData) {
	db := database.Db()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM vcs_providers WHERE id = ?", id)
	if err != nil {
		log.Println("Can't get provider by id. ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&provider.Id,
			&provider.UserId,
			&provider.ProviderType,
			&provider.ProviderAuthToken,
			&provider.ProviderAccountId,
			&provider.ProviderAccountLogin,
		)
		if err != nil {
			log.Println("Can't scan provider by id. ", err)
			return
		}
	}

	return
}
