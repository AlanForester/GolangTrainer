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

const (
	maxExecTime = 5 * time.Second
	reqConType  = "application/json"
)

type body struct {
	Timeout interface{} `json:"timeout"`
}

func slowHandle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ready := r.Context().Value("ready").(chan bool)
	//w.WriteHeader(200)

	if r.Header.Get("Content-Type") != reqConType {
		http.Error(w, "", 404)
		return
	}

	if r.Method == "POST" {
		switch r.URL.Path {
		case "/api/slow":
			bd, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "1", 404)
				return
			}
			var b body
			err = json.Unmarshal(bd, &b)
			if err != nil {
				http.Error(w, "2", 404)
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

		go http.DefaultServeMux.ServeHTTP(w, req)

		dead := time.NewTimer(time.Duration(maxExecTime))

		select {
		case <-dead.C:
			w.WriteHeader(400)
			resp, err := json.Marshal(map[string]string{
				"error": "timeout too long",
			})
			if err != nil {
				http.Error(w, "3", 404)
				return
			}
			w.Header().Set("Content-Type", reqConType)
			_, _ = w.Write(resp)
			return
		case <-ready:
			resp, err := json.Marshal(map[string]string{
				"status": "ok",
			})
			if err != nil {
				http.Error(w, "3", 404)
				return
			}
			w.Header().Set("Content-Type", reqConType)
			_, _ = w.Write(resp)
			return
		}
	}
}

func defaultHandler() http.Handler {
	return filterMiddleware()
}

func filterMiddleware() http.Handler {

	filter := timeoutHandler()

	return filter
}

func main() {

	//mux.HandleFunc("/", defaultHandle(mux))
	//filter := http.TimeoutHandler(filterMiddleware(http.HandlerFunc(slowHandle)), maxExecTime, "")
	http.HandleFunc("/api/slow", slowHandle)
	//mux.HandleFunc("/api/slow", slowHandle(mux))

	srv := http.Server{
		Addr:    ":8080",
		Handler: defaultHandler(),
	}

	defer srv.Close()

	fmt.Printf("Listening [127.0.0.1:8080]...\n")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}

}
