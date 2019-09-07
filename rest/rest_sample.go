package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"./utils"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	// gorm.Model
	Id        int
	FirstName string
	LastName  string
}

var db *gorm.DB

func init() {
	db = utils.GetConnection()
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is the home page. Welcome!")
}

func returnAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	db.Table("users").Find(&users)
	respondWithJson(w, http.StatusOK, users)
}

func returnSingleUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var user User
	db.Table("users").Where("id = ?", key).First(&user)

	if user.Id == 0 {
		respondWithError(w, http.StatusNotFound, "Item does not exist")
		return
	}
	respondWithJson(w, http.StatusOK, user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := json.Unmarshal(body, &user); err != nil {
		log.Println(body)
		log.Println(err)
		respondWithError(w, http.StatusBadRequest, "JSON failed")
		return
	}
	db.Table("users").Create(&user)
	respondWithJson(w, http.StatusOK, user)

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/users", returnAllUsers).Methods("GET")
	myRouter.HandleFunc("/users/{id}", returnSingleUser).Methods("GET")
	myRouter.HandleFunc("/users", createUser).Methods("POST")
	log.Fatal(http.ListenAndServe(":55555", myRouter))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	defer db.Close()
	handleRequests()
}
