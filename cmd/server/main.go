package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jairo/assetter/internal/api"
	"github.com/jairo/assetter/internal/storage"
	"github.com/jairo/assetter/internal/stability"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	apiKey := os.Getenv("STABILITY_API_KEY")
	if apiKey == "" {
		log.Fatal("STABILITY_API_KEY environment variable is required")
	}

	fileStore := storage.NewLocalStorage("uploads")
	stabilityClient := stability.NewClient(apiKey)
	
	handler := api.NewHandler(fileStore, stabilityClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomePage)
	mux.HandleFunc("/upload", handler.Upload)
	mux.HandleFunc("/status/", handler.Status)
	mux.HandleFunc("/download/", handler.Download)
	mux.HandleFunc("/api/models", handler.ListModels)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
