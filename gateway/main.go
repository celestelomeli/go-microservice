package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func proxyHandler(target string) http.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		proxy.ServeHTTP(w, r)
	}
}

func main() {
	http.HandleFunc("/products", proxyHandler("http://productservice:8081"))
	http.HandleFunc("/products/", proxyHandler("http://productservice:8081"))

	http.HandleFunc("/orders", proxyHandler("http://orderservice:8082"))
	http.HandleFunc("/orders/", proxyHandler("http://orderservice:8082"))

	http.HandleFunc("/users", proxyHandler("http://userservice:8083"))
	http.HandleFunc("/users/", proxyHandler("http://userservice:8083"))

	log.Println("API Gateway listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
