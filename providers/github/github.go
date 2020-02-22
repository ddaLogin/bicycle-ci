package github

import (
	"bicycle-ci/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Конфиг GitHub провайдера
type Config struct {
	ClientId     string
	ClientSecret string
	OAuthHost    string
	ApiHost      string
	Image        string
}

var config Config

// Установка настроек
func SetConfig(cfg Config) {
	config = cfg
}

// Ответ от GitHub'а
type GitHubResponse struct {
	Response []byte
	Status   int
}

// Авторизационный АПИ токен
type GitHubAccessToken struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	Scope     string `json:"scope"`
}

// Пользователь GitHab'а
type GitHubUser struct {
	Id    int    `json:"id"`
	Login string `json:"login"`
}

// GitHub репозиторий
type GitHubRepo struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
		Id    int    `json:"id"`
	} `json:"owner"`
}

// Ключ деплоя на стороне гитхаба
type GitHubDeployKey struct {
	Id        int    `json:"id"`
	Url       string `json:"url"`
	Title     string `json:"title"`
	Verified  bool   `json:"verified"`
	CreatedAt string `json:"created_at"`
	ReadOnly  bool   `json:"read_only"`
}

// GitHub провайдер
type GitHub struct {
	Data models.ProviderData
}

// Название провайдера
func (gh *GitHub) SetProviderData(providerData models.ProviderData) {
	gh.Data = providerData
}

// Название провайдера
func (gh GitHub) GetTitle() string {
	return "GitHub"
}

// Ссылка на картинку
func (gh GitHub) GetImageUrl() string {
	return config.Image
}

// Генерация ссылки для OAuth авторизации
func (gh GitHub) GetAuthLink() string {
	link, _ := url.Parse(config.OAuthHost + "/login/oauth/authorize")
	query, _ := url.ParseQuery(link.RawQuery)

	query.Add("client_id", config.ClientId)
	query.Add("scope", "repo")

	link.RawQuery = query.Encode()

	return link.String()
}

// Обработка oAuth авторизации
func (gh GitHub) OAuthCallback(req *http.Request) string {
	return getAccessToken(req.URL.Query().Get("code"))
}

// Запрос на основную информацию аккаунта
func (gh GitHub) UpdateProviderData(provider *models.ProviderData) {
	response, err := get(config.ApiHost+"/user", make(map[string]string), provider.ProviderAuthToken)
	if err != nil {
		return
	}

	user := GitHubUser{}
	err = json.Unmarshal(response.Response, &user)
	if err != nil {
		log.Fatal("Can't parse GitHub provider data from response. ", err, string(response.Response))
		return
	}

	provider.ProviderAccountId = user.Id
	provider.ProviderAccountLogin = user.Login
}

// Загрузить список репозиториев
func (gh GitHub) LoadProjects() (projects map[int]*models.Project) {
	repos := getRepositories(gh.Data.ProviderAuthToken)
	projects = make(map[int]*models.Project)

	for _, value := range repos {
		project := models.Project{
			UserId:        gh.Data.UserId,
			Name:          value.FullName,
			Provider:      gh.Data.Id,
			RepoId:        value.Id,
			RepoName:      value.Name,
			RepoOwnerName: value.Owner.Login,
			RepoOwnerId:   strconv.Itoa(value.Owner.Id),
		}

		projects[project.RepoId] = &project
	}

	return
}

// Загрузить один репозиторий
func (gh GitHub) LoadProjectToEnable(ownerName string, repoName string) (project models.Project) {
	repo := getRepository(ownerName, repoName, gh.Data.ProviderAuthToken)

	project.UserId = gh.Data.UserId
	project.Name = repo.FullName
	project.Provider = gh.Data.Id
	project.RepoId = repo.Id
	project.RepoName = repo.Name
	project.RepoOwnerName = repo.Owner.Login
	project.RepoOwnerId = strconv.Itoa(repo.Owner.Id)

	return
}

