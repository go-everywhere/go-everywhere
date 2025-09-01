//go:build !js

package main

import (
	"context"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	// This test ensures the database function doesn't panic
	// and returns valid instances
	
	// Create a timeout to prevent hanging
	done := make(chan bool)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("database() panicked: %v", r)
			}
			done <- true
		}()
		
		olricDB, client := database()
		if olricDB == nil {
			t.Error("database() returned nil Olric instance")
		}
		if client == nil {
			t.Error("database() returned nil Client")
		}
		
		// Clean shutdown
		if olricDB != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			olricDB.Shutdown(ctx)
		}
	}()
	
	select {
	case <-done:
		// Test completed
	case <-time.After(10 * time.Second):
		t.Error("database() test timed out")
	}
}

func TestShutdown(t *testing.T) {
	// Test that shutdown doesn't panic
	olricDB, _ := database()
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("shutdown() panicked: %v", r)
		}
	}()
	
	shutdown(olricDB)
}