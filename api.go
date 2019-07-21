package main

import (
	"log"
	"net/http"

	"github.com/IvNSml/GoAPI/accounts"
	"github.com/IvNSml/GoAPI/crud"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	router := mux.NewRouter()
	init := crud.ConnectToDB()
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
	subrouter.HandleFunc("/", crud.CreateCustomer).Methods(http.MethodPost)
	subrouter.HandleFunc("", crud.ListOfCustomers).Methods(http.MethodGet)
	subrouter.HandleFunc("/{id}", crud.RetrieveCustomer).Methods(http.MethodGet)

	subrouter.HandleFunc("/{id}", crud.ReplaceCustomer).Methods(http.MethodPut)
	subrouter.HandleFunc("/{id}", crud.UpdateCustomer).Methods(http.MethodPatch)
	subrouter.HandleFunc("/{id}", crud.DeleteCustomer).Methods(http.MethodDelete)

	subrouter.HandleFunc("/{id}/create_acc", accounts.CreateAccount).Methods(http.MethodPost)
	subrouter.HandleFunc("/{account_id}/delete", accounts.DeleteAccount).Methods(http.MethodDelete)
	subrouter.HandleFunc("/{account_id}/send", accounts.SendMoney).Methods(http.MethodPost)
	subrouter.HandleFunc("/{account_id}/block", accounts.BlockAcc).Methods(http.MethodGet)
	subrouter.HandleFunc("/{account_id}/total", accounts.GetMoney).Methods(http.MethodGet)

	subrouter.HandleFunc("/get_by_timestamp", accounts.GetByDate).Methods(http.MethodPost)
	http.Handle("/", router)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
