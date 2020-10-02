package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSadPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(defaultHandler))

	req, err := http.NewRequest("GET", "/api", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := filterMiddleware(mux)

	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

}

func TestHappyPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/api/slow", http.HandlerFunc(slowHandle))

	handler := filterMiddleware(mux)
	res := httptest.NewRecorder()
	form := url.Values{
		"timeout": []string{"1000"},
	}

	js, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/api/slow", bytes.NewReader(js))
	req.Header.Add("Content-Type", "application/json")

	handler.ServeHTTP(res, req)

	got := res.Code

	if got != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			got, http.StatusOK)
	}

	body := res.Body.String()
	expected := `{"status":"ok"}`
	if string(body) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			body, expected)
	}
}
