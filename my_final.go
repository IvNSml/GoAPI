package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	DB_DRIVER  = "postgres"
	DBUSERNAME = "postgres"
	PASSWORD   = "1234"
	DbPort     = "5432"
	DbName     = "postgres"
) //Postgres schema

type Customer struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}
type Account struct {
	AccountId          string      `json:"account_id"`
	CustomerId         string      `json:"customer_id"`
	Total              *float64    `json:"total"`
	TimeOfTransactions []uuid.Time `json:"time_of_transactions"`
	IsBlocked          bool        `json:"is_blocked"`
}
type SendMoneyForm struct {
	Amount             float64 `json:"amount"`
	Receiver           string  `json:"receiver"`
	ReceiversAccNumber string  `json:"receivers_acc_number"`
}

func main() {
	router := mux.NewRouter()
	init := ConnectToDB()
	defer init.Close() //DROP TABLE FROM PREVIOUS SERVERS LAUNCH
	_, err := init.Exec("CREATE TABLE IF NOT EXISTS customers(" +
		"id VARCHAR(50) NOT NULL PRIMARY KEY," +
		"first_name VARCHAR(20) NOT NULL," +
		"last_name VARCHAR(20) NOT NULL," +
		"email VARCHAR(20)," +
		"phone VARCHAR(15));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = init.Exec("CREATE TABLE IF NOT EXISTS accounts(" +
		"customer_id VARCHAR(50) NOT NULL," +
		"account_id VARCHAR(50) NOT NULL PRIMARY KEY," +
		"total NUMERIC," +
		"time_of_transactions TIMESTAMP WITH TIME ZONE[]," +
		"is_blocked BOOLEAN DEFAULT FALSE," +
		"FOREIGN KEY (customer_id) REFERENCES customers(id));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = init.Exec("CREATE TABLE IF NOT EXISTS transactions ;")
	subrouter := router.PathPrefix("/customers").Subrouter()

	subrouter.HandleFunc("/", CreateCustomer).Methods(http.MethodPost)
	subrouter.HandleFunc("/", ListOfCustomers).Methods(http.MethodGet)
	subrouter.HandleFunc("/{id}", RetrieveCustomer).Methods(http.MethodGet)

	subrouter.HandleFunc("/{id}", ReplaceCustomer).Methods(http.MethodPut)
	subrouter.HandleFunc("/{id}", UpdateCustomer).Methods(http.MethodPatch)
	subrouter.HandleFunc("/{id}", DeleteCustomer).Methods(http.MethodDelete)

	subrouter.HandleFunc("/{id}/", CreateAccount).Methods(http.MethodPost)
	subrouter.HandleFunc("/{id}/{account_id}", DeleteAccount).Methods(http.MethodDelete)
	subrouter.HandleFunc("/{id}/{account_id}/send", SendMoney).Methods(http.MethodPost)
	http.Handle("/", router)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
func ConnectToDB() (database *sql.DB) {
	database, _ = sql.Open(DB_DRIVER, DbName+"://"+DBUSERNAME+":"+PASSWORD+"@localhost:"+DbPort)
	if err := database.Ping(); err != nil {
		log.Fatal(err)
		return
	}
	return
}

func CheckIfExists(id string) bool {
	db := ConnectToDB()
	var b string
	err := db.QueryRow("SELECT * FROM customers WHERE id=$1;", id).Scan(&b)
	if err == sql.ErrNoRows {
		return false
	}
}

func CreateCustomer(w http.ResponseWriter, r *http.Request) { //here I leave as it, because
	// Handlefunc accepts only two args
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var customer Customer
	bytearr, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(bytearr, &customer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid data:%v", err)))
		return
	}
	customer.ID = uuid.New().String()
	_, err = db.Query("INSERT INTO customers"+
		"(id,first_name,last_name,email,phone)"+
		"VALUES ("+
		"$1,$2,$3,$4,$5);", customer.ID, customer.FirstName, customer.LastName, customer.Email, customer.Phone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(customer.ID))
	fmt.Printf("CREATED USER with id (%s)", customer.ID)
	fmt.Println()
	return
}

func RetrieveCustomer(w http.ResponseWriter, r *http.Request) { //regexp-это регулярное выражение;
	db := ConnectToDB()
	defer db.Close()
	id := mux.Vars(r)["id"]
	var c Customer
	err := db.QueryRow("SELECT * FROM customers WHERE id = $1;", id).Scan(&c)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No such customer"))
			return
		}
	}
	bytearr, err := json.MarshalIndent(&c, "", " ")
	w.Write(bytearr)
}
func ListOfCustomers(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB() //paging/пагинация
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
	fmt.Printf("REPLACED USER with id (%s)", id)
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
			"%s=$1 WHERE id=$2;", key), val, id)
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

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var account Account
	account.CustomerId = mux.Vars(r)["id"]
	account.AccountId = uuid.New().String()
	account.Total = nil
	bytearr, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(bytearr, &account); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	if account.Total == nil { //how to check if the json data is valid?!
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	_, err = db.Exec("INSERT INTO accounts"+
		"(total,customer_id,account_id)"+
		"VALUES ("+
		"$1,$2,$3);", account.Total, account.CustomerId, account.AccountId) //there is a foreign key,so it
	// doesn't have a sense to check if explicitly
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Wrong customer"))
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(account.AccountId))
	fmt.Printf("CREATED ACCOUNT WITH ID (%s) TO USER WITH ID(%s)", account.AccountId, account.CustomerId)
	fmt.Println()
	return
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	customer_id := mux.Vars(r)["id"]
	account_id := mux.Vars(r)["account_id"]
	result, err := db.Exec("DELETE FROM accounts WHERE customer_id=$1 AND account_id=$2", customer_id, account_id)
	rowsAffected, _ := result.RowsAffected()
	if err != nil || rowsAffected != 1 || result == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Printf("DELETED ACCOUNT (%s) OWNED BY USER (%s)", account_id, customer_id)
	fmt.Println()
}

