package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Прослойка для фильтрации и останова долгих запросов
func filterMiddleware() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Инициализация канала готовности и передача его по цепи запросов в контекст
		ready := make(chan bool)
		ctx := context.WithValue(r.Context(), "ready", ready)
		req := r.WithContext(ctx)

		// Поиск хандлера среди маршрутов
		handler, p := http.DefaultServeMux.Handler(r)

		go handler.ServeHTTP(w, req) // Запуск обработчика в фоне

		// Если путь не найден, не запускаем таймер для протухания
		if p != "/" {

			// Таймер дедлайна если длительность выше максимальной(5 сек)
			dead := time.NewTimer(time.Duration(5 * time.Second))

			// Возвращаем первый попавшийся канал
			select {
			case <-dead.C: // Канал дедлайна
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
			case <-ready: // Канал успеха
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
