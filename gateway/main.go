package main

import (
	"log"                 //logging functions
	"net/http"            // HTTP server and client functionality
	"net/http/httputil"   //allows creation of reverse proxies
	"net/url"             // parsing and structuring URL strings 

)

func proxyHandler(target string) http.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Forward everything else to the target
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// Forward any request starting with "/products" to the service running on port 8081
	http.HandleFunc("/products", proxyHandler("http://productservice:8081"))
	// captures any nested routes like "/products/123"
	http.HandleFunc("/products/", proxyHandler("http://productservice:8081"))
	
	// Forward "/orders" requests to service on port 8082
	http.HandleFunc("/orders", proxyHandler("http://orderservice:8082"))
    http.HandleFunc("/orders/", proxyHandler("http://orderservice:8082"))

	// Forward "/users" requests to service on port 8083
    http.HandleFunc("/users", proxyHandler("http://userservice:8083"))
    http.HandleFunc("/users/", proxyHandler("http://userservice:8083"))

	// Log a message that the gateway is starting
	log.Println("API Gateway listening on port 8080")
	// Start the server on port 8080 and crash the app if there’s an error
	log.Fatal(http.ListenAndServe(":8080", nil))
}