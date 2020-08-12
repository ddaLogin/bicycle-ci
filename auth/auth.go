package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Сервис авторизации
type Service struct {
	sessionName string
	secretKey   string
	loginRoute  string // Пример - "/login"
	storage     *sessions.CookieStore
}

// JWT токен авторизации
type Token struct {
	userId string
	jwt.StandardClaims
}

// Конструктор сервиса авториации
func NewService(sessionName string, secretKey string, loginRoute string) *Service {
	service := Service{sessionName: sessionName, secretKey: secretKey, loginRoute: loginRoute}

	service.storage = sessions.NewCookieStore([]byte(secretKey))

	return &service
}

// Авторизовывает пользователя
func (s *Service) Auth(login string, password string, w http.ResponseWriter, req *http.Request) bool {
	if login == "Vcs-trigger" {
		return false
	}

	password = s.HashPassword(password)

	user := models.GetUserByLoginAndPassword(login, password)

	if user != nil && (models.User{}) != *user {
		session, err := s.storage.Get(req, s.sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}

		session.Values["id"] = user.Id
		session.Values["token"] = s.generateToken(user)

		// Save it before we write to the response/return from the handler.
		err = session.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}

		return true
	}

	return false
}

// Мидлеваре авторизации роутов
func (s *Service) AuthMiddleware(next func(w http.ResponseWriter, req *http.Request, user *models.User)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user := s.GetCurrentUser(req)

		if user == nil || (models.User{}) == *user {
			http.Redirect(w, req, s.loginRoute, http.StatusSeeOther)
			return
		}

		next(w, req, user)
	})
}

// Получить авторизованного пользователя
func (s *Service) GetCurrentUser(req *http.Request) (currentUser *models.User) {
	session, err := s.storage.Get(req, s.sessionName)
	if err != nil {
		log.Println("Не удалось получить сессию", err)
		return
	}

	id := session.Values["id"]
	token := session.Values["token"]
	user := models.GetUserById(fmt.Sprintf("%v", id))

	if user != nil && (models.User{}) != *user {
		if s.isValidToken(user, fmt.Sprintf("%v", token)) {
			currentUser = user
		}
	}

	return
}

// Проверка токена авторизации
func (s *Service) isValidToken(user *models.User, jwtToken string) bool {
	token := &Token{}

	tkn, err := jwt.ParseWithClaims(jwtToken, token, func(token *jwt.Token) (interface{}, error) {
		return []byte(user.Salt), nil
	})
	if err != nil {
		return false
	}

	return tkn.Valid
}

// Генерация токена авторизации
func (s *Service) generateToken(user *models.User) (authToken string) {
	expirationTime := time.Now().Add(120 * time.Minute)
	token := &Token{
		userId: fmt.Sprintf("%v", user.Id),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	authToken, err := tkn.SignedString([]byte(user.Salt))

	if err != nil {
		log.Println("Не удалось сгенерировать авторизационный токен", err)
		return
	}

	return
}

// Хеширование пароля
func (s *Service) HashPassword(inputPassword string) string {
	hashProvider := sha256.New()
	hashProvider.Write([]byte(inputPassword))

	return hex.EncodeToString(hashProvider.Sum(nil))
}

// Генерация соли
func (s *Service) GenerateSalt(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	str := make([]rune, n)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
