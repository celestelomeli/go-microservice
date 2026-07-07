package main

import (
	"encoding/json"
	"fmt"
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
	if err := db.AutoMigrate(&Order{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestGetOrders(t *testing.T) {
	setupTestDB(t)
	db.Create(&[]Order{
		{UserID: 1, ProductID: 1, Quantity: 2, Total: 40},
		{UserID: 1, ProductID: 3, Quantity: 1, Total: 75},
	})

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var orders []Order
	if err := json.NewDecoder(rec.Body).Decode(&orders); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(orders))
	}
}

func TestGetOrderByID(t *testing.T) {
	setupTestDB(t)
	db.Create(&Order{UserID: 1, ProductID: 1, Quantity: 2, Total: 40})

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var order Order
	if err := json.NewDecoder(rec.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if order.Total != 40 {
		t.Errorf("expected total 40, got %v", order.Total)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/orders/999", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetOrderInvalidID(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/orders/abc", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteOrder(t *testing.T) {
	setupTestDB(t)
	db.Create(&Order{UserID: 1, ProductID: 1, Quantity: 2, Total: 40})

	req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	var count int64
	db.Model(&Order{}).Count(&count)
	if count != 0 {
		t.Errorf("expected order to be deleted, %d remain", count)
	}
}

func TestDeleteOrderNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodDelete, "/orders/999", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// fakeService stands in for userservice/productservice, answering every
// request with the given status and body.
func fakeService(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
}

func setFakeBackends(t *testing.T, userStatus int, userBody string, productStatus int, productBody string) {
	t.Helper()
	users := fakeService(t, userStatus, userBody)
	products := fakeService(t, productStatus, productBody)
	origUser, origProduct := userServiceURL, productServiceURL
	userServiceURL, productServiceURL = users.URL, products.URL
	t.Cleanup(func() {
		userServiceURL, productServiceURL = origUser, origProduct
		users.Close()
		products.Close()
	})
}

func TestCreateOrderSuccess(t *testing.T) {
	setupTestDB(t)
	setFakeBackends(t,
		http.StatusOK, `{"id":1,"name":"Demo User","email":"demo@example.com"}`,
		http.StatusOK, `{"id":2,"name":"Mouse","price":20}`,
	)

	body := strings.NewReader(`{"user_id":1,"product_id":2,"quantity":3}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", body)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var order Order
	if err := json.NewDecoder(rec.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if order.Total != 60 {
		t.Errorf("expected total 60 (3 x 20), got %v", order.Total)
	}
	if order.ID == 0 {
		t.Error("expected database-assigned ID, got 0")
	}
}

func TestCreateOrderUnknownUser(t *testing.T) {
	setupTestDB(t)
	setFakeBackends(t,
		http.StatusNotFound, `not found`,
		http.StatusOK, `{"id":2,"name":"Mouse","price":20}`,
	)

	body := strings.NewReader(`{"user_id":999,"product_id":2,"quantity":1}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", body)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown user, got %d", rec.Code)
	}
}

func TestUpdateOrderRecalculatesTotal(t *testing.T) {
	setupTestDB(t)
	db.Create(&Order{UserID: 1, ProductID: 2, Quantity: 1, Total: 20})
	setFakeBackends(t,
		http.StatusOK, `{"id":1,"name":"Demo User","email":"demo@example.com"}`,
		http.StatusOK, `{"id":2,"name":"Mouse","price":20}`,
	)

	body := strings.NewReader(`{"product_id":2,"quantity":5}`)
	req := httptest.NewRequest(http.MethodPut, "/orders/1", body)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var order Order
	if err := json.NewDecoder(rec.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if order.Total != 100 {
		t.Errorf("expected total 100 (5 x 20), got %v", order.Total)
	}
}

// Validation happens before any external call, so no fake backends are needed.
func TestCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"invalid JSON", `{not json`},
		{"missing user_id", `{"product_id":1,"quantity":2}`},
		{"missing product_id", `{"user_id":1,"quantity":2}`},
		{"missing quantity", `{"user_id":1,"product_id":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDB(t)

			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			ordersRouter(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestUpdateOrderNotFound(t *testing.T) {
	setupTestDB(t)

	body := strings.NewReader(`{"product_id":1,"quantity":2}`)
	req := httptest.NewRequest(http.MethodPut, "/orders/999", body)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestOrdersRouterMethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodPatch, "/orders/1", nil)
	rec := httptest.NewRecorder()
	ordersRouter(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
