package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSadPath(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.

	req, err := http.NewRequest("GET", "/api/slow", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := defaultHandler()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

}

func TestHappyPath(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	handler := timeoutHandler()
	slowServer := httptest.NewServer(handler)
	res := httptest.NewRecorder()
	form := url.Values{

		"timeout": []string{"100"},
	}
	js, _ := json.Marshal(form)
	//readBody := bufio.NewReader(bytes.NewReader(js))
	req := httptest.NewRequest("POST", "/api/slow", bytes.NewReader(js))
	//req.URL.Parse("/api/slow")
	//req.RequestURI = slowServer.URL+"/api/slow"
	req.Header.Add("Content-Type", "application/json")

	slowServer.Client().Do(req)
	handler.ServeHTTP(res, req)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//req.Header.Add("Content-Type", "application/json")
	//log.Printf("Req: %#v", resul)
	//res := httptest.NewRecorder()
	//handler := defaultHandler()
	//handler.ServeHTTP(res, req)
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.

	got := res.Code
	log.Printf("S: %v", got)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("B: %v", string(body))
	//req := httptest.NewRequest("POST", "/api/slow", readBody)
	//req.Header.Add("Content-Type", "application/json")
	//_ = req.Write(json.NewEncoder(io.WriteString(new(io.ReadWriter))))
	//handler.ServeHTTP(rr, req)
	//res, err := http.PostForm("/api/slow",form)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.

	// Check the status code is what we expect.
	if res.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			res.Code, http.StatusOK)
	}
	//
	//slowServer.Close()
	//body, err := ioutil.ReadAll(res.Body) // считываем body ответа
	// Check the response body is what we expect.
	//expected := `{"status": "ok"}`
	//if string(body) != expected {
	//	t.Errorf("handler returned unexpected body: got %v want %v",
	//		rr.Body.String(), expected)
	//}
}
