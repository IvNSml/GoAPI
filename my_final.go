package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_"github.com/lib/pq"
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

	subrouter.HandleFunc("/{id}", ReplaceCustomer).Methods(http.MethodPut)
	subrouter.HandleFunc("/{id}", UpdateCustomer).Methods(http.MethodPatch)
	subrouter.HandleFunc("/{id}", DeleteCustomer).Methods(http.MethodDelete)
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
	defer r.Body.Close()
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
		err = rows.Scan(&current.ID, &current.FirstName, &current.LastName, &current.Email, &current.Phone)
		customers = append(customers, current)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}
	}
	if len(customers) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No customers"))
		return
	}

	bytearr, err := json.MarshalIndent(&customers, "", " ")
	w.Write(bytearr)
}

func ReplaceCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var current Customer
	bytearr, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytearr, &current)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Inсorrect json data"))
		fmt.Println(err)
	}
	id := mux.Vars(r)["id"]
	_, err = db.Query("UPDATE customers SET "+
		"first_name=$1,last_name=$2,email=$3,phone=$4 WHERE id=$5;", current.FirstName,
		current.LastName, current.Email, current.Phone, id)
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusNoContent)
	fmt.Printf("UPDATED USER with id (%s)", id)
	fmt.Println()
}
func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	jsonDict := make(map[string]string)
	bytearr, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytearr, &jsonDict)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Inсorrect json data"))
		fmt.Println(err)
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	for key, val := range jsonDict {
		result, err := db.Exec(fmt.Sprintf("UPDATE customers SET "+
			"%s=$1 WHERE id=$2;", key), val, id) //Don't know how to do this other way
		rowsAffected, _ := result.RowsAffected()
		if err != nil || rowsAffected != 1 || result == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cant process your request"))
			fmt.Println(err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	fmt.Printf("UPDATED USER with id (%s)", id)
	fmt.Println()
}
func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	result, err := db.Exec("DELETE FROM customers WHERE id=$1", id)
	rowsAffected, _ := result.RowsAffected()
	if err != nil || rowsAffected != 1 || result == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Printf("DELETED USER with id (%s)", id)
	fmt.Println()
}
func () {

}
