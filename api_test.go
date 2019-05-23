package main

import (
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

const WILLBEEXEC = 50

type WrongCust struct {
	ID        int
	FirstName int `json:"first_name"`
	LastName  int `json:"last_name"`
	Email     float64 `json:"email"`
	Phone     float32 `json:"phone"`
}

func handler() http.Handler {
	r:=mux.NewRouter()
	s := r.PathPrefix("/customers").Subrouter()
	s.HandleFunc("/", crud.CreateCustomer).Methods(http.MethodPost)
	s.HandleFunc("", crud.ListOfCustomers).Methods(http.MethodGet)
	//I don't know how to send a request
	//properly; mux.Vars can't catch regexp {id}. So I found this article http://mrgossett.com/post/mux-vars-problem/
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

func mkstr(size int) string {
	str := make([]byte, size)
	for i := range str {
		rand.Seed(time.Now().UnixNano())
		str[i] = letters[rand.Intn(len(letters))]
	}
	if size == 0 {
		return mkstr(10)
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
		err=testStat(bytearr,http.StatusCreated,http.MethodPost,"/customers/")
		if err!=nil{
			fmt.Println(err)
		}
		c:=WrongCust{FirstName:rand.Intn(100),LastName:-10,Email:rand.NormFloat64(),Phone:float32(rand.NormFloat64())}
		//other req border
		bytearr,err=json.Marshal(&c)
		if err!=nil{
			fmt.Println(err)
		}
		err=testStat(bytearr,http.StatusBadRequest,http.MethodPost,"/customers/")
		if err!=nil{
			fmt.Println(err)
		}
	}
}
func TestRetrieveCustomer(t *testing.T)  {
	db:=crud.ConnectToDB()
	defer db.Close()
	var (c crud.Customer
	arr []crud.Customer)
	rows,err:=db.Query("SELECT * FROM customers;")
	if err!=nil{
		log.Fatal(err)
	}
	for rows.Next(){
		rows.Scan(&c.ID,
			&c.FirstName, &c.LastName, &c.Email, &c.Phone)
		arr=append(arr,c)
	}
	if err!=rows.Err(){
		log.Fatal(err)
	}
	defer rows.Close()
	for _,c:=range arr{
		err:=testStat(nil, http.StatusOK, http.MethodGet,fmt.Sprintf("/customers/%s",c.ID))
		if err!=nil{
			fmt.Println(err)
		}
	}
}
func testStat(data []byte, expectStatus int, method string,url string) error{
	srv := httptest.NewServer(handler())
	defer srv.Close()
	client := &http.Client{}
	switch {
	case method == http.MethodGet:
		resp,err := http.Get(fmt.Sprintf("%s%s",srv.URL,url))
		if err!=nil{
			log.Fatal(err)
		}
		if resp.StatusCode != expectStatus {
			err:=fmt.Errorf("ERROR: While sending get expected status %d,got %d\n",
				expectStatus, resp.StatusCode)
			return err
		} else {
			fmt.Printf("Got status:%d\n", expectStatus)
			return nil
		}
	case method == http.MethodPost:
		resp, err := client.Post(fmt.Sprintf("%s%s", srv.URL, url), "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != expectStatus {
			err=fmt.Errorf("ERROR: While sending post expected status %d,got %d\n",
				expectStatus, resp.StatusCode)
			return err
		} else {
			fmt.Printf("Got status:%d\n", expectStatus)
			return nil
		}

	case method == http.MethodPatch:
		r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("%s%s", srv.URL, url), bytes.NewBuffer(data))
		r.Header.Set("Content-type", "application/json")
		resp, err := client.Do(r)
		if err != nil {
			fmt.Println(err)
		}
		if resp.StatusCode != expectStatus {
			err=fmt.Errorf("ERROR: While sending post expected status %d,got %d\n",
				expectStatus, resp.StatusCode)
			return err
		} else {
			fmt.Printf("Got status:%d\n", expectStatus)
			return  nil
		}
	}
	return fmt.Errorf("Error!")
}