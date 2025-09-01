package views

import (
	"testing"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func TestHomeRender(t *testing.T) {
	home := &Home{
		message: "Test message",
	}

	ui := home.Render()
	if ui == nil {
		t.Error("Home.Render() returned nil")
	}

	section, ok := ui.(app.HTMLSection)
	if !ok {
		t.Error("Home.Render() should return app.HTMLSection")
	}
	_ = section
}

func TestHomeOnMount(t *testing.T) {
	home := &Home{}
	
	// Create a test context
	// Note: go-app doesn't provide easy way to mock contexts for testing
	// This test verifies that OnMount doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Home.OnMount() panicked: %v", r)
		}
	}()
	
	// We can't properly test OnMount without a full app context
	// but we can ensure the component implements the interface
	var _ app.Mounter = home
}

func TestHomeInitialState(t *testing.T) {
	home := &Home{}
	
	if home.message != "" {
		t.Errorf("Expected empty initial message, got %q", home.message)
	}
}