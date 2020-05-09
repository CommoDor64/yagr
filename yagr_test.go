package yagr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestExtractSinglePath(t *testing.T) {
// 	path := "/path/to/file"
// 	p, _ := extractSinglePath(&path)
// 	if p != "path" {
// 		t.Fatal("path is incorrect")
// 	}
// 	fmt.Println(p)
// }
func TestRouter_Insert(t *testing.T) {
	ro := NewRouter()

	// ro.Insert("/path/to/file")
	ro.Insert("/anothr/{param:value}/file", "GET", func(w http.ResponseWriter, r *http.Request) {
		log.Println("test")
		w.Write([]byte("200"))
	})
}
func TestRouter_Search(t *testing.T) {
	ro := NewRouter()

	// ro.Insert("/path/to/file")
	ro.Insert("/another/{param:int}/file", "GET", func(w http.ResponseWriter, r *http.Request) {
		log.Println("test")
		w.Write([]byte("200"))
	})

	// test existing route with query param
	_, err := ro.Search("/another/3/file")
	if err != nil {
		t.Fatal("fail")
	}

	// test non existing route
	_, err = ro.Search("/anotheri/file")
	if err == nil {
		t.Fatal("fail")
	}
}

type body struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestYAGR(t *testing.T) {
	yagr := NewYAGR()
	yagr.Get("/", func(w http.ResponseWriter, r *http.Request) {

		log.Println(r.Context().Value(key("params")))
		w.Write([]byte(r.URL.Path))
	})

	yagr.Get("/path/{param:int}/all", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Context().Value(key("params")))
		w.Write([]byte(r.URL.Path))
	})

	yagr.Get("/path/{param:int}/all/{otherparam:int}", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Context().Value(key("params")))
		w.Write([]byte(r.URL.Path))
	})

	yagr.Get("/path/{param:int}", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Context().Value(key("params")))
		w.Write([]byte(r.URL.Path))
	})

	yagr.Post("/path", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Context().Value(key("params")))
		var body body
		b, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(b, &body)
		log.Println(body)
		w.Write([]byte(r.URL.Path))
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	yagr.ServeHTTP(w, r)
	resp := w.Result()
	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path/3", nil)
	yagr.ServeHTTP(w, r)
	resp = w.Result()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path/3/all", nil)
	yagr.ServeHTTP(w, r)
	resp = w.Result()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path/3/all/5", nil)
	yagr.ServeHTTP(w, r)
	resp = w.Result()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}
