package accounts

import (
	"database/sql"
	"encoding/json"
	"final/crud"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	db := crud.ConnectToDB()
	defer db.Close()
	defer r.Body.Close()
	var account crud.Account
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
	db := crud.ConnectToDB()
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
	db:=crud.ConnectToDB()
	var r bool
	err := db.QueryRow("SELECT is_blocked FROM accounts WHERE account_id=$1", acc_id).Scan(&r)
	if err!=nil || err==sql.ErrNoRows{
		return true,err
	}
	return r,nil
}
func SendMoney(w http.ResponseWriter, r *http.Request) {
	db := crud.ConnectToDB()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	var sendForm crud.SendMoneyForm
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
	if (crud.SendMoneyForm{}) == sendForm {
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
	db := crud.ConnectToDB()
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
	db:=crud.ConnectToDB()
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