package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Прослойка для фильтрации и останова долгих запросов
func filterMiddleware(mux *http.ServeMux) http.HandlerFunc {
	// Инициализация канала готовности и передача его по цепи запросов в контекст

	return func(w http.ResponseWriter, r *http.Request) {
		// Канал для возврата статуса
		ready := make(chan bool) // true - найден, false - не найден
		ctx := context.WithValue(r.Context(), "ready", ready)
		// Запись канала в контекст
		r = r.WithContext(ctx)

		// Принимаем только запросы с JSON телом
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "", 404)
			return
		}

		handler, p := mux.Handler(r)

		log.Printf("Path: %v", p)

		go func(ch chan bool) {
			w.Header().Set("Content-Type", "application/json")
			// Поиск хандлера среди маршрутов

			handler.ServeHTTP(w, r)
		}(ready)

		// Таймер дедлайна если длительность выше максимальной(5 сек)
		dead := time.NewTimer(time.Duration(5 * time.Second))

		// Возвращаем первый попавшийся канал
		select {
		case <-dead.C: // Канал дедлайна
			http.Error(w, "", 400)
			w.Header().Set("Content-Type", "application/json")
			resp, err := json.Marshal(map[string]string{
				"error": "timeout too long",
			})
			if err != nil {
				http.Error(w, "", 404)
				return
			}
			_, _ = w.Write(resp)
			return
		case found := <-ready: // Канал статуса
			switch found {
			case true:
				w.Header().Set("Content-Type", "application/json")
				resp, err := json.Marshal(map[string]string{
					"status": "ok",
				})
				if err != nil {
					http.Error(w, "", 404)
					return
				}
				_, _ = w.Write(resp)
			case false:
				w.WriteHeader(404)
				_, _ = w.Write(nil)
			}
			return
		}
	}
}
