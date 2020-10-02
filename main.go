package main

import (
	"fmt"
	"net/http"
)

func main() {

	//filter := http.TimeoutHandler(filterMiddleware(http.HandlerFunc(slowHandle)), maxExecTime, "")
	http.HandleFunc("/api/slow", slowHandle)
	http.HandleFunc("/", defaultHandler)
	//mux.HandleFunc("/api/slow", slowHandle(mux))

	srv := http.Server{
		Addr:    ":8080",
		Handler: filterMiddleware(),
	}

	defer srv.Close()

	fmt.Printf("Listening [127.0.0.1:8080]...\n")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}

}
