package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type body struct {
	Timeout interface{} `json:"timeout"`
}

func slowHandle(w http.ResponseWriter, r *http.Request) {
	//defer r.Body.Close()

	ready := r.Context().Value("ready").(chan bool)
	//w.WriteHeader(200)

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "", 404)
		return
	}

	if r.Method == "POST" {
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
			switch tm := b.Timeout.(type) {
			case int64:
				timeout = time.Duration(tm) * time.Millisecond
			case string:
				convTm, _ := strconv.Atoi(tm)
				timeout = time.Duration(convTm) * time.Millisecond
			}
			timer := time.NewTimer(timeout)

			select {
			case <-timer.C:
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

func filterMiddleware() http.Handler {

	filter := timeoutHandler()

	return filter
}

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
