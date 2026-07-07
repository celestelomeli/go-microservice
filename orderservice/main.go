package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Order maps to the "orders" table.
type Order struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
}

// Product and User are used to decode responses from the other services.
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

var db *gorm.DB

// Service URLs are overridable via env for other environments and tests;
// defaults match the docker-compose service names.
var (
	productServiceURL = envOr("PRODUCT_SERVICE_URL", "http://productservice:8081")
	userServiceURL    = envOr("USER_SERVICE_URL", "http://userservice:8083")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func initDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&Order{}); err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}
}

// createOrderHandler handles POST /orders. It validates the user and product
// against the other services, calculates the total, and saves the order.
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		return
	}

	// IDs are assigned by the database, never by the client.
	order.ID = 0

	if order.UserID == 0 || order.ProductID == 0 || order.Quantity == 0 {
		http.Error(w, "UserID, ProductID and Quantity are required", http.StatusBadRequest)
		return
	}

	_, err = getUser(order.UserID)
	if err != nil {
		http.Error(w, "Invalid user: "+err.Error(), http.StatusBadRequest)
		return
	}

	product, err := getProduct(order.ProductID)
	if err != nil {
		http.Error(w, "Error fetching product details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	order.Total = product.Price * float64(order.Quantity)

	result := db.Create(&order)
	if result.Error != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// getOrdersHandler handles GET /orders.
func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	result := db.Find(&orders)
	if result.Error != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// getOrderByIDHandler handles GET /orders/{id}.
func getOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order Order
	result := db.First(&order, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch order", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// deleteOrderHandler handles DELETE /orders/{id}.
func deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	result := db.Delete(&Order{}, id)
	if result.Error != nil {
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateOrderHandler handles PUT /orders/{id}.
func updateOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var existing Order
	result := db.First(&existing, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch order", http.StatusInternalServerError)
		}
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var updateData Order
	if err := json.Unmarshal(body, &updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updateData.ProductID == 0 || updateData.Quantity == 0 {
		http.Error(w, "ProductID and Quantity are required", http.StatusBadRequest)
		return
	}

	product, err := getProduct(updateData.ProductID)
	if err != nil {
		http.Error(w, "Failed to fetch product info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	existing.ProductID = updateData.ProductID
	existing.Quantity = updateData.Quantity
	existing.Total = product.Price * float64(updateData.Quantity)

	saveResult := db.Save(&existing)
	if saveResult.Error != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// getProduct fetches product info from productservice.
func getProduct(productID int) (Product, error) {
	url := fmt.Sprintf("%s/products/%d", productServiceURL, productID)

	resp, err := http.Get(url)
	if err != nil {
		return Product{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Product{}, fmt.Errorf("product service returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Product{}, fmt.Errorf("error reading response body: %w", err)
	}

	var product Product
	if err := json.Unmarshal(body, &product); err != nil {
		return Product{}, fmt.Errorf("error unmarshalling product JSON: %w", err)
	}

	return product, nil
}

// getUser fetches user info from userservice.
func getUser(userID int) (User, error) {
	url := fmt.Sprintf("%s/users/%d", userServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return User{}, fmt.Errorf("error calling user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("user service error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, fmt.Errorf("error reading response body: %w", err)
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, fmt.Errorf("unmarshal error: %w", err)
	}

	return user, nil
}

// ordersRouter dispatches /orders requests by method and path.
func ordersRouter(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/orders":
		createOrderHandler(w, r)

	case r.Method == http.MethodGet && (r.URL.Path == "/orders" || r.URL.Path == "/orders/"):
		getOrdersHandler(w, r)

	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/orders/"):
		getOrderByIDHandler(w, r)

	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/orders/"):
		deleteOrderHandler(w, r)

	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/orders/"):
		updateOrderHandler(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()

	http.HandleFunc("/orders", ordersRouter)
	http.HandleFunc("/orders/", ordersRouter)

	log.Println("Order Service listening on port 8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}
