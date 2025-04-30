package main

// Import standard libraries
import (
	"database/sql"       // For using raw SQL queries
	"encoding/json"      // For encoding/decoding JSON data
	"fmt"                // For formatted output 
	"log"                // For logging errors and messages
	"net/http"           // For building HTTP server and handling requests
	"strconv"            // For converting string to integer
	"strings"            // For manipulating strings like URL paths

	// GORM packages
	"gorm.io/driver/sqlite" // SQLite driver for GORM
	"gorm.io/gorm"          // Go ORM for interacting with DB using structs

	_ "github.com/mattn/go-sqlite3" // SQLite driver for raw SQL use (underscore = side-effect import)
)

// Define struct to represent user data
type User struct {
	ID    int    `json:"id" gorm:"primaryKey"` // Primary key (auto-incremented)
	Name  string `json:"name"`                 // Maps to name column in DB
	Email string `json:"email"`                // Maps to email column in DB
}

// Declare global DB variables for GORM and raw SQL
var (
	db     *gorm.DB // GORM database connection
	rawDB  *sql.DB  // Raw SQL database connection
)

// initDB sets up and connects to SQLite DB using GORM and raw SQL
func initDB() {
	var err error

	// Connect using GORM to SQLite file "users.db"
	db, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to GORM DB:", err)
	}

	// Auto-migrate the User struct; creates the users table if not present
	db.AutoMigrate(&User{})

	// Set up a raw SQL connection to the same DB
	rawDB, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal("Failed to connect to raw SQL DB:", err)
	}

	// Seed a user if the table is empty
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		db.Create(&User{Name: "Celeste", Email: "celeste@example.com"})
	}
}

// getAllUsersHandler handles GET requests to /users
// It uses raw SQL to fetch and return all users as JSON
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Run a raw SQL query to get all users
	rows, err := rawDB.Query("SELECT id, name, email FROM users")
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Read each row and add to a slice of users
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			http.Error(w, "Row scan failed", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// Respond with the user list as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// getUserHandler handles GET requests to /users/{id}
// uses GORM to fetch a specific user by ID
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL path
	//TrimPrefix from "strings" library
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr) // Convert string to int
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Use GORM to get the user from DB by ID
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Respond with the found user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// main sets up the server and route handlers
func main() {
	initDB() // Set up the DB connections

	// Route for getting all users (raw SQL)
	http.HandleFunc("/users", getAllUsersHandler)

	// Route for getting a user by ID (GORM)
	http.HandleFunc("/users/", getUserHandler)

	fmt.Println("User service running on port 8083...")

	// Start HTTP server on port 8083
	log.Fatal(http.ListenAndServe(":8083", nil))
}
