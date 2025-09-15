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
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func newTestDB(t *testing.T) (*embed.Etcd, *clientv3.Client, *db.Client) {
	// Create unique data directory for each test
	dataDir := filepath.Join(os.TempDir(), fmt.Sprintf("etcd-test-%d", time.Now().UnixNano()))

	cfg := embed.NewConfig()
	cfg.Dir = dataDir
	cfg.LogLevel = "error"

	// Set a unique name for this test instance
	cfg.Name = fmt.Sprintf("test-%d", time.Now().UnixNano())

	// Use random ports to avoid conflicts
	lcurl, _ := url.Parse("http://127.0.0.1:0")
	cfg.ListenClientUrls = []url.URL{*lcurl}
	cfg.AdvertiseClientUrls = []url.URL{*lcurl}

	lpurl, _ := url.Parse("http://127.0.0.1:0")
	cfg.ListenPeerUrls = []url.URL{*lpurl}
	cfg.AdvertisePeerUrls = []url.URL{*lpurl}

	cfg.InitialCluster = fmt.Sprintf("%s=%s", cfg.Name, lpurl.String())
	cfg.StrictReconfigCheck = false
	cfg.InitialClusterToken = "etcd-test-cluster"

	// Start embedded etcd
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("Failed to start embedded etcd: %v", err)
	}

	// Wait for etcd to be ready
	select {
	case <-e.Server.ReadyNotify():
		// Server is ready
	case <-time.After(10 * time.Second):
		e.Server.Stop()
		e.Close()
		os.RemoveAll(dataDir)
		t.Fatal("Embedded etcd took too long to start")
	}

	// Get the actual client URL
	clientURL := e.Clients[0].Addr().String()

	// Create client connection
	config := clientv3.Config{
		Endpoints:   []string{clientURL},
		DialTimeout: 5 * time.Second,
	}

	etcdClient, err := clientv3.New(config)
	if err != nil {
		e.Server.Stop()
		e.Close()
		os.RemoveAll(dataDir)
		t.Fatalf("Failed to create etcd client: %v", err)
	}

	client := db.NewClient(etcdClient)
	db.InitializeNamespaces(client)

	// Clean up function
	t.Cleanup(func() {
		etcdClient.Close()
		e.Server.Stop()
		e.Close()
		os.RemoveAll(dataDir)
	})

	return e, etcdClient, client
}

func TestCreateUser(t *testing.T) {
	_, _, client := newTestDB(t)

	user := models.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := CreateUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["name"] != user.Name {
		t.Errorf("Expected name %s, got %v", user.Name, response["name"])
	}

	if response["email"] != user.Email {
		t.Errorf("Expected email %s, got %v", user.Email, response["email"])
	}

	if response["id"] == nil {
		t.Error("Expected ID to be set")
	}
}

func TestGetUser(t *testing.T) {
	_, _, client := newTestDB(t)

	// Create a user first
	user := models.User{
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	userID := "user:test123"
	err := client.Put(context.Background(), "users", userID, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test getting the user
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	w := httptest.NewRecorder()

	handler := GetUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["name"] != user.Name {
		t.Errorf("Expected name %s, got %v", user.Name, response["name"])
	}
}

func TestGetUserNotFound(t *testing.T) {
	_, _, client := newTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user:nonexistent", nil)
	w := httptest.NewRecorder()

	handler := GetUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestListUsers(t *testing.T) {
	_, _, client := newTestDB(t)

	// Create multiple users
	users := []struct {
		ID   string
		User models.User
	}{
		{"user:1", models.User{Name: "User 1", Email: "user1@example.com"}},
		{"user:2", models.User{Name: "User 2", Email: "user2@example.com"}},
		{"user:3", models.User{Name: "User 3", Email: "user3@example.com"}},
	}

	for _, u := range users {
		err := client.Put(context.Background(), "users", u.ID, u.User)
		if err != nil {
			t.Fatalf("Failed to create test user %s: %v", u.ID, err)
		}
	}

	// Test listing users
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	handler := ListUsers(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	userList, ok := response["users"].([]interface{})
	if !ok {
		t.Fatal("Expected users to be an array")
	}

	if len(userList) < len(users) {
		t.Errorf("Expected at least %d users, got %d", len(users), len(userList))
	}
}

func TestUpdateUser(t *testing.T) {
	_, _, client := newTestDB(t)

	// Create a user first
	userID := "user:update123"
	originalUser := models.User{
		Name:  "Original Name",
		Email: "original@example.com",
	}

	err := client.Put(context.Background(), "users", userID, originalUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update the user
	updatedUser := models.User{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	body, _ := json.Marshal(updatedUser)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%s", userID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := UpdateUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["name"] != updatedUser.Name {
		t.Errorf("Expected name %s, got %v", updatedUser.Name, response["name"])
	}

	if response["email"] != updatedUser.Email {
		t.Errorf("Expected email %s, got %v", updatedUser.Email, response["email"])
	}

	// Verify the update persisted
	data, err := client.Get(context.Background(), "users", userID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	var storedUser models.User
	json.Unmarshal(data, &storedUser)

	if storedUser.Name != updatedUser.Name {
		t.Errorf("Update not persisted: expected name %s, got %s", updatedUser.Name, storedUser.Name)
	}
}

func TestDeleteUser(t *testing.T) {
	_, _, client := newTestDB(t)

	// Create a user first
	userID := "user:delete123"
	user := models.User{
		Name:  "To Delete",
		Email: "delete@example.com",
	}

	err := client.Put(context.Background(), "users", userID, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Delete the user
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%s", userID), nil)
	w := httptest.NewRecorder()

	handler := DeleteUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	// Verify deletion
	_, err = client.Get(context.Background(), "users", userID)
	if err != db.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after deletion, got: %v", err)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	_, _, client := newTestDB(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/user:nonexistent", nil)
	w := httptest.NewRecorder()

	handler := DeleteUser(client)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestUserRouter(t *testing.T) {
	_, _, client := newTestDB(t)

	router := UserRouter(client)

	// Test routing to ListUsers
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	router(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Router failed for GET /api/users: got status %d", w.Result().StatusCode)
	}

	// Test invalid method
	req = httptest.NewRequest(http.MethodPatch, "/api/users", nil)
	w = httptest.NewRecorder()
	router(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Router should return 405 for PATCH /api/users: got status %d", w.Result().StatusCode)
	}
}