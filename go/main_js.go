//go:build js

package main

import (
	"assette/views"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	app.Route("/", func() app.Composer { return &views.Home{} })
	app.Route("/generate", func() app.Composer { return &views.Generate{} })
	app.Route("/models", func() app.Composer { return &views.Models{} })

	app.RunWhenOnBrowser()
}
