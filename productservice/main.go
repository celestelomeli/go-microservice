package main // Declares package 

// Imports packages from Go library
import (
	"encoding/json" // Working with JSON data, convert Go structs to JSON and vice versa 
	"fmt" // Formatted input and output 
	"log" // Logging errors 
	"net/http" // Creating HTTP servers and handling requests
	"strconv" // Converting strings to numbers, numbers to strings
)

// Struct to hold product data.
type Product struct { //
	ID int `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
}

// Products is a mock database. Hardcoded piece of Product structs.
var products = []Product{
	{ID: 1, Name: "Laptop", Price: 1300.00},
	{ID: 2, Name: "Mouse", Price: 20.00},
	{ID: 3, Name: "Keyboard", Price: 75.00},
	{ID: 4, Name: "Monitor", Price: 500.00},
}

//getProductHandler returns all products
// Handler function for HTTP request 
func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//w is response writer, write HTTP response to this
	//r is the request, contains the info about the incoming HTTP request
	//set the content type header in the response to tell the client 
	//the response will be in JSON format. So the client knows how to interpret the data

	json.NewEncoder(w).Encode(products)
	//creates new JSON encoder to convert Go data structures into JSON
	// we pass the response writer (w) to the encoder. it will
	// write the json output to the http response 
	// encode method takes GO data structure and converts it into JSON string
	// and then writes that string to the response writer (w)

	//  In essence, this line does the following:
    //  1.  Takes the 'products' data (a Go slice).
    //  2.  Converts it into a JSON string.
    //  3.  Sends that JSON string as the body of the HTTP response.

}
	//getProduct Handler returns single product by ID
func getProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get ID from URL path 
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		 //  - err != nil:  This checks if an error occurred during the string-to-integer
		// http.Error:  This is a helper function from the "net/http" package
        //    for sending an HTTP error response.
		//  - http.StatusBadRequest:  This is a constant from the "net/http" package
        //     that represents the HTTP status code "400 Bad Request".  This
        //     indicates that the client sent an invalid request (in this case,
        //     an invalid product ID).
        //  - return:  This is a crucial statement.  If an error occurs, we
        //     *must* return from the function.  This prevents the rest of the
        //     function from executing, which would likely lead to a crash or
        //     incorrect behavior.
        return
	}

	for _, product := range products {
		if product.ID == id {
			json.NewEncoder(w).Encode(product)
			return
		}
	}
 //if loop finishes without finding a product with matching ID, we get here

	http.Error(w, "Product not found", http.StatusNotFound)
	//  We send this error *only* if the product ID wasn't found in our
    //  'products' slice.
}

func main() {
	http.HandleFunc("/products", getProductsHandler)
	http.HandleFunc("/products/", getProductHandler)
	fmt.Println("Product Service listening on port 8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}