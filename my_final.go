package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
)

var DB_DRIVER string = "postgres"
var DBUSERNAME string = "postgres"
var PASSWORD string = "1234"
var DB_PORT string = "5432"
var DB_NAME string = "postgres" //Postgres schema

type Customer struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

func main() {
	router := mux.NewRouter()
	init := ConnectToDB()
	_, err := init.Query("DROP TABLE IF EXISTS customers;" +
		"CREATE TABLE customers(" +
		"id VARCHAR(50) NOT NULL," +
		"first_name VARCHAR(50) NOT NULL," +
		"last_name VARCHAR(50) NOT NULL," +
		"email VARCHAR(50) NOT NULL," +
		"phone VARCHAR(10) NOT NULL);")
	if err != nil {
		log.Fatal(err)
	}
	init.Close()
	subrouter := router.PathPrefix("/customers").Subrouter()

	subrouter.HandleFunc("/", CreateCustomer).Methods(http.MethodPost)
	subrouter.HandleFunc("/", ListOfCustomers).Methods(http.MethodGet)
	subrouter.HandleFunc("/{id}", RetrieveCustomer).Methods(http.MethodGet)

	/*subrouter.HandleFunc("/{id}", ReplaceCustomer).Methods(http.MethodPut)
	subrouter.HandleFunc("/{id}", UpdateCustomer).Methods(http.MethodPatch)
	subrouter.HandleFunc("/{id}", DeleteAccount).Methods(http.MethodDelete)*/
	http.Handle("/", router)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
func ConnectToDB() (database *sql.DB) {
	database, _ = sql.Open(DB_DRIVER, DB_NAME+"://"+DBUSERNAME+":"+PASSWORD+"@localhost:"+DB_PORT)
	if err := database.Ping(); err != nil {
		log.Fatal(err)
		return
	}
	return
}
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	var customer Customer
	bytearr, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(bytearr, &customer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	currentId := uuid.New().String()
	_, err = db.Query("INSERT INTO customers"+
		"(id,first_name,last_name,email,phone)"+
		"VALUES ("+
		"$1,$2,$3,$4,$5);", currentId, customer.FirstName, customer.LastName, customer.Email, customer.Phone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Printf("CREATED USER with id (%s)", currentId)
	fmt.Println()
	return
}

func RetrieveCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	id := mux.Vars(r)["id"]
	rows, err := db.Query("SELECT * FROM customers WHERE id = $1;", id)

	if err != nil {
		fmt.Println(err)
	}
	var current Customer

	for rows.Next() {
		if err = rows.Scan(&current.ID, &current.FirstName, &current.LastName, &current.Email, &current.Phone); err != nil {
			fmt.Println(err)
		}
	}
	if current.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No such customer"))
		return
	}
	bytearr, err := json.MarshalIndent(&current, "", " ")
	w.Write(bytearr)
}
func ListOfCustomers(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	var (
		current   Customer
		customers []Customer
	)
	rows, err := db.Query("SELECT * FROM customers")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error was occured during requesting data"))
		fmt.Println(err)
		return
	}
	for rows.Next() {
		rows.Scan(&current.ID, &current.FirstName, &current.LastName, &current.Email, &current.Phone)
		bytearr, err := json.MarshalIndent(&current, "", " ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}
		customers = append(customers, current)
		w.Write(bytearr)
	}
	if len(customers) == 0 {
		w.Write([]byte("No customers"))
	}
}
