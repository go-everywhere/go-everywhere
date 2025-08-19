//go:build !js

package main

import (
	"context"
	"log"
	"time"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
)

func database() (*olric.Olric, *olric.EmbeddedClient) {
	// local, lan, wan
	c := config.New("lan")

	// Callback function. It's called when this node is ready to accept connections.
	ctx, cancel := context.WithCancel(context.Background())
	c.Started = func() {
		defer cancel()
		log.Println("[INFO] Olric is ready to accept connections")
	}

	// Create a new Olric instance.
	db, err := olric.New(c)
	if err != nil {
		log.Fatalf("Failed to create Olric instance: %v", err)
	}

	// Start the instance. It will form a single-node cluster.
	go func() {
		// Call Start at background. It's a blocker call.
		err = db.Start()
		if err != nil {
			log.Fatalf("olric.Start returned an error: %v", err)
		}
	}()

	<-ctx.Done()
	return db, db.NewEmbeddedClient()
}

func shutdown(db *olric.Olric) {
	// Don't forget the call Shutdown when you want to leave the cluster.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.Shutdown(ctx)
	if err != nil {
		log.Printf("Failed to shutdown Olric: %v", err)
	}

	log.Println("[DONE] Olric DB shutdown")
}
