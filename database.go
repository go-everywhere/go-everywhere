//go:build !js

package main

import (
	"assette/db"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func database() (*embed.Etcd, *clientv3.Client, *db.Client) {
	// Create data directory for etcd
	dataDir := filepath.Join(os.TempDir(), "etcd-data")
	os.RemoveAll(dataDir) // Clean up any previous data

	cfg := embed.NewConfig()
	cfg.Dir = dataDir
	cfg.LogLevel = "error" // Reduce log verbosity

	// Configure listening URLs
	lcurl, _ := url.Parse("http://127.0.0.1:2379")
	cfg.ListenClientUrls = []url.URL{*lcurl}
	cfg.AdvertiseClientUrls = []url.URL{*lcurl}

	lpurl, _ := url.Parse("http://127.0.0.1:2380")
	cfg.ListenPeerUrls = []url.URL{*lpurl}
	cfg.AdvertisePeerUrls = []url.URL{*lpurl}

	cfg.InitialCluster = fmt.Sprintf("default=%s", lpurl.String())

	// Disable strict reconfiguration check
	cfg.StrictReconfigCheck = false

	// Start embedded etcd
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatalf("Failed to start embedded etcd: %v", err)
	}

	// Wait for etcd to be ready
	select {
	case <-e.Server.ReadyNotify():
		log.Println("[INFO] Embedded etcd is ready to accept connections")
	case <-time.After(10 * time.Second):
		e.Server.Stop()
		log.Fatalf("Embedded etcd took too long to start")
	}

	// Create client connection to embedded etcd
	clientConfig := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	etcdClient, err := clientv3.New(clientConfig)
	if err != nil {
		e.Server.Stop()
		log.Fatalf("Failed to create etcd client: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = etcdClient.Get(ctx, "health-check")
	if err != nil {
		log.Printf("[WARNING] etcd connection test failed: %v", err)
	}

	// Initialize namespaces
	client := db.NewClient(etcdClient)
	db.InitializeNamespaces(client)

	return e, etcdClient, client
}

func shutdown(e *embed.Etcd, client *clientv3.Client) {
	// Close client connection first
	if client != nil {
		err := client.Close()
		if err != nil {
			log.Printf("Failed to close etcd client: %v", err)
		}
	}

	// Stop embedded etcd server
	if e != nil {
		e.Server.Stop()
		e.Close()

		// Clean up data directory
		cfg := e.Config()
		if cfg.Dir != "" {
			os.RemoveAll(cfg.Dir)
		}
	}

	log.Println("[DONE] Embedded etcd shutdown")
}