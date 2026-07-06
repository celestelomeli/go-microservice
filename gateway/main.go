package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// envOr returns the env value for key, or fallback when unset;
// defaults match the docker-compose service names.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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
	productURL := envOr("PRODUCT_SERVICE_URL", "http://productservice:8081")
	orderURL := envOr("ORDER_SERVICE_URL", "http://orderservice:8082")
	userURL := envOr("USER_SERVICE_URL", "http://userservice:8083")

	http.HandleFunc("/products", proxyHandler(productURL))
	http.HandleFunc("/products/", proxyHandler(productURL))

	http.HandleFunc("/orders", proxyHandler(orderURL))
	http.HandleFunc("/orders/", proxyHandler(orderURL))

	http.HandleFunc("/users", proxyHandler(userURL))
	http.HandleFunc("/users/", proxyHandler(userURL))

	log.Println("API Gateway listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
