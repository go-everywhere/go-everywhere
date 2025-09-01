//go:build !js

package db

import (
	"context"
	"testing"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
)

func newTestOlric(t *testing.T) (*olric.Olric, *olric.EmbeddedClient) {
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
	return olricDB, olricDB.NewEmbeddedClient()
}

func TestNewClient(t *testing.T) {
	olricDB, embedded := newTestOlric(t)
	defer olricDB.Shutdown(context.Background())

	client := NewClient(embedded)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.embedded != embedded {
		t.Error("Client embedded client mismatch")
	}

	if client.dmaps == nil {
		t.Error("Client dmaps map not initialized")
	}
}

func TestGetDMap(t *testing.T) {
	olricDB, embedded := newTestOlric(t)
	defer olricDB.Shutdown(context.Background())

	client := NewClient(embedded)

	// Test creating a new DMap
	dm, err := client.GetDMap("test-map")
	if err != nil {
		t.Fatalf("Failed to get DMap: %v", err)
	}

	if dm == nil {
		t.Error("GetDMap returned nil DMap")
	}

	// Test getting the same DMap again (should be cached)
	dm2, err := client.GetDMap("test-map")
	if err != nil {
		t.Fatalf("Failed to get cached DMap: %v", err)
	}

	if dm != dm2 {
		t.Error("GetDMap did not return cached DMap")
	}
}

func TestUsers(t *testing.T) {
	olricDB, embedded := newTestOlric(t)
	defer olricDB.Shutdown(context.Background())

	client := NewClient(embedded)

	dm, err := client.Users()
	if err != nil {
		t.Fatalf("Failed to get Users DMap: %v", err)
	}

	if dm == nil {
		t.Error("Users() returned nil DMap")
	}
}

func TestPutAndGet(t *testing.T) {
	olricDB, embedded := newTestOlric(t)
	defer olricDB.Shutdown(context.Background())

	client := NewClient(embedded)

	testData := []byte("test value")

	// Test Put
	err := client.Put(context.Background(), "test-map", "test-key", testData)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Test Get
	res, err := client.Get(context.Background(), "test-map", "test-key")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	var retrievedData []byte
	err = res.Scan(&retrievedData)
	if err != nil {
		t.Fatalf("Failed to scan data: %v", err)
	}

	if string(retrievedData) != string(testData) {
		t.Errorf("Data mismatch: expected %s, got %s", testData, retrievedData)
	}
}

func TestInitializeDMaps(t *testing.T) {
	olricDB, embedded := newTestOlric(t)
	defer olricDB.Shutdown(context.Background())

	// Test that InitializeDMaps doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitializeDMaps panicked: %v", r)
		}
	}()

	InitializeDMaps(embedded)
}