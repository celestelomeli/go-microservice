package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestProxyForwardsToBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "backend response")
	}))
	defer backend.Close()

	handler := proxyHandler(backend.URL)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "backend response" {
		t.Errorf("expected backend response to pass through, got %q", body)
	}
}

func TestProxySetsCORSHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer backend.Close()

	handler := proxyHandler(backend.URL)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin *, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods to be set")
	}
}

func TestProxyPreflightSkipsBackend(t *testing.T) {
	backendHit := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHit = true
	}))
	defer backend.Close()

	handler := proxyHandler(backend.URL)
	req := httptest.NewRequest(http.MethodOptions, "/products", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for preflight, got %d", rec.Code)
	}
	if backendHit {
		t.Error("OPTIONS preflight should be answered by the gateway, not forwarded")
	}
}

func TestProxyBackendUnreachable(t *testing.T) {
	// The reverse proxy logs the dial failure we provoke here; discard it
	// so a passing run reads clean.
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	// Port 1 is reserved and nothing listens there, so the proxy's error
	// handler should answer with 502 rather than hanging or panicking.
	handler := proxyHandler("http://127.0.0.1:1")
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 when backend is down, got %d", rec.Code)
	}
}
