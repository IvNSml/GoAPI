package main

import (
	"bou.ke/monkey"
	"bytes"
	"encoding/json"
	"final/_vendor-20190519220328/github.com/gorilla/mux"
	"final/accounts"
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

type WrongCust struct {
	ID        int
	FirstName int `json:"first_name"`
	LastName  int `json:"last_name"`
	Email     float64 `json:"email"`
	Phone     float32 `json:"phone"`
}

func mkstr(size int) string {
	if size==0{
		log.Fatal("Error: size is 0;")
	}
	str := make([]byte, size)
	for i := range str {
		rand.Seed(time.Now().UnixNano())
		str[i] = letters[rand.Intn(len(letters))]
	}
	if size == 0 {
		return mkstr(size)
	}
	return string(str)
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

	for _,c := range customers {
		bytearr,err:=json.Marshal(&c)
		if err!=nil{
			fmt.Println(err)
		}
		testStat(bytearr,http.StatusCreated,http.MethodPost,"/customers/")
	}
	c:=WrongCust{FirstName:145,LastName:-10,Email:float64(14),Phone:float32(1251)}
	bytearr,err:=json.Marshal(&c)
	if err!=nil{
		fmt.Println(err)
	}
	testStat(bytearr,http.StatusBadRequest,http.MethodPost,"/customers/")
}
func handler() http.Handler {
	r:=mux.NewRouter()
	s := r.PathPrefix("/customers").Subrouter()
	s.HandleFunc("/", crud.CreateCustomer).Methods(http.MethodPost)
	s.HandleFunc("", crud.ListOfCustomers).Methods(http.MethodGet)
	s.HandleFunc("/{id}", crud.RetrieveCustomer).Methods(http.MethodGet)

	s.HandleFunc("/{id}", crud.ReplaceCustomer).Methods(http.MethodPut)
	s.HandleFunc("/{id}", crud.UpdateCustomer).Methods(http.MethodPatch)
	s.HandleFunc("/{id}", crud.DeleteCustomer).Methods(http.MethodDelete)

	s.HandleFunc("/{id}/create_acc", accounts.CreateAccount).Methods(http.MethodPost)
	s.HandleFunc("/{account_id}/delete", accounts.DeleteAccount).Methods(http.MethodDelete)
	s.HandleFunc("/{account_id}/send", accounts.SendMoney).Methods(http.MethodPost)
	s.HandleFunc("/{account_id}/block", accounts.BlockAcc).Methods(http.MethodGet)
	s.HandleFunc("/{account_id}/total", accounts.GetMoney).Methods(http.MethodGet)

	s.HandleFunc("/get_by_timestamp", accounts.GetByDate).Methods(http.MethodPost)
	return s
}
func testStat(data []byte, expectStatus int, method string,url string) {
	srv:=httptest.NewServer(handler())
	defer srv.Close()
	client:=http.Client{}
	switch  {
	case method == http.MethodGet:
		_=httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, url), nil)
	case method == http.MethodPost:
		resp,err:=client.Post(fmt.Sprintf("%s%s", srv.URL, url),"application/json",bytes.NewBuffer(data))
		if err!=nil{
			log.Fatal(err)
		}
		if resp.StatusCode != expectStatus {
			fmt.Printf("ERROR: While sending post expected status %d,got %d",
				expectStatus, resp.StatusCode)
			return
		}else{
			fmt.Printf("Got status:",expectStatus)
		}

	case method==http.MethodPatch :
		monkey.Patch()
	}
}
