package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Структура для входящих параметров
type body struct {
	// Интерфейс используется для приема типа строка или целое
	Timeout interface{} `json:"timeout"` // Таймаут в миллисек
}

func slowHandle(w http.ResponseWriter, r *http.Request) {

	// Обрабатываем на пост методе
	if r.Method == "POST" {

		// Селектор для выбора маршрута
		switch r.URL.Path {
		case "/api/slow":
			bd, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "", 404)
				return
			}

			var b body
			err = json.Unmarshal(bd, &b)
			if err != nil {
				http.Error(w, "", 404)
				return
			}

			var timeout time.Duration
			// Определяем тип входящего параметра таймаут
			switch tm := b.Timeout.(type) {
			case float64:
				timeout = time.Duration(tm) * time.Millisecond
			case string:
				convTm, err := strconv.Atoi(tm)
				if err != nil {
					log.Printf("Errpr: %v", err)
				}
				timeout = time.Duration(convTm) * time.Millisecond
			}

			// Запуск таймера для ожидания
			timer := time.NewTimer(timeout)

			select {
			case <-timer.C:
				// Таймер сработал, передаем сообщение о успехе
				// Достаем с контекста канал для уведомления завершения функции
				if ch, ok := r.Context().Value("ready").(chan bool); ok {
					// Отправляем в канал статус "найдено"
					ch <- true
				}
			}
			return
		}
	}

	http.Error(w, "", 404)
}

// Обработчик по умолчанию
// Если маршрут не будет найден то отдаем 404 с пустым телом
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if ch, ok := r.Context().Value("ready").(chan bool); ok {
		// Отправляем в канал статус "не найдено"
		ch <- false
	}
	return
}
