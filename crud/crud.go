package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	CONNSTR   = "postgres://postgres:1234@localhost:5432"
	DB_DRIVER = "postgres"
)

type Customer struct {
	ID        string
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}
type Account struct {
	AccountId  string
	CustomerId string
	Total      *float64 `json:"total"`
	IsBlocked  bool     `json:"is_blocked"`
}

func GetID(r *http.Request) (string, error) {
	if !strings.Contains(r.RequestURI,"/customers/"){
		return "",fmt.Errorf("ERROR: get bad URL")
	}
	id:=r.RequestURI[len(r.RequestURI)-36:]
	return id,nil
}

func ConnectToDB() (database *sql.DB) {
	database, err := sql.Open(DB_DRIVER, CONNSTR)
	if err = database.Ping(); err != nil {
		log.Fatal(err)
		return
	}
	return database
}
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var c Customer
	byrearr, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(byrearr, &c)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid data:%v", err)))
		return
	}
	if c.FirstName == "" || c.LastName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("First name and last name are required"))
		return
	}
	c.ID = uuid.New().String()
	rows, err := db.Exec("INSERT INTO customers"+
		"(id,first_name,last_name,email,phone)"+
		"VALUES ("+
		"$1,$2,$3,$4,$5);", &c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
	if err!=nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		return
	}
	ra, _ := rows.RowsAffected()
	if ra != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(c.ID))
	fmt.Printf("CREATED USER with id %s", c.ID)
	fmt.Println()
	return
}
func RetrieveCustomer(w http.ResponseWriter, r *http.Request) { //we wont use regular expression because of uuid
	db := ConnectToDB()
	defer db.Close()
	id,err:=GetID(r)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintln(err)))
		return
	}
	var c Customer
	err = db.QueryRow("SELECT * FROM customers WHERE id = $1;", id).Scan(&c.ID,
		&c.FirstName, &c.LastName, &c.Email, &c.Phone)
	switch {
	case err == sql.ErrNoRows:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No customer with id %s", id)
		return
	case err != nil:
		fmt.Println(err)
		return
	default:
		bytearr, err := json.MarshalIndent(&c, "", " ")
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Write(bytearr)
		return
	}
}

func ListOfCustomers(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	var customers []Customer
	rows, err := db.Query("SELECT * FROM customers;")
	defer rows.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error was occured during requesting data"))
		fmt.Println(err)
		return
	}
	for rows.Next() {
		var c Customer
		err = rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
		customers = append(customers, c)
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
	var c Customer
	bytearr, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytearr, &c)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Inсorrect json data"))
		fmt.Println(err)
	}
	id := mux.Vars(r)["id"]
	_, err = db.Exec("UPDATE customers SET "+
		"first_name=$1,last_name=$2,email=$3,phone=$4 WHERE id=$5;", c.FirstName,
		c.LastName, c.Email, c.Phone, id)
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusNoContent)
	fmt.Printf("REPLACED CUSTOMER WITH ID %s", id)
	fmt.Println()
}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	jsonDict := make(map[string]interface{})
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
		result, err := db.Exec(fmt.Sprintf("UPDATE customers SET %s=$1 WHERE id=$2;", key),
			val, id)
		rowsAffected, _ := result.RowsAffected()
		if err != nil || rowsAffected != 1 || result == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid data"))
			fmt.Println(err)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Printf("UPDATED USER with id %s", id)
	fmt.Println()
}
func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	result, err := db.Exec("DELETE FROM customers WHERE id=$1", id)
	rowsAffected, _ := result.RowsAffected()
	switch {
	case err != nil:
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
		return
	case rowsAffected != 1:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		return
	default:
		w.WriteHeader(http.StatusOK)
		fmt.Printf("DELETED CUSTOMER %s", id)
		fmt.Println()
	}
}
