package main 

import (
	"encoding/json" // Encode/decode JSON
	"log"           // Logging package for printing to console
	"net/http"      // Provides HTTP client 
	"strconv"       // Converts strings to other types
)

// Define a struct to represent user object 
type User struct {
	ID int `json:"id"`
	Name string `json:"name"` // maps name field to "name" in JSON
	Email string `json:"email"` //maps Email field to "email" in JSON
}

// slice  of predefined User objects 
var users = []User{
	{ID: 1, Name: "Celeste", Email: "celeste@gmail.com"},
	{ID: 2, Name: "AJ", Email: "aj@gmail.com"},
}

// get UsersHandler is HTTP handler function for GET requests to "/users"
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header to say we're returning JSON
	w.Header().Set("Content-Type", "application/json")
	// Encode the 'users' slice to JSON and write it to the response
	json.NewEncoder(w).Encode(users)
}

// getUserHandler is HTTP handler function for GET requests to "/users/{id}"
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// tell client we're sending JSON
	w.Header().Set("Content_Type", "application/json")
	// Extract user ID from URL path by trimming prefix "/users/"
	idStr := r.URL.Path[len("/users/"):]
	// Convert string (idStr) to an integer
	id, err := strconv.Atoi(idStr)
	// if conversion fails respond with 400 Bad Request 
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Search through users slice to find a user with a matching ID 
	for _, user := range users {
		if user.ID == id {
			// If found encode that user as JSON and return in the response
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	// user not found, respond with 404 Not Found 
	http.NotFound(w, r)

}

func main () {
	// Register HTTP handler for route "/users" to list all users
	http.HandleFunc("/users", getUsersHandler)
	// Register HTTP handler for any route starting with "/users/"
	// for individual user lookups like "/users/1"
	http.HandleFunc("/users/", getUserHandler)
	//log message that server starting on port 8083
	log.Println("User service listening on port 8083")
	// Start HTTP server on port 8083; 
	// log.fatal makes sure to log it if it crashes 
	log.Fatal(http.ListenAndServe(":8083", nil))
}