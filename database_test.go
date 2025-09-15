//go:build !js

package main

import (
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

		embeddedEtcd, etcdClient, client := database()
		if embeddedEtcd == nil {
			t.Error("database() returned nil embedded etcd")
		}
		if etcdClient == nil {
			t.Error("database() returned nil etcd client")
		}
		if client == nil {
			t.Error("database() returned nil Client")
		}

		// Clean shutdown
		if embeddedEtcd != nil && etcdClient != nil {
			shutdown(embeddedEtcd, etcdClient)
		}
	}()

	select {
	case <-done:
		// Test completed
	case <-time.After(30 * time.Second):
		t.Error("database() test timed out")
	}
}

func TestShutdown(t *testing.T) {
	// Test that shutdown doesn't panic
	embeddedEtcd, etcdClient, _ := database()

	if embeddedEtcd == nil || etcdClient == nil {
		t.Skip("Failed to initialize database")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("shutdown() panicked: %v", r)
		}
	}()

	shutdown(embeddedEtcd, etcdClient)
}