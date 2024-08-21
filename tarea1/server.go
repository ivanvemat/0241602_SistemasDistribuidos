package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Name string `json:"name"`
	Age uint8 `json:"age"`
}

var users []User

func addUser(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	if user.Name == "" || user.Username == "" || user.Age == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "User data missing"}`))
		return
	}

	users = append(users, user)	
	w.WriteHeader(http.StatusOK)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	for _, user := range users {
		if user.Username == username {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "User not found"}`))
}

func main() {
	router := mux.NewRouter()

	fmt.Println("Running server...")
	router.HandleFunc("/add-user", addUser).Methods("POST")
	router.HandleFunc("/get-user/{username}", getUser).Methods("GET")
	http.ListenAndServe(":8080", router)
}