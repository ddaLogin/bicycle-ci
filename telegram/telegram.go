package telegram

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Настройка бота
type Config struct {
	Token  string
	Host   string
	ChatId string
}

// Ответ от телеграма
type Response struct {
	Response []byte
	Status   int
}

// Сервис для работы с Telegram
type Service struct {
	config Config
}

// Конструктор telegram сервиса
func NewService(config Config) *Service {
	return &Service{config: config}
}

// Отправить сообщение в чат
func (s *Service) SendMessage(message string) {
	url := s.config.Host + "/bot" + s.config.Token + "/sendMessage"
	query := []byte(`{"chat_id": "` + s.config.ChatId + `", "text": "` + message + `", "parse_mode": "Markdown"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(query))
	if err != nil {
		log.Fatal("Ошибка при отправке сообщения в Telegram. ", err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Ошибка при чтение ответа от Telegram. ", err)
		return
	}
	defer resp.Body.Close()

	response := Response{}

	// Parse response
	response.Status = resp.StatusCode
	response.Response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Ошибка при чтение тела ответа от Telegram. ", err)
		return
	}

	return
}
