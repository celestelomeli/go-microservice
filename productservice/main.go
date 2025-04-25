package main // Declares package 

// Import libraries
import (
	"encoding/json" //  Working with JSON data
	"fmt"           //  Formatted input/output
	"log"           //  Logging errors
	"net/http"      //  Creating HTTP servers and handling requests
	"strconv"       //  Converting strings to numbers and vice versa
)

// Product struct to hold product data
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Mock product database
var products = []Product{
	{ID: 1, Name: "Laptop", Price: 1300.00},
	{ID: 2, Name: "Mouse", Price: 20.00},
	{ID: 3, Name: "Keyboard", Price: 75.00},
	{ID: 4, Name: "Monitor", Price: 500.00},
}

// getProductsHandler returns all products in JSON format
func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json") 
	// Encode and send products as JSON
	json.NewEncoder(w).Encode(products) 
}

// getProductHandler returns a single product by ID
func getProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Set content type to JSON
	
	// Extract ID from URL and convert to integer
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest) // Error message
		return
	}

	// Find the product by ID
	for _, product := range products {
		if product.ID == id {
			json.NewEncoder(w).Encode(product) // Return product as JSON
			return
		}
	}

	// Return error if product not found
	http.Error(w, "Product not found", http.StatusNotFound)
}

func main() {
	// Handle HTTP requests
	http.HandleFunc("/products", getProductsHandler)     // Get all products
	http.HandleFunc("/products/", getProductHandler)     // Get a product by ID
	fmt.Println("Product Service listening on port 8081") 
	err := http.ListenAndServe(":8081", nil)             // Start server
	
	if err != nil {                                      // Log any server startup errors
		log.Fatal(err) 
	}
}
