//go:build !js

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMessage(t *testing.T) {
	handler := GetMessage()

	// Test GET request
	req := httptest.NewRequest(http.MethodGet, "/api/message", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected non-empty message text")
	}
	
	if response.Text != "Welcome to the Go PWA template!" {
		t.Errorf("Expected message 'Welcome to the Go PWA template!', got %q", response.Text)
	}
}

func TestGetMessageInvalidMethod(t *testing.T) {
	handler := GetMessage()

	// Test POST request (should fail)
	req := httptest.NewRequest(http.MethodPost, "/api/message", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for POST request, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}