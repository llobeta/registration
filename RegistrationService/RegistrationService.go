// programs start running in package main! Other resources must be imported.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)
type Car struct{
	ID string `json:"id"`
	Brand string `json:"brand"`
	Model string `json:"model"`
	HP string `json:"horse_power"`
}
type Configuration struct{
	UserDB string `json:"UserDB"`
	PasswordDB string `json:"PasswordDB"`
	Server string `json:"Server"`
	Port string `json:"Port"`
	Database string `json:"Database"`
}
func read_config() Configuration {
	file,_ := os.Open("../config.json")
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("error:", err)
		}/* j */
	}(file)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err!= nil{
		fmt.Println("error:", err)
	}
	return configuration
}
func connect_to_database() *sql.DB{
	fmt.Println("Connecting to DB")
	config := read_config()
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.UserDB , config.PasswordDB,
		config.Server, config.Port, config.Database)
	print(dataSourceName)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil{
		panic(err.Error())
	}
	return db
}


func generate_id() string{
	rand.Seed(int64(time.Now().UnixNano()))
	return strconv.Itoa(rand.Intn(99999999))
}
func register_entry(db_driver *sql.DB, car Car){
	query := fmt.Sprintf("INSERT INTO stock VALUES ('%s', '%s', '%s', '%s')", car.ID , car.Brand, car.Model, car.HP)
	fmt.Println(query)
	insert, err := db_driver.Query(query)
	if err != nil{
		panic(err.Error())
	}
	defer insert.Close()
}

func processRequest(w http.ResponseWriter, r *http.Request) {
	var car Car
	// Create SQL driver and defer its closure until the end of main
	dbDriver := connect_to_database()
	defer func(db_driver *sql.DB) {
		err := db_driver.Close()
		if err != nil {
			log.Fatal("DB Cannot be closed")
		}
	}(dbDriver)

	defer func() {no_info := recover()
		if no_info != nil{
			w.WriteHeader(http.StatusBadRequest)
			log.Print("[REQUESTFAILED] Bad Request received.")
			return

		}}()
	/* Begin processing the request. Get the values */
	car.ID = generate_id()
	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil{
		panic("Bad Request. Body not recognized!")
	}
	if car.Brand == "" || car.Model == "" || car.HP == ""{
		panic("No valid info")
	}
	register_entry(dbDriver, car)

	/* Write answer */
	w.Header().Set("Content-Type", "application/json")
	carJson,_ := json.Marshal(car)
	/* Send and get possible errors */
	_, errWrite := w.Write(carJson)

	if errWrite != nil {
		log.Fatal("Failed Reply")

	}

	/* Logging purposess */
	log.Print(fmt.Sprintf("[REQUESTED] Car ID: %s Car Brand: %s Car Model: %s, Car HP: %s",
		car.ID, car.Brand, car.Model, car.HP))
}

func handleRequests(){
	http.HandleFunc("/", processRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func main() {handleRequests()}

