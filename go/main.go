//go:build !js

package main

import (
	"assette/api"
	"assette/views"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	app.Route("/", func() app.Composer { return &views.Home{} })
	app.Route("/generate", func() app.Composer { return &views.Generate{} })
	app.Route("/models", func() app.Composer { return &views.Models{} })

	db, client := database()

	http.HandleFunc("/api/generate", api.Generate(client))
	http.Handle("/", &app.Handler{
		Name:        "Assette",
		Description: "An asset creator for Indies",
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(":8000", nil); err != nil {
			log.Fatal(err)
		}
	}()

	<-signalChan
	shutdown(db)
}
