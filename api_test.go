package main

import (
	"bytes"
	"encoding/json"
	"final/crud"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
const WILLBEEXEC = 150
func mkstr(size int) string {
	if size==0{
		return mkstr(5)
	}
	str:=make([]byte,size)
	for i:=range str{
		rand.Seed(time.Now().UnixNano())
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
func TestCreateCustomer(t *testing.T) {
	var customers []crud.Customer
	for i := 0; i < 100; i++ {
		customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)), Email: mkstr(rand.Intn(20)), Phone: mkstr(rand.Intn(15)),})
		customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)), Email: mkstr(rand.Intn(20))})
	}
	customers = append(customers, crud.Customer{FirstName: mkstr(rand.Intn(20)),
		LastName: mkstr(rand.Intn(20)),})
}
func testStat(c *crud.Customer,expectStatus int){
	var bytearr []byte
	var err error
		bytearr, err = json.Marshal(c)
		if err != nil {
			log.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/customers/",
			bytes.NewBuffer(bytearr))
		fmt.Println(string(bytearr))
		rec := httptest.NewRecorder()
		crud.CreateCustomer(rec, req)
		result := rec.Result()
		if result.StatusCode != http.StatusCreated {
			fmt.Printf("ERROR: While sending post expected status %d,got %d",
				expectStatus, result.StatusCode)
			return
		}
}