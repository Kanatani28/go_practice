package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"./utils"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Item struct {
	Title       string `json:"title"`
	Description string `json:"decription"`
}

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
	fmt.Println(users[0].FirstName)
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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/users", returnAllUsers)
	myRouter.HandleFunc("/users/{id}", returnSingleUser)
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