// Загружает на сервер VCS деплой ключ
func (gh GitHub) UploadProjectDeployKey(keyName string, key string, project models.Project) int {
	url := fmt.Sprintf("%v/repos/%v/%v/keys", config.ApiHost, project.RepoOwnerName, project.RepoName)
	body := []byte(`{"title": "` + keyName + `", "key": "` + strings.TrimSpace(key) + `", "read_only": "true"}`)

	response, err := post(url, body, gh.Data.ProviderAuthToken)
	if err != nil {
		return 0
	}

	if 201 != response.Status {
		log.Println("Error while uploading deploy key. ", response.Status, string(response.Response), string(body))

		return 0
	}

	deployKey := GitHubDeployKey{}

	err = json.Unmarshal(response.Response, &deployKey)
	if err != nil {
		log.Println("Can't parse GitHub deploy key from response. ", err, string(response.Response))
		return 0
	}

	return deployKey.Id
}

// Запрашивает авторизационый токен
func getAccessToken(code string) (token string) {
	link, _ := url.Parse(config.OAuthHost + "/login/oauth/access_token")
	query, _ := url.ParseQuery(link.RawQuery)

	query.Add("client_id", config.ClientId)
	query.Add("client_secret", config.ClientSecret)
	query.Add("code", code)

	link.RawQuery = query.Encode()

	response, err := post(link.String(), []byte(``), "")
	if err != nil {
		return
	}

	accessToken := GitHubAccessToken{}

	err = json.Unmarshal(response.Response, &accessToken)
	if err != nil {
		log.Fatal("Can't parse GitHub access token from response. ", err, string(response.Response))
		return
	}

	token = accessToken.Token

	return
}

// Подгружает список репозиториев
func getRepositories(token string) (repos []GitHubRepo) {
	params := make(map[string]string)

	response, err := get(config.ApiHost+"/user/repos", params, token)
	if err != nil {
		return
	}

	err = json.Unmarshal(response.Response, &repos)
	if err != nil {
		log.Fatal("Can't parse user repos from response. ", err, string(response.Response))
		return
	}

	return
}

// Подгружает список репозиториев
func getRepository(ownerLogin string, repoName string, token string) (repo GitHubRepo) {
	response, err := get(fmt.Sprintf("%v/repos/%v/%v", config.ApiHost, ownerLogin, repoName), make(map[string]string), token)
	if err != nil {
		return
	}

	err = json.Unmarshal(response.Response, &repo)
	if err != nil {
		log.Fatal("Can't parse user one repo from response. ", err, string(response.Response))
		return
	}

	return
}

//// Подписываемся на события репозитория
//func CreatePushHook(owner string, repo string) (hook Hook) {
//	url := fmt.Sprintf("%v/repos/%v/%v/hooks", apiHost, owner, repo)
//
//	response, err := post(url, []byte(`{
//		"name": "web",
//		"active": true,
//		"events": [
//			"push",
//		],
//		"config": {
//			"url": "https://localhost:8090/hook/` + owner + `/` + repo + `",
//			"content_type": "json",
//			"insecure_ssl": "0"
//		}
//	}`))
//	if err != nil {
//		return
//	}
//
//	err = json.Unmarshal(response, &hook)
//	if err != nil {
//		log.Fatal("Can't parse hook response. ", err, string(response))
//		return
//	}
//
//	return
//}

// Выполняет POST запрос
func post(url string, query []byte, token string) (response GitHubResponse, err error) {
	// Generate request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(query))
	if err != nil {
		log.Fatal("Error POST request reading. ", err)
		return
	}

	return send(req, token)
}

// Выполняет GET запрос
func get(baseUrl string, params map[string]string, token string) (response GitHubResponse, err error) {
	link, _ := url.Parse(baseUrl)
	query, _ := url.ParseQuery(link.RawQuery)

	for key, value := range params {
		query.Add(key, value)
	}

	link.RawQuery = query.Encode()

	// Generate request
	req, err := http.NewRequest("GET", link.String(), bytes.NewBuffer([]byte(``)))
	if err != nil {
		log.Fatal("Error GET request reading. ", err)
		return
	}

	return send(req, token)
}

// Выполняет отправку запроса и обработку ответа
func send(req *http.Request, token string) (response GitHubResponse, err error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	if "" != token {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error response reading. ", err)
		return
	}
	defer resp.Body.Close()

	// Parse response
	response.Status = resp.StatusCode
	response.Response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error body reading. ", err)
		return
	}

	return
}
