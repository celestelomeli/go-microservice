package main

//import libraries
import (
    "encoding/json" // encode/decode JSON
    "fmt"           // format strings
    "io"            // read data from requests/responses
    "log"           // logging errors
    "net/http"      // Handle HTTP requests
)

// Order struct to hold order data
// struct used to encode/decode JSON
type Order struct {
    ID        int     `json:"id"`
    UserID    int     `json:"user_id"`     // adding new user info
    ProductID int     `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Total     float64 `json:"total"`
}


// Global variable to store orders (in-memory)
// Substitute with a database
var orders = []Order{}
var nextOrderID = 1 // simulate auto-increment IDs like database

// createOrderHandler handles creating a new order
// handles POST requests to /orders
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	 // Method check, check if the request method is POST
	 //only allow POST requests, return 405 error if not
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

	// read JSON payload user sent/read request body into memory
	body, err := io.ReadAll(r.Body)
	if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	//close the body after reading 
	defer r.Body.Close()
	
	//convert JSON from request body into Go struct/order struct
	var order Order
	err = json.Unmarshal(body, &order)
	if err != nil {
        http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		return
	}
	// Check if the required fields are present in the order data
	// productid and quantity should not be 0
    if order.ProductID == 0 || order.Quantity == 0 {
        http.Error(w, "ProductID and Quantity are required", http.StatusBadRequest)
        return
    }
    user, err := getUser(order.UserID)
    if err != nil {
        http.Error(w, "Invalid user: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // Call product service to fetch product details 
    product, err := getProduct(order.ProductID)
    if err != nil {
        http.Error(w, "Error fetching product details: "+err.Error(), http.StatusInternalServerError)
        return //  Return after error to prevent further processing
    }




    // Calculate the total cost of order using product price and quantity
    order.Total = product.Price * float64(order.Quantity)
    order.ID = nextOrderID   //assign unique id to order
    nextOrderID++            //increment id for the next order

	// add newly created order to in memory list/orders slice
    orders = append(orders, order) 

    // Respond with the created order
	//set response headers to indicate returning JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated) // Use 201 Created for successful POST requests
    //encode order as JSON and send it back to client
	json.NewEncoder(w).Encode(order)
}
// Product struct to hold product details fetched from another service
type Product struct {
    ID    int     `json:"id"`    
    Name  string  `json:"name"` 
    Price float64 `json:"price"` 
}

// Add user struct to hole user details fetched from userservice
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// function to get product details by making HTTP GET request to another microservice
func getProduct(productID int) (Product, error) {
	//url for product service endpoint
    url := fmt.Sprintf("http://productservice:8081/products/%d", productID)
    //make a GET request to product service
	resp, err := http.Get(url)
    if err != nil {
        return Product{}, fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close() //close response body after reading 

	//if status code is not 200 OK return error
    if resp.StatusCode != http.StatusOK {
        return Product{}, fmt.Errorf("product service returned status: %s", resp.Status)
    }
    
	//read body of response
    // io.ReadAll reads entire response body from HTTP response and returns as byte slice
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return Product{}, fmt.Errorf("error reading response body: %w", err)
    }
	//convert json response body into a Go product struct
    var product Product
    // Unmarshall vs marshall 
    // unmarshal converts JSON or other formats into Go data types like structs
    //body byte slice passed that holds raw JSON string
    err = json.Unmarshal(body, &product) 
    if err != nil {
        return Product{}, fmt.Errorf("error unmarshalling product JSON: %w", err)
    }
    
    return product, nil
}


func getUser(userID int) (User, error) {
    url := fmt.Sprintf("http://userservice:8083/users/%d", userID) // or whatever port userservice runs on
    resp, err := http.Get(url)
    if err != nil {
        return User{}, fmt.Errorf("error calling user service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return User{}, fmt.Errorf("user service error: %s", resp.Status)
    }

    var user User
    body, _ := io.ReadAll(resp.Body)
    err = json.Unmarshal(body, &user)
    if err != nil {
        return User{}, fmt.Errorf("unmarshal error: %w", err)
    }

    return user, nil
}

// getOrdersHandler to return all orders as JSON
func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
   //encode and return the list of all orders 
	json.NewEncoder(w).Encode(orders)
}

func main() {
    // define route handlers
    http.HandleFunc("/orders", createOrderHandler) //handle POST requests
    http.HandleFunc("/orders/", getOrdersHandler) //handle GET requests

    // Start the HTTP server on port 8082
    fmt.Println("Order Service listening on port 8082")
    err := http.ListenAndServe(":8082", nil)
    if err != nil {
        log.Fatal(err)
    }
}




