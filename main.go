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
	embeddedEtcd, etcdClient, client := database()

	app.Route("/", func() app.Composer { return &views.Home{} })
	app.Route("/profile", func() app.Composer { return &views.Profile{} })

	http.HandleFunc("/api/users", api.UserRouter(client))
	http.HandleFunc("/api/users/", api.UserRouter(client))
	http.HandleFunc("/api/message", api.GetMessage())
	http.Handle("/", &app.Handler{
		Name:        "Go PWA",
		Description: "A Go PWA template",
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(":8000", nil); err != nil {
			log.Fatal(err)
		}
	}()

	<-signalChan
	shutdown(embeddedEtcd, etcdClient)
}
