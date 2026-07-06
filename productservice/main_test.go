package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB swaps the package-level db for an in-memory SQLite database.
func setupTestDB(t *testing.T) {
	t.Helper()
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		// Silence GORM's error-level logging; not-found tests trigger it by design.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestGetProducts(t *testing.T) {
	setupTestDB(t)
	db.Create(&[]Product{
		{Name: "Laptop", Price: 1300},
		{Name: "Mouse", Price: 20},
	})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	getProductsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var products []Product
	if err := json.NewDecoder(rec.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
}

func TestGetProductByID(t *testing.T) {
	setupTestDB(t)
	db.Create(&Product{Name: "Laptop", Price: 1300})

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rec := httptest.NewRecorder()
	getProductHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var product Product
	if err := json.NewDecoder(rec.Body).Decode(&product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if product.Price != 1300 {
		t.Errorf("expected price 1300, got %v", product.Price)
	}
}

func TestGetProductNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	rec := httptest.NewRecorder()
	getProductHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetProductInvalidID(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	rec := httptest.NewRecorder()
	getProductHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateProduct(t *testing.T) {
	setupTestDB(t)

	body := strings.NewReader(`{"name":"Webcam","price":89.99}`)
	req := httptest.NewRequest(http.MethodPost, "/products", body)
	rec := httptest.NewRecorder()
	createProductHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var product Product
	if err := json.NewDecoder(rec.Body).Decode(&product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if product.ID == 0 {
		t.Error("expected database-assigned ID, got 0")
	}
}

func TestCreateProductIgnoresClientID(t *testing.T) {
	setupTestDB(t)

	body := strings.NewReader(`{"id":42,"name":"Webcam","price":89.99}`)
	req := httptest.NewRequest(http.MethodPost, "/products", body)
	rec := httptest.NewRecorder()
	createProductHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var product Product
	if err := json.NewDecoder(rec.Body).Decode(&product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if product.ID == 42 {
		t.Error("client-supplied ID should be ignored")
	}
}

func TestCreateProductInvalidJSON(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{not json`))
	rec := httptest.NewRecorder()
	createProductHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
