package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"final/accounts"
	"final/crud"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)
var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const WILLBEEXEC = 50

type WrongCust struct {
	ID        int
	FirstName int     `json:"first_name"`
	LastName  int     `json:"last_name"`
	Email     float64 `json:"email"`
	Phone     float32 `json:"phone"`
}

func handler() http.Handler {
	r := mux.NewRouter()
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

func mkstr(size int) string {
	size++
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
		customers = append(customers, crud.Customer{
			FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)),
			Email: mkstr(rand.Intn(20)),
			Phone: mkstr(rand.Intn(15)),
		})

		customers = append(customers, crud.Customer{
			FirstName: mkstr(rand.Intn(20)),
			LastName: mkstr(rand.Intn(20)),
		})
	}
	for _, c := range customers {
		bytearr, err := json.Marshal(&c)
		if err != nil {
			t.Error(err)
		}
		err = testFuncs(bytearr, http.StatusCreated, http.MethodPost, "/customers/",nil,nil)
		if err != nil {
			t.Error(err)
		}
		c := WrongCust{
			FirstName: rand.Intn(100),
			LastName: -10,
			Email: rand.NormFloat64(),
			Phone: float32(rand.NormFloat64()),
		}
		//other req border
		bytearr, err = json.Marshal(&c)
		if err != nil {
			t.Error(err)
		}
		err = testFuncs(bytearr, http.StatusBadRequest, http.MethodPost, "/customers/",
			nil,nil)
		if err != nil {
			t.Error(err)
		}
	}
}
func TestRetrieveCustomer(t *testing.T) {
	db := crud.ConnectToDB()
	defer db.Close()
	var (
		c   crud.Customer
		arr []crud.Customer
	)
	rows, err := db.Query("SELECT * FROM customers;")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
		arr = append(arr, c)
	}
	if err != rows.Err() {
		log.Fatal(err)
	}
	for _, c := range arr {
		err = testFuncs(nil, http.StatusOK, http.MethodGet,
			fmt.Sprintf("/customers/%s", c.ID),db,nil)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestUpdateCustomer(t *testing.T) {
	db:=crud.ConnectToDB()
	rows,err:=db.Query("SELECT id FROM customers;")
	if err!=nil{
		t.Error(err)
	}
	var ids []string
	for rows.Next(){
		if err!=rows.Scan(&ids){
			t.Error(err)
		}
	}
	defer rows.Close()
	var customers []crud.Customer
	for i, _:=range customers{
		switch {
		case i%5==0:
			c:=crud.Customer{
				FirstName:mkstr(rand.Intn(20)),
			}
			bytearr,err:=json.Marshal(&c)
			if err!=nil{
				t.Error(err)
			}
			err=testFuncs(bytearr,http.StatusOK,http.MethodPatch,fmt.Sprintf("/customers/%s",ids[i]),db,c)
			if err != nil {
				t.Error(err)
			}
		case i%5==1:
			c:=crud.Customer{
				LastName:mkstr(rand.Intn(20)),
			}
			bytearr,err:=json.Marshal(&c)
			if err!=nil{
				t.Error(err)
			}
			err=testFuncs(bytearr,http.StatusOK,http.MethodPatch,fmt.Sprintf("/customers/%s",ids[i]),db,c)
			if err != nil {
				t.Error(err)
			}
		case i%5==2:
			c:=crud.Customer{
				Email:mkstr(rand.Intn(20)),
			}
			bytearr,err:=json.Marshal(&c)
			if err!=nil{
				t.Error(err)
			}
			err=testFuncs(bytearr,http.StatusOK,http.MethodPatch,fmt.Sprintf("/customers/%s",ids[i]),db,c)
			if err != nil {
				t.Error(err)
			}
		case i%5==3:
			c:=crud.Customer{
				Phone:mkstr(rand.Intn(20)),
			}
			bytearr,err:=json.Marshal(&c)
			if err!=nil{
				t.Error(err)
			}
			err=testFuncs(bytearr,http.StatusOK,http.MethodPatch,fmt.Sprintf("/customers/%s",ids[i]),db,c)
			if err != nil {
				t.Error(err)
			}
		case i%5==4:
			c:=crud.Customer{
				ID:mkstr(rand.Intn(20)),
			}
			bytearr,err:=json.Marshal(&c)
			if err!=nil{
				t.Error(err)
			}
			err=testFuncs(bytearr,http.StatusOK,http.MethodPatch,fmt.Sprintf("/customers/%s",ids[i]),db,c)
			if err != nil {
				t.Error(err)
			}
		}
	}
	empty:=crud.Customer{
		ID:        "123",
		FirstName: "Ivan",
		LastName:  "",
	}
	bytearrEMPTY,err:=json.Marshal(&empty)
	if err!=nil{
		t.Error(err)
	}
	wrong:=WrongCust{
		ID:        123,
		FirstName: -321,
		LastName:  15,
		Email: float64(65),
		Phone: float32(12),
	}
	bytearrWRONG,err:=json.Marshal(&wrong)
	err=testFuncs(bytearrWRONG,http.StatusBadRequest,http.MethodPatch,fmt.Sprintf("/customers/%s"),db,wrong)
	if err!=nil{
		t.Error(err)
	}
	err=testFuncs(bytearrEMPTY,http.StatusBadRequest,http.MethodPatch,"/customers/%s",db,empty)
	if err!=nil{
		t.Error(err)
	}
	for _,c:= range customers {
		bytearr, err := json.Marshal(&c)
		if err != nil {
			t.Error(err)
		}
		err = testFuncs(bytearr, http.StatusCreated, http.MethodPost, "/customers/", nil,nil)
		if err != nil {
			t.Error(err)
		}
	}
}
//var structure only for patch and put
func testFuncs(data []byte, expectStatus int, method string, url string,db *sql.DB, s interface{}) error {
	srv := httptest.NewServer(handler())
	defer srv.Close()
	client := &http.Client{}
	switch {
	case method == http.MethodGet:
		r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, url), nil)
		r = mux.SetURLVars(r, map[string]string{"id": url[len(url)-36:]})
		res, err := client.Do(r)
		defer res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		var c crud.Customer
		c.ID=mux.Vars(r)["id"]
		err = db.QueryRow("SELECT * FROM customers WHERE id=$1;",c.ID).Scan(
			&c.ID,&c.FirstName,&c.LastName,&c.Email,&c.Phone)
		switch {
		case err == sql.ErrNoRows:
			err=fmt.Errorf("ERROR: Wrong id")
			return err
		case err != nil:
			return err
		}
		//
		var body crud.Customer
		b, err := ioutil.ReadAll(res.Body)
		if err!=nil{
			return err
		}
		err=json.Unmarshal(b,&body)
		if err!=nil{
			return err
		}

		if !reflect.DeepEqual(body,c){
			return fmt.Errorf("ERROR: Wrong data recieved")

		}
		if res.StatusCode != expectStatus {
			err = fmt.Errorf("ERROR: While sending get expected status %d,got %d\n",
				expectStatus, res.StatusCode)
			return err
		} else {
			fmt.Printf("ERROR: Data are correct; Got status:%d\n", expectStatus)
			return nil
		}

		case method == http.MethodPost:
		resp, err := client.Post(fmt.Sprintf("%s%s", srv.URL, url), "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != expectStatus {
			return fmt.Errorf("ERROR: While sending post expected status %d,got %d\n",
				expectStatus, resp.StatusCode)
		} else {
			fmt.Printf("Got status:%d\n", expectStatus)
			return nil
		}

	case method == http.MethodPatch:
		r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("%s%s", srv.URL, url), bytes.NewBuffer(data))
		r.Header.Set("Content-type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": url[len(url)-36:]})
		resp, err := client.Do(r)
		if err != nil {
			fmt.Println(err)
		}
		///////
		var c crud.Customer
		c.ID=mux.Vars(r)["id"]
		err=db.QueryRow("SELECT * FROM customers WHERE id=$1;",c.ID).Scan(
			&c.ID,&c.FirstName,&c.LastName,&c.Phone,&c.Email)
		switch {
			case err == sql.ErrNoRows:
				return fmt.Errorf("ERROR: Wrong id")
			case err != nil:
				return err
		}
		if !strings.Contains(string(s),){
			return fmt.Errorf("ERROR: Wrong data recieved")
		}
		if resp.StatusCode != expectStatus {
			 return fmt.Errorf("ERROR: While sending patch expected status %d,got %d\n",
				expectStatus, resp.StatusCode)

		} else {
			fmt.Printf("Got status:%d\n", expectStatus)
		}
	}
	return fmt.Errorf("error")
}
