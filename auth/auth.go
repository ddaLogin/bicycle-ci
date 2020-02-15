package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
	"time"
)

// Пользователь
type User struct {
	Login    string
	Password string
	Salt     string
	Secret   string
}

// JWT токен авторизации
type AuthToken struct {
	Secret string
	jwt.StandardClaims
}

// База пользователей
var users = make(map[interface{}]User)

// Сессия
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

// Загрузка пользователей в память
func SetUsers(usersArray []User) {
	for _, user := range usersArray {
		user.Salt = saltGenerate()
		user.Secret = saltGenerate()
		users[user.Login] = user
	}
}

// Авторизация пользователя
func Auth(w http.ResponseWriter, req *http.Request) (result bool, err error) {
	login := req.FormValue("login")
	password := req.FormValue("password")

	if user, ok := users[login]; ok {
		if checkPassword(user.Password, password) {

			session, err := store.Get(req, "session-name")
			session.Values["login"] = user.Login
			session.Values["token"] = generateAuthToken(user.Salt, user.Secret)

			// Save it before we write to the response/return from the handler.
			err = session.Save(req, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return result, err
			}

			result = true
		}
	}

	return
}

// Получить авторизованного пользователя
func GetCurrentUser(req *http.Request) (currentUser User) {
	session, err := store.Get(req, "session-name")
	if err != nil {
		log.Println("Can't check auth.", err)

		return
	}

	login := session.Values["login"]
	token := session.Values["token"]

	if user, ok := users[login]; ok {
		if checkAuthToken(user.Salt, fmt.Sprintf("%v", token), user.Secret) {
			currentUser = user
		}
	}

	return
}

// Проверка пароля
func checkPassword(password string, inputPassword string) bool {
	hasher := sha256.New()
	hasher.Write([]byte(inputPassword))

	return password == hex.EncodeToString(hasher.Sum(nil))
}

// Генерация соли для токена
func saltGenerate() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Генерация токена авторизации
func generateAuthToken(salt string, secret string) (authToken string) {
	expirationTime := time.Now().Add(5 * time.Minute)
	auth := &AuthToken{
		Secret: secret,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth)
	authToken, err := token.SignedString([]byte(salt))

	if err != nil {
		log.Println("Can't generate new auth token", err)
		return
	}

	return
}

// Проверка токена авторизации
func checkAuthToken(salt string, token string, secret string) bool {
	auth := &AuthToken{}

	tkn, err := jwt.ParseWithClaims(token, auth, func(token *jwt.Token) (interface{}, error) {
		return []byte(salt), nil
	})
	if err != nil {
		return false
	}

	return tkn.Valid && auth.Secret == secret
}
