package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
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

// healthzHandler: the gateway has no dependencies of its own to check
// (a backend being down is that backend's problem, reported per-request
// as 502), so healthy just means alive and serving.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	productURL := envOr("PRODUCT_SERVICE_URL", "http://productservice:8081")
	orderURL := envOr("ORDER_SERVICE_URL", "http://orderservice:8082")
	userURL := envOr("USER_SERVICE_URL", "http://userservice:8083")

	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/products", proxyHandler(productURL))
	http.HandleFunc("/products/", proxyHandler(productURL))

	http.HandleFunc("/orders", proxyHandler(orderURL))
	http.HandleFunc("/orders/", proxyHandler(orderURL))

	http.HandleFunc("/users", proxyHandler(userURL))
	http.HandleFunc("/users/", proxyHandler(userURL))

	log.Println("API Gateway listening on port 8080")
	// WriteTimeout is generous because the gateway waits on downstream
	// services: an order creation can legitimately take several seconds
	// while orderservice calls its neighbors. An upstream's patience must
	// exceed its downstreams' worst case.
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
