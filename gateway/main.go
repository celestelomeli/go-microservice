package main

import (
	"log"                 //logging functions
	"net/http"            // HTTP server and client functionality
	"net/http/httputil"   //allows creation of reverse proxies
	"net/url"             // parsing and structuring URL strings 
)

// proxyHandler returns HTTP handler that forwards requests to specific target server
func proxyHandler(target string) http.HandlerFunc {
	// parse target URL string into structured 
	// url.parse from net/url parses raw URL into structured *url.URL object
	// url variable will hold parsed structured URL object
	url, err := url.Parse(target)
	// From log library, logs error message
	//fatalf = shortcut for log.Print() + os.Exit(1)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}
	// creates reverse proxy that knows how to forward requeststo specified host in url variable
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {
		// tells reverse proxy to take over
		//reads incoming request (r), sends to backend server, writes 
		// backend's response back through w 
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// Forward any request starting with "/products" to the service running on port 8081
	http.HandleFunc("/products", proxyHandler("http://localhost:8081"))
	// captures any nested routes like "/products/123"
	http.HandleFunc("/products/", proxyHandler("http://localhost:8081"))
	
	// Forward "/orders" requests to service on port 8082
	http.HandleFunc("/orders", proxyHandler("http://localhost:8082"))
    http.HandleFunc("/orders/", proxyHandler("http://localhost:8082"))

	// Forward "/users" requests to service on port 8083
    http.HandleFunc("/users", proxyHandler("http://localhost:8083"))
    http.HandleFunc("/users/", proxyHandler("http://localhost:8083"))

	// Log a message that the gateway is starting
	log.Println("API Gateway listening on port 8080")
	// Start the server on port 8080 and crash the app if there’s an error
	log.Fatal(http.ListenAndServe(":8080", nil))
}