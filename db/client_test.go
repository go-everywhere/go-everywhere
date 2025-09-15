//go:build !js

package db

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func newTestEtcd(t *testing.T) (*embed.Etcd, *clientv3.Client) {
	// Create unique data directory for each test
	dataDir := filepath.Join(os.TempDir(), fmt.Sprintf("etcd-test-%d", time.Now().UnixNano()))

	cfg := embed.NewConfig()
	cfg.Dir = dataDir
	cfg.LogLevel = "error"

	// Set a unique name for this test instance
	cfg.Name = fmt.Sprintf("test-%d", time.Now().UnixNano())

	// Use fixed ports for simplicity in testing
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

	// Clean up function
	t.Cleanup(func() {
		etcdClient.Close()
		e.Server.Stop()
		e.Close()
		os.RemoveAll(dataDir)
	})

	return e, etcdClient
}

func TestNewClient(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.etcdClient != etcdClient {
		t.Error("Client etcd client mismatch")
	}
}

func TestPutAndGet(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)

	testData := map[string]string{
		"name":  "Test User",
		"email": "test@example.com",
	}

	// Test Put
	err := client.Put(context.Background(), "test-namespace", "test-key", testData)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Test Get
	retrievedData, err := client.Get(context.Background(), "test-namespace", "test-key")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrievedData == nil {
		t.Error("Retrieved data is nil")
	}
}

func TestGetNonExistent(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)

	// Test Get on non-existent key
	_, err := client.Get(context.Background(), "test-namespace", "non-existent-key")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)

	// Put data first
	testData := "test value"
	err := client.Put(context.Background(), "test-namespace", "delete-key", testData)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Delete the data
	deleted, err := client.Delete(context.Background(), "test-namespace", "delete-key")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	if deleted != 1 {
		t.Errorf("Expected 1 deleted key, got: %d", deleted)
	}

	// Verify deletion
	_, err = client.Get(context.Background(), "test-namespace", "delete-key")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after deletion, got: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)

	// Put multiple items
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		err := client.Put(context.Background(), "test-getall", key, value)
		if err != nil {
			t.Fatalf("Failed to put data for key %s: %v", key, err)
		}
	}

	// Get all items
	allData, err := client.GetAll(context.Background(), "test-getall")
	if err != nil {
		t.Fatalf("Failed to get all data: %v", err)
	}

	if len(allData) != len(testData) {
		t.Errorf("Expected %d items, got %d", len(testData), len(allData))
	}
}

func TestInitializeNamespaces(t *testing.T) {
	_, etcdClient := newTestEtcd(t)

	client := NewClient(etcdClient)

	// Test that InitializeNamespaces doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitializeNamespaces panicked: %v", r)
		}
	}()

	InitializeNamespaces(client)
}