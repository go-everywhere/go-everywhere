package models

import (
	"encoding/json"
	"testing"
)

func TestUserStruct(t *testing.T) {
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", user.Name)
	}

	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %q", user.Email)
	}
}

func TestUserJSONMarshaling(t *testing.T) {
	user := User{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	// Test marshaling
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal User: %v", err)
	}

	// Test unmarshaling
	var unmarshaledUser User
	err = json.Unmarshal(data, &unmarshaledUser)
	if err != nil {
		t.Fatalf("Failed to unmarshal User: %v", err)
	}

	if unmarshaledUser.Name != user.Name {
		t.Errorf("Name mismatch after marshaling: expected %q, got %q", user.Name, unmarshaledUser.Name)
	}

	if unmarshaledUser.Email != user.Email {
		t.Errorf("Email mismatch after marshaling: expected %q, got %q", user.Email, unmarshaledUser.Email)
	}
}

func TestUserEmptyValues(t *testing.T) {
	var user User

	if user.Name != "" {
		t.Errorf("Expected empty name for zero value User, got %q", user.Name)
	}

	if user.Email != "" {
		t.Errorf("Expected empty email for zero value User, got %q", user.Email)
	}
}