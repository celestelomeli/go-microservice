package main

// Import Go standard libraries
import (
	"encoding/json" // Encode/decode JSON
	"fmt"           // Format strings
	"log"           // Logging errors/info
	"net/http"      // HTTP server and request handling
	"os"            // Access env variables for DB config
	"strconv"       // Convert string to int
	"strings"       // Manipulate strings like URL paths

	// GORM and PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/////////////////////////////////////////////////////////////
// Database Model
/////////////////////////////////////////////////////////////

// Product maps to the "products" table in the database
type Product struct {
	ID    int     `json:"id" gorm:"primaryKey"` // Auto-increment ID
	Name  string  `json:"name"`                 // Product name
	Price float64 `json:"price"`                // Product price
}

/////////////////////////////////////////////////////////////
// Global DB Connection
/////////////////////////////////////////////////////////////

var db *gorm.DB // Shared DB connection for use in all handlers

/////////////////////////////////////////////////////////////
// Database Initialization
/////////////////////////////////////////////////////////////

func initDB() {
	// Build Postgres connection string from env vars
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Open GORM DB connection
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-create the products table if it doesn't exist
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatal("Failed to migrate schema:", err)
	}

	// Seed data only if the table is empty
	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		db.Create(&[]Product{
			{Name: "Laptop", Price: 1300.00},
			{Name: "Mouse", Price: 20.00},
			{Name: "Keyboard", Price: 75.00},
			{Name: "Monitor", Price: 500.00},
		})
	}
}

/////////////////////////////////////////////////////////////
// HTTP Handlers
/////////////////////////////////////////////////////////////

// getProductsHandler handles GET /products
// Returns all products from the DB as JSON
func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	var products []Product
	result := db.Find(&products) // SELECT * FROM products
	if result.Error != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// getProductHandler handles GET /products/{id}
// Returns a single product by ID
func getProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract the ID from the URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product Product
	result := db.First(&product, id) // SELECT * FROM products WHERE id = ?
	if result.Error != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(product)
}

// createProductHandler handles POST /products
// Accepts a JSON body and inserts a new product
func createProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var product Product

	// Decode JSON body into a Product struct
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save the new product to the DB
	result := db.Create(&product) // INSERT INTO products ...
	if result.Error != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	// Return the created product as confirmation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

/////////////////////////////////////////////////////////////
// Main Entry Point
/////////////////////////////////////////////////////////////

func main() {
	initDB() // Connect to DB and prepare schema

	// Route for /products handles both GET and POST
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getProductsHandler(w, r)
		} else if r.Method == http.MethodPost {
			createProductHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Route for /products/{id} to fetch a single product
	http.HandleFunc("/products/", getProductHandler)

	// Start HTTP server on port 8081
	fmt.Println("Product Service listening on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
