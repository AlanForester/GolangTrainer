package main

import (
	"fmt"
	"net/http"
)

// Главная функция
func main() {

	// Обработчики путей
	http.HandleFunc("/api/slow", slowHandle)
	http.HandleFunc("/", defaultHandler) // Выберется если путь не найден или /

	// Инициализация сервера
	srv := http.Server{
		Addr:    ":8080",
		Handler: filterMiddleware(),
	}

	defer srv.Close()

	fmt.Printf("Listening [127.0.0.1:8080]...\n")
	if err := srv.ListenAndServe(); err != nil { // Слушаем порт
		fmt.Printf("Server failed: %s\n", err)
	}

}
