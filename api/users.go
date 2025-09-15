//go:build !js

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"assette/db"
	"assette/models"
)

// CreateUser creates a new user
func CreateUser(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate a simple ID (in production, use UUID or similar)
		userID := fmt.Sprintf("user:%d", generateID())

		if err := client.Put(context.Background(), "users", userID, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the created user with ID
		response := map[string]interface{}{
			"id":    userID,
			"name":  user.Name,
			"email": user.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUser retrieves a single user by ID
func GetUser(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL path (e.g., /api/users/user:1)
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		userID := pathParts[3]

		userData, err := client.Get(context.Background(), "users", userID)
		if err != nil {
			if err == db.ErrKeyNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var user models.User
		err = json.Unmarshal(userData, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":    userID,
			"name":  user.Name,
			"email": user.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ListUsers retrieves all users
func ListUsers(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Get all users from the users namespace
		usersData, err := client.GetAll(context.Background(), "users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		users := []map[string]interface{}{}
		for userID, userData := range usersData {
			var user models.User
			err = json.Unmarshal(userData, &user)
			if err != nil {
				continue // Skip malformed user data
			}

			users = append(users, map[string]interface{}{
				"id":    userID,
				"name":  user.Name,
				"email": user.Email,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	}
}

// UpdateUser updates an existing user
func UpdateUser(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		userID := pathParts[3]

		// Check if user exists
		_, err := client.Get(context.Background(), "users", userID)
		if err != nil {
			if err == db.ErrKeyNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := client.Put(context.Background(), "users", userID, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":    userID,
			"name":  user.Name,
			"email": user.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteUser deletes a user
func DeleteUser(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		userID := pathParts[3]

		// Check if user exists
		_, err := client.Get(context.Background(), "users", userID)
		if err != nil {
			if err == db.ErrKeyNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete the user
		_, err = client.Delete(context.Background(), "users", userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UserRouter handles routing for all user endpoints
func UserRouter(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Route to appropriate handler based on path and method
		if path == "/api/users" {
			switch r.Method {
			case http.MethodGet:
				ListUsers(client)(w, r)
			case http.MethodPost:
				CreateUser(client)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasPrefix(path, "/api/users/") {
			switch r.Method {
			case http.MethodGet:
				GetUser(client)(w, r)
			case http.MethodPut:
				UpdateUser(client)(w, r)
			case http.MethodDelete:
				DeleteUser(client)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

// Simple ID generator (in production, use UUID or database sequence)
var idCounter = 0

func generateID() int {
	idCounter++
	return idCounter
}