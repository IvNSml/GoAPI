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
	IsBlocked  bool    `json:"is_blocked"`
}
type SendMoneyForm struct {
	SendersAccNumber   string  `json:"senders_acc_number"`
	ReceiversAccNumber string  `json:"receivers_acc_number"`
	Amount             float64 `json:"amount"`
}

func main() {
	router := mux.NewRouter()
	init := ConnectToDB()
	defer init.Close() //RESET TABLES AFTER RESETTING SERVER
	_, err := init.Exec("DROP TABLE IF EXISTS customers CASCADE;" +
		"CREATE TABLE customers(" +
		"id VARCHAR(50) NOT NULL PRIMARY KEY," +
		"first_name VARCHAR(20) NOT NULL," +
		"last_name VARCHAR(20) NOT NULL," +
		"email VARCHAR(20)," +
		"phone VARCHAR(15));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = init.Exec("DROP TABLE IF EXISTS accounts CASCADE;" +
		"CREATE TABLE accounts(" +
		"customer_id VARCHAR(50) NOT NULL," +
		"account_id VARCHAR(50) NOT NULL PRIMARY KEY," +
		"total NUMERIC CHECK(total>=0) NOT NULL," +
		"is_blocked BOOLEAN DEFAULT FALSE," +
		"FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE);")
	if err != nil {
		log.Fatal(err)
	}
	_, err = init.Exec("DROP TABLE IF EXISTS transactions;" +
		"CREATE TABLE transactions(" +
		"senders_account_id VARCHAR(50) NOT NULL," +
		"receivers_account_id VARCHAR(50) NOT NULL," +
		"amount NUMERIC CHECK(amount>=0) NOT NULL," +
		"time_of_transaction TIMESTAMP WITH TIME ZONE," +
		"FOREIGN KEY (senders_account_id) REFERENCES accounts(account_id)," +
		"FOREIGN KEY (receivers_account_id) REFERENCES accounts(account_id));")
	if err != nil {
		log.Fatal(err)
	}
	//We could connect to database only once in main function and pass sql.DB object in each
	// function,but golang doesnt accepts overrides
	subrouter := router.PathPrefix("/customers").Subrouter()
	subrouter.HandleFunc("/", CreateCustomer).Methods(http.MethodPost)
	subrouter.HandleFunc("", ListOfCustomers).Methods(http.MethodGet)
	subrouter.HandleFunc("/{id}", RetrieveCustomer).Methods(http.MethodGet)

	subrouter.HandleFunc("/{id}", ReplaceCustomer).Methods(http.MethodPut)
	subrouter.HandleFunc("/{id}", UpdateCustomer).Methods(http.MethodPatch)
	subrouter.HandleFunc("/{id}", DeleteCustomer).Methods(http.MethodDelete)

	subrouter.HandleFunc("/{id}/create_acc", CreateAccount).Methods(http.MethodPost)
	subrouter.HandleFunc("/{account_id}/delete", DeleteAccount).Methods(http.MethodDelete)
	subrouter.HandleFunc("/{account_id}/send", SendMoney).Methods(http.MethodPost)
	subrouter.HandleFunc("/{account_id}/block", BlockAcc).Methods(http.MethodGet)
	subrouter.HandleFunc("/{account_id}/total", GetMoney).Methods(http.MethodGet)
	http.Handle("/", router)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
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
	_, err = db.Query("INSERT INTO customers"+
		"(id,first_name,last_name,email,phone)"+
		"VALUES ("+
		"$1,$2,$3,$4,$5);", &c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
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
	id := mux.Vars(r)["id"]
	var c Customer
	err := db.QueryRow("SELECT * FROM customers WHERE id = $1;", id).Scan(&c.ID,
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

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var account Account
	account.CustomerId = mux.Vars(r)["id"]
	account.AccountId = uuid.New().String()
	account.Total = nil //evaluate with nil
	bytearr, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(bytearr, &account); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	if account.Total == nil || *account.Total < 0 { //check if Total(required field) is still nil
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		return
	}
	_, err = db.Exec("INSERT INTO accounts"+
		"(total,customer_id,account_id)"+
		"VALUES ("+
		"$1,$2,$3);", account.Total, account.CustomerId, account.AccountId) //there is a foreign key,so it
	// doesn't have a sense to check if exist explicitly
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Wrong customer"))
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(account.AccountId))
	fmt.Printf("CREATED ACCOUNT WITH ID %s TO USER WITH ID %s", account.AccountId, account.CustomerId)
	fmt.Println()
	return
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	account_id := mux.Vars(r)["account_id"]
	result, err := db.Exec("DELETE FROM accounts WHERE account_id=$1", account_id)
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
		fmt.Printf("DELETED ACCOUNT %s", account_id)
		fmt.Println()
	}
}
func CheckBlocked (acc_id string)  (bool, error) {//Check if blocked and check if account id is valid
	db:=ConnectToDB()
	var r bool
	err := db.QueryRow("SELECT is_blocked FROM accounts WHERE account_id=$1", acc_id).Scan(&r)
	if err!=nil || err==sql.ErrNoRows{
		return true,err
	}
	return r,nil
}
func SendMoney(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	var sendForm SendMoneyForm
	bytearr, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(bytearr, &sendForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid data"))
		fmt.Println(err)
		return
	}
	rec,err:=CheckBlocked(sendForm.ReceiversAccNumber)
	if err!=nil{
		fmt.Println(err)
		return
	}
	send,err:=CheckBlocked(sendForm.ReceiversAccNumber)
	if err!=nil{
		fmt.Println(err)
		return
	}
	if send{
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w,"Account %s is blocked or invalid data", sendForm.SendersAccNumber)
		return
	}
	if rec{
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w,"Account %s is blocked or invalid data", sendForm.ReceiversAccNumber)
		return
	}
	if (SendMoneyForm{}) == sendForm {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("All fields are obligatory to fill"))
		return
	}
	rows, err := db.Exec(fmt.Sprintf("UPDATE accounts SET total=total-%f WHERE account_id=$1;", sendForm.Amount),
		sendForm.SendersAccNumber)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid data:", err)
		return
	}
	rows, err = db.Exec(fmt.Sprintf("UPDATE accounts SET total=total+%f WHERE account_id=$1;", sendForm.Amount),
		sendForm.ReceiversAccNumber)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid data:", err)
		return
	}
	rows, err = db.Exec("INSERT INTO transactions (senders_account_id,receivers_account_id,"+
		"amount, time_of_transaction) VALUES ($1,$2,$3,now());", sendForm.SendersAccNumber,
		sendForm.ReceiversAccNumber, sendForm.Amount)
	aff, _ := rows.RowsAffected()
	if aff != 1 {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Transaction denied:", err)
		tx.Rollback()
		return
	}
	tx.Commit()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transaction approved;"))
	fmt.Printf("SENT %f FROM ACCOUNT %s TO ACCOUNT %s", sendForm.Amount, sendForm.SendersAccNumber,
		sendForm.ReceiversAccNumber)
	fmt.Println()
}

func BlockAcc(w http.ResponseWriter, r *http.Request) {
	db := ConnectToDB()
	defer db.Close()
	acc_id := mux.Vars(r)["account_id"]
	b,err:=CheckBlocked(acc_id)
	if err!=nil{
		fmt.Println(err)
		return
	}
	if b{
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w,"Account %s is already blocked or invalid data",acc_id)
		return
	}
	_, err = db.Exec("UPDATE accounts SET is_blocked=TRUE WHERE account_id=$1;", acc_id)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"Account %s is blocked",acc_id)
	fmt.Printf("Account %s is blocked",acc_id)
}
func GetMoney(w http.ResponseWriter, r *http.Request)  {
	db:=ConnectToDB()
	acc_id:=mux.Vars(r)["account_id"]
	defer db.Close()
	b,err:=CheckBlocked(acc_id)
	if err!=nil{
		fmt.Println(err)
		return
	}
	if b{
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w,"Account %s is blocked or invalid data",acc_id)
		return
	}
	var total float64
	err = db.QueryRow("SELECT total FROM accounts WHERE account_id=$1", acc_id).Scan(&total)
	if err!=nil{
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"Total on account: %f", total)
}
func SendNotification(account_id string){

}