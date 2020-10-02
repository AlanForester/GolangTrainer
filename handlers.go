package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Структура для входдящих параметров
type body struct {
	// Интерфейс используется для приема типа строка или целое
	Timeout interface{} `json:"timeout"` // Таймаут в миллисек
}

func slowHandle(w http.ResponseWriter, r *http.Request) {

	// Достаем с контекста канал для уведомления завершения функции
	ready := r.Context().Value("ready").(chan bool)

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "", 404)
		return
	}

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
			case int64:
				timeout = time.Duration(tm) * time.Millisecond
			case string:
				convTm, _ := strconv.Atoi(tm)
				timeout = time.Duration(convTm) * time.Millisecond
			}

			// Запуск таймера для ожидания
			timer := time.NewTimer(timeout)

			select {
			case <-timer.C:
				// Таймер сработал, передаем сообщение о успехе
				ready <- false
			}

			return
		}
	}
}

func timeoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ready := make(chan bool)

		ctx := context.WithValue(r.Context(), "ready", ready)
		req := r.WithContext(ctx)

		handler, p := http.DefaultServeMux.Handler(r)

		go handler.ServeHTTP(w, req)

		if p != "/" { // Not fp

			dead := time.NewTimer(time.Duration(5 * time.Second))

			select {
			case <-dead.C:
				w.WriteHeader(400)
				resp, err := json.Marshal(map[string]string{
					"error": "timeout too long",
				})
				if err != nil {
					http.Error(w, "", 404)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(resp)
				return
			case <-ready:
				resp, err := json.Marshal(map[string]string{
					"status": "ok",
				})
				if err != nil {
					http.Error(w, "", 404)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(resp)
				return
			}
		}
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	_, _ = w.Write(nil)
}