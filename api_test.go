package main

import (
	"bytes"
	"final/_vendor-20190519220328/github.com/gorilla/mux"
	"final/crud"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const WILLBEEXEC = 150

func mkstr(size int) string {
	if size == 0 {
		return mkstr(5)
	}
	str := make([]byte, size)
	for i := range str {
		rand.Seed(time.Now().UnixNano())
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}

type WrongCust struct {
	crud.Customer
	FirstName int
}

func TestCreateCustomer(t *testing.T) {
	var customers []crud.Customer
	for i := 0; i < WILLBEEXEC; i++ {
		customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)), Email: mkstr(rand.Intn(20)), Phone: mkstr(rand.Intn(15))})
		customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)), Email: mkstr(rand.Intn(20))})
	}
	customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
		LastName: mkstr(rand.Intn(20))})

	for _, c := range customers {
		testStat(&c, http.StatusCreated)
	}
}
func testStat(data []byte, expectStatus int, method string) {
	srv := httptest.NewServer(func() http.Handler {
		r := mux.NewRouter()
		r.HandleFunc("/", crud.CreateCustomer).Methods(http.MethodPost)
		return r
	}())
	defer srv.Close()

	if err != nil {
		t.Fatal(err)
	}
	if data == nil && method == http.MethodGet {
		httptest.NewRequest(http.MethodGet, "http://localhost:8080/customers/", nil)
	}
	if data != nil && method == http.MethodPost {
		req := httptest.NewRequest(method, "http://localhost:8080/customers/",
			bytes.NewBuffer(data))
		fmt.Println(string(data))
	}
	rec := httptest.NewRecorder()
	crud.CreateCustomer(rec, req)
	result := rec.Result()
	if result.StatusCode != http.StatusCreated {
		fmt.Printf("ERROR: While sending post expected status %d,got %d",
			expectStatus, result.StatusCode)
		return
	}
}
