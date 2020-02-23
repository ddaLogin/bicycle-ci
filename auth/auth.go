package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"time"
)

// JWT токен авторизации
type Token struct {
	userId string
	jwt.StandardClaims
}

// Сессия
var sessionName = "bicycle-session"
var store = sessions.NewCookieStore([]byte("QWDqwdiqwdoqid12d1dqwd"))

// Middleware авторизации
func RequireAuthentication(next func(w http.ResponseWriter, req *http.Request, user models.User)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user := GetCurrentUser(req)

		if (models.User{}) == user {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		next(w, req, user)
	})
}

// Авторизация пользователя
func Auth(w http.ResponseWriter, req *http.Request) (result bool, err error) {
	login := req.FormValue("login")
	password := hashPassword(req.FormValue("password"))
	user := models.GetUserByLoginAndPassword(login, password)

	if (models.User{}) != user {
		session, err := store.Get(req, sessionName)
		session.Values["id"] = user.Id
		session.Values["token"] = generateToken(user)

		// Save it before we write to the response/return from the handler.
		err = session.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return result, err
		}

		result = true
	}

	return
}

// Получить авторизованного пользователя
func GetCurrentUser(req *http.Request) (currentUser models.User) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		log.Println("Can't check auth. ", err)
		return
	}

	id := session.Values["id"]
	token := session.Values["token"]
	user := models.GetUserById(fmt.Sprintf("%v", id))

	if (models.User{}) != user {
		if isValidToken(user, fmt.Sprintf("%v", token)) {
			currentUser = user
		}
	}

	return
}

// Генерация токена авторизации
func generateToken(user models.User) (authToken string) {
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
		log.Println("Can't generate new auth token", err)
		return
	}

	return
}

// Проверка токена авторизации
func isValidToken(user models.User, jwtToken string) bool {
	token := &Token{}

	tkn, err := jwt.ParseWithClaims(jwtToken, token, func(token *jwt.Token) (interface{}, error) {
		return []byte(user.Salt), nil
	})
	if err != nil {
		return false
	}

	return tkn.Valid
}

// Хеширование пароля
func hashPassword(inputPassword string) string {
	hasher := sha256.New()
	hasher.Write([]byte(inputPassword))

	return hex.EncodeToString(hasher.Sum(nil))
}
