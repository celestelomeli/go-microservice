package main

// Import packages
import (
	"encoding/json" // For encoding structs to JSON and decoding JSON to structs
	"fmt"           // For string formatting 
	"log"           // For logging errors
	"net/http"      // For creating HTTP servers and routing
	"os"            // For reading environment variables
	"strconv"       // For converting string to int
	"strings"       // For working with strings 

	"gorm.io/driver/postgres" // GORM PostgreSQL driver
	"gorm.io/gorm"            // GORM ORM library
)

/////////////////////////////////////////////////////////////
// DB Model
/////////////////////////////////////////////////////////////

// User represents the users table in the database
type User struct {
	ID    int    `json:"id" gorm:"primaryKey"` // ID as the primary key
	Name  string `json:"name"`                 // Name field 
	Email string `json:"email"`                // Email field 
}

/////////////////////////////////////////////////////////////
// Global DB Connection
/////////////////////////////////////////////////////////////

var db *gorm.DB // Interacts with the database from handlers

/////////////////////////////////////////////////////////////
// Initialize Database Connection
/////////////////////////////////////////////////////////////

func initDB() {
	// Format PostgreSQL connection string (DSN) using environment variables
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),     // (postgres)
		os.Getenv("DB_USER"),     // (celeste)
		os.Getenv("DB_PASSWORD"), // (secret)
		os.Getenv("DB_NAME"),     // (microservice_db)
		os.Getenv("DB_PORT"),     // (5432)
	)

	// Open a GORM connection to the Postgres DB
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Automatically create the users table if it doesn't already exist
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to auto-migrate users table:", err)
	}

	// Count how many users exist in the DB
	var count int64
	db.Model(&User{}).Count(&count)

	// If the users table is empty, insert a default user
	if count == 0 {
		db.Create(&User{Name: "Celeste", Email: "celeste@example.com"})
	}
}

/////////////////////////////////////////////////////////////
// HTTP Handlers
/////////////////////////////////////////////////////////////

// getAllUsersHandler handles GET requests to /users
// It returns a list of all users as JSON
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User // Create a slice to store users from DB

	// Query all users from the database
	result := db.Find(&users)
	if result.Error != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Set the response header to tell the client we're sending JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the users slice into JSON and write it to the response
	json.NewEncoder(w).Encode(users)
}

// getUserHandler handles GET requests to /users/{id}
// It returns a single user by their ID
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Remove "/users/" prefix to extract just the ID
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")

	// Convert the extracted string to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User // A variable to hold the user we find

	// Look up the user by ID using GORM
	result := db.First(&user, id)
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Set content type to JSON before responding
	w.Header().Set("Content-Type", "application/json")

	// Encode the user struct into JSON and write to the response
	json.NewEncoder(w).Encode(user)
}

/////////////////////////////////////////////////////////////
// Entry Point
/////////////////////////////////////////////////////////////

func main() {
	initDB() // Initialize DB connection and auto-migrate schema

	// Register HTTP handlers
	http.HandleFunc("/users", getAllUsersHandler)   // Route to get all users
	http.HandleFunc("/users/", getUserHandler)      // Route to get a user by ID

	// Start the web server
	fmt.Println("User service running on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
