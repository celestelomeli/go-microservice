package main

import (
	"encoding/json"       // For encoding and decoding JSON data
	"fmt"                 // For formatting strings
	"io"                  // For reading request/response bodies
	"log"                 // For logging errors and messages
	"net/http"            // For building the HTTP server and client
	"os"                  // For accessing environment variables

	"gorm.io/driver/postgres" // GORM PostgreSQL driver
	"gorm.io/gorm"            // GORM ORM library
)

/////////////////////////////////////////////////////////////
// Models
/////////////////////////////////////////////////////////////

// Order struct shows the schema of the "orders" table in the database
type Order struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
}

// Product and User are just used to decode responses from other services
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

/////////////////////////////////////////////////////////////
// Global database variable
/////////////////////////////////////////////////////////////

var db *gorm.DB // Shared DB connection used for all handlers

/////////////////////////////////////////////////////////////
// Database initialization
/////////////////////////////////////////////////////////////

func initDB() {
	// Build the connection string using environment variables
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Open connection using GORM
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Automatically create or update the "orders" table
	if err := db.AutoMigrate(&Order{}); err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}
}

/////////////////////////////////////////////////////////////
// HTTP Handlers
/////////////////////////////////////////////////////////////

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read JSON body sent by the client
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Clean up memory

	// Convert JSON to Go Order struct
	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if order.ProductID == 0 || order.Quantity == 0 {
		http.Error(w, "ProductID and Quantity are required", http.StatusBadRequest)
		return
	}

	// Fetch user from userservice to confirm they exist
	_, err = getUser(order.UserID)
	if err != nil {
		http.Error(w, "Invalid user: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch product info from productservice
	product, err := getProduct(order.ProductID)
	if err != nil {
		http.Error(w, "Error fetching product details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total cost (price * quantity)
	order.Total = product.Price * float64(order.Quantity)

	// Insert the new order into the database
	result := db.Create(&order)
	if result.Error != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Send back the created order as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// Slice to hold fetched orders
	var orders []Order

	// Use GORM to retrieve all orders from DB
	result := db.Find(&orders)
	if result.Error != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	// Return all orders as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

/////////////////////////////////////////////////////////////
// External API calls to other microservices
/////////////////////////////////////////////////////////////

func getProduct(productID int) (Product, error) {
	// Build the URL for the productservice endpoint
	url := fmt.Sprintf("http://productservice:8081/products/%d", productID)

	// Send a GET request to productservice
	resp, err := http.Get(url)
	if err != nil {
		// Could not reach the service
		return Product{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error if productservice returns a non-200 status
	if resp.StatusCode != http.StatusOK {
		return Product{}, fmt.Errorf("product service returned status: %s", resp.Status)
	}

	// Read the body from the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Product{}, fmt.Errorf("error reading response body: %w", err)
	}

	// Decode the JSON response into a Product struct
	var product Product
	if err := json.Unmarshal(body, &product); err != nil {
		return Product{}, fmt.Errorf("error unmarshalling product JSON: %w", err)
	}

	// Return the product struct back to the caller
	return product, nil
}

func getUser(userID int) (User, error) {
	// Build the URL to call userservice
	url := fmt.Sprintf("http://userservice:8083/users/%d", userID)

	// Send a GET request to userservice
	resp, err := http.Get(url)
	if err != nil {
		// Could not connect to service
		return User{}, fmt.Errorf("error calling user service: %w", err)
	}
	defer resp.Body.Close()

	// Handle error if the user service did not return 200 OK
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("user service error: %s", resp.Status)
	}

	// Read the body of the response
	body, _ := io.ReadAll(resp.Body)

	// Decode the JSON into a User struct
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, fmt.Errorf("unmarshal error: %w", err)
	}

	// Return the User struct back to the handler
	return user, nil
}

/////////////////////////////////////////////////////////////
// Main entry point
/////////////////////////////////////////////////////////////

func main() {
	// Connect to database and run auto-migrations
	initDB()

	// Register routes for creating and fetching orders
	http.HandleFunc("/orders", createOrderHandler)
	http.HandleFunc("/orders/", getOrdersHandler)

	// Start HTTP server on port 8082
	fmt.Println("Order Service listening on port 8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}
