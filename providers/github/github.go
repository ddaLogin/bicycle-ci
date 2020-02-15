package github

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const host = "https://github.com"

// Авторизационный токен
type AccessToken struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	Scope     string `json:"scope"`
}

// Конфиг GitHub провайдера
type Config struct {
	ClientId     string
	ClientSecret string
}

var cfg Config

// Установка настроек
func SetConfig(c Config) {
	cfg = c
}

// Возвращает ссылку для авторизации в гитхабе
func GetOAuthLink() string {
	link, _ := url.Parse(host + "/login/oauth/authorize")
	query, _ := url.ParseQuery(link.RawQuery)

	query.Add("client_id", cfg.ClientId)
	query.Add("scope", "repo")

	link.RawQuery = query.Encode()

	return link.String()
}

// Запрашивает авторизационый токен
func GetAccessToken(code string) (token AccessToken) {
	link, _ := url.Parse(host + "/login/oauth/access_token")
	query, _ := url.ParseQuery(link.RawQuery)

	query.Add("client_id", cfg.ClientId)
	query.Add("client_secret", cfg.ClientSecret)
	query.Add("code", code)

	link.RawQuery = query.Encode()

	response, err := post(link.String(), []byte(``))
	if err != nil {
		return
	}

	err = json.Unmarshal(response, &token)
	if err != nil {
		log.Fatal("Can't parse access token from response", err, response)
		return
	}

	return
}

// Подписываемся на события репозитория
func CreatePushHook(owner string, repo string) {
	//url := fmt.Sprintf("%s/repost/%s/%s/hooks", host, owner, repo)
	//
	//query := []byte(`{
	//	"name": "web",
	//	"active": true,
	//	"events": [
	//		"push",
	//		"pull_request"
	//	],
	//	"config": {
	//		"url": "http://localhost:8090/github/hook2",
	//		"content_type": "json",
	//		"insecure_ssl": "0"
	//	}
	//}`)

	//result, _ := post(url, query)

	//var response interface{}

	//err := json.Unmarshal(result, &response)
	//if err != nil {
	//	return
	//}

	//fmt.Printf("%+v", string(result));
}

// Выполняет пост запрос
func post(url string, query []byte) (response []byte, err error) {
	// Generate request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(query))
	if err != nil {
		log.Fatal("Error request reading.", err)
		return
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error response reading. ", err)
		return
	}
	defer resp.Body.Close()

	// Parse response
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error body reading. ", err)
		return
	}

	return
}
