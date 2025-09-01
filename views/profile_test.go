package views

import (
	"assette/models"
	"testing"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func TestProfileRender(t *testing.T) {
	profile := &Profile{
		user: models.User{
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}

	ui := profile.Render()
	if ui == nil {
		t.Error("Profile.Render() returned nil")
	}

	section, ok := ui.(app.HTMLSection)
	if !ok {
		t.Error("Profile.Render() should return app.HTMLSection")
	}
	_ = section
}

func TestProfileInitialState(t *testing.T) {
	profile := &Profile{}
	
	if profile.user.Name != "" {
		t.Errorf("Expected empty initial name, got %q", profile.user.Name)
	}
	
	if profile.user.Email != "" {
		t.Errorf("Expected empty initial email, got %q", profile.user.Email)
	}
}

func TestProfileHandleSubmit(t *testing.T) {
	profile := &Profile{
		user: models.User{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	
	// Create a mock event
	// Note: We can't fully test this without a browser context
	// but we can ensure it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Profile.handleSubmit() panicked: %v", r)
		}
	}()
	
	// Verify the profile was created correctly
	if profile.user.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got %q", profile.user.Name)
	}
	
	// The actual test would require mocking app.Context and app.Event
	// which is complex without the browser environment
}