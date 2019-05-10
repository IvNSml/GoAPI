package main

import (
	"bytes"
	_ "bytes"
	_ "database/sql"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	_ "fmt"
	_ "github.com/lib/pq"
	_ "io/ioutil"
	"log"
	_ "log"
	"net/http"
	_ "net/http"
)

type Customer struct {
	//Account number is generated authomaticly, so we can get it from GetAll function
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

var DB_DRIVER string = "postgres"
var DBUSERNAME string = "postgres"
var PASSWORD string = "1234"
var DB_PORT string = "5432"
var DB_NAME string = "postgres" //Postgres schema

type Employee struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

func main() {
	Ivan := Customer{
		FirstName: "Ivan",
		LastName:  "Smolyakov",
		Email:     "ghgchth@gmail.com",
		Phone:     "1213",
	}
	bytearr, err := json.Marshal(&Ivan)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.Post("http://localhost:8080/customers/", "application/json", bytes.NewBuffer(bytearr))
	//GET REQ

	//response, err:=http.Get("http://localhost:8080/customers/")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//bytearr,err = ioutil.ReadAll(response.Body)
	fmt.Println(response)

}

/*db, err := sql.Open(DB_DRIVER, DB_NAME+"://"+DBUSERNAME+":"+PASSWORD+"@localhost:"+DB_PORT)
if err != nil {
	log.Fatal(err)
}

rows, err := db.Query("SELECT * FROM customers WHERE email=$1", "DADWDW")
if err != nil {
	log.Fatal(err)
}
defer rows.Close()
var id,city,population,district,countrycode string
for rows.Next(){
	_=rows.Scan(&id,&city,&countrycode,&district,&population)
	fmt.Println(id,city,population,district,countrycode)
}
*/
