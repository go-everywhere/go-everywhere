//go:build !js

package api

import (
	"assette/db"
	"assette/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
)

func newTestDB(t *testing.T) (*olric.Olric, *db.Client) {
	c := config.New("local")
	c.BindAddr = "127.0.0.1"
	c.MemberlistConfig.BindAddr = "127.0.0.1"

	ctx, cancel := context.WithCancel(context.Background())
	c.Started = func() {
		defer cancel()
	}

	olricDB, err := olric.New(c)
	if err != nil {
		t.Fatalf("Failed to create Olric: %v", err)
	}

	go func() {
		if err := olricDB.Start(); err != nil {
			// Log error but don't fail test
		}
	}()

	<-ctx.Done()
	
	embedded := olricDB.NewEmbeddedClient()
	db.InitializeDMaps(embedded)
	client := db.NewClient(embedded)
	
	return olricDB, client
}

func TestCreateUser(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	handler := CreateUser(client)

	user := models.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	body, _ := json.Marshal(user)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		t.Errorf("Response body: %s", w.Body.String())
	}

	// Check response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["name"] != user.Name {
		t.Errorf("Expected name %s, got %v", user.Name, response["name"])
	}

	if response["email"] != user.Email {
		t.Errorf("Expected email %s, got %v", user.Email, response["email"])
	}

	if response["id"] == nil {
		t.Error("Expected ID in response")
	}
}

func TestGetUser(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	// First create a user
	user := models.User{
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}
	userData, _ := json.Marshal(user)
	client.Put(context.Background(), "users", "user:1", userData)

	handler := GetUser(client)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user:1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["name"] != user.Name {
		t.Errorf("Expected name %s, got %v", user.Name, response["name"])
	}
}

func TestGetUserNotFound(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	handler := GetUser(client)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user:999", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestListUsers(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	// Create multiple users
	users := []models.User{
		{Name: "User 1", Email: "user1@example.com"},
		{Name: "User 2", Email: "user2@example.com"},
	}

	for i, user := range users {
		userData, _ := json.Marshal(user)
		client.Put(context.Background(), "users", fmt.Sprintf("user:%d", i+1), userData)
	}

	handler := ListUsers(client)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	userList, ok := response["users"].([]interface{})
	if !ok {
		t.Fatal("Expected users array in response")
	}

	if len(userList) != len(users) {
		t.Errorf("Expected %d users, got %d", len(users), len(userList))
	}

	count, ok := response["count"].(float64)
	if !ok || int(count) != len(users) {
		t.Errorf("Expected count %d, got %v", len(users), response["count"])
	}
}

func TestUpdateUser(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	// First create a user
	originalUser := models.User{
		Name:  "Original Name",
		Email: "original@example.com",
	}
	userData, _ := json.Marshal(originalUser)
	client.Put(context.Background(), "users", "user:1", userData)

	// Update the user
	updatedUser := models.User{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}
	body, _ := json.Marshal(updatedUser)

	handler := UpdateUser(client)

	req := httptest.NewRequest(http.MethodPut, "/api/users/user:1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Verify the update
	res, _ := client.Get(context.Background(), "users", "user:1")
	var storedData []byte
	res.Scan(&storedData)
	
	var storedUser models.User
	json.Unmarshal(storedData, &storedUser)

	if storedUser.Name != updatedUser.Name {
		t.Errorf("Expected name %s, got %s", updatedUser.Name, storedUser.Name)
	}

	if storedUser.Email != updatedUser.Email {
		t.Errorf("Expected email %s, got %s", updatedUser.Email, storedUser.Email)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	updatedUser := models.User{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}
	body, _ := json.Marshal(updatedUser)

	handler := UpdateUser(client)

	req := httptest.NewRequest(http.MethodPut, "/api/users/user:999", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteUser(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	// First create a user
	user := models.User{
		Name:  "To Delete",
		Email: "delete@example.com",
	}
	userData, _ := json.Marshal(user)
	client.Put(context.Background(), "users", "user:1", userData)

	handler := DeleteUser(client)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/user:1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deletion
	_, err := client.Get(context.Background(), "users", "user:1")
	if err != olric.ErrKeyNotFound {
		t.Error("Expected user to be deleted")
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	handler := DeleteUser(client)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/user:999", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUserRouter(t *testing.T) {
	olricDB, client := newTestDB(t)
	defer olricDB.Shutdown(context.Background())

	handler := UserRouter(client)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "List users",
			method:     http.MethodGet,
			path:       "/api/users",
			body:       nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Create user",
			method:     http.MethodPost,
			path:       "/api/users",
			body:       models.User{Name: "Test", Email: "test@example.com"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Invalid method for /api/users",
			method:     http.MethodDelete,
			path:       "/api/users",
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Get non-existent user",
			method:     http.MethodGet,
			path:       "/api/users/user:999",
			body:       nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

