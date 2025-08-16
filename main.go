package main

import (
	"assette/views"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	// If run with "go run" compile the frontend
	if strings.Contains(os.Args[0], "go-build") {
		c := exec.Command("go", "build", "-o", "web/app.wasm")
		goenv, err := os.ReadFile(".goenv")
		if err != nil {
			log.Fatalf("Error reading .goenv, run 'go env > .goenv' %v", err)
		}
		c.Env = append(c.Env, "GOARCH=wasm", "GOOS=js")
		c.Env = strings.Split(strings.ReplaceAll(string(goenv), "'", ""), "\n")
		if res, err := c.CombinedOutput(); err != nil {
			log.Fatalf("Error compiling frontend: %v %v", err, string(res))
		}
	}

	// Steps
	// 1. Global Style + Item prompt
	// 2. Stability 3D
	// 3. Tetra3D to isometric/top down/platformer

	app.Route("/", func() app.Composer { return &views.Home{} })

	if app.IsClient {
		app.RunWhenOnBrowser()
		return
	}
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
	})

	log.Println("Listening on http://localhost:8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