func SendMoney(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	tx, err := db.Begin()
	if tx != nil {
		fmt.Println(err)
		return
	}
	var sendForm SendMoneyForm
	customer_id := mux.Vars(r)["id"]
	account_id := mux.Vars(r)["account_id"]
	bytearr, err := ioutil.ReadAll(r.Body)
	var receiversMoney, sendersMoney float64
	if err = json.Unmarshal(bytearr, &sendForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	rows, err := db.Query("SELECT total FROM accounts WHERE account_id=$1 AND customer_id=$2;", account_id, customer_id)
	if err != nil {
		fmt.Println(err)
		return
	}
	rows.Scan(&sendersMoney, &receiversMoney)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !rows.NextResultSet() {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rows, err = db.Query("SELECT total FROM accounts WHERE account_id=$1 AND customer_id=$2;", sendForm.Receiver, sendForm.ReceiversAccNumber)
	if err != nil {
		fmt.Println(err)
		return
	}
	rows.Scan(&sendersMoney, &receiversMoney)
	if err != nil {
		fmt.Println(err)
		return
	}
	sendersMoney -= sendForm.Amount
	receiversMoney += sendForm.Amount
	res, err := db.Exec("UPDATE accounts SET total=$1 WHERE customer_id=$2 AND account_id=$3;"+
		"UPDATE accounts SET total=$4, WHERE customer_id=$5 AND account_id=$6", sendersMoney, account_id, customer_id, receiversMoney, sendForm.Receiver, sendForm.ReceiversAccNumber)
	rowsAffected, _ := res.RowsAffected()
	if err != nil || rowsAffected != 1 || res == nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err = tx.Rollback(); err != nil {
			fmt.Println(err)
			return
		}
	} //time of committing of any operation or what?
	res, err = db.Exec("UPDATE accounts SET time_of_transactions=array_append(time_of_transactions,now()) WHERE account_id='$1';", account_id)
	if err != nil {
		fmt.Println(err)
		if err = tx.Rollback(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Printf("SENT (%f) FROM USER (%s) TO USER (%s)", sendForm.Amount, account_id, sendForm.Receiver)
	fmt.Println()
}
