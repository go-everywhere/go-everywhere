//go:build js

package main

import (
	"assette/views"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	app.Route("/", func() app.Composer { return &views.Home{} })
	app.Route("/profile", func() app.Composer { return &views.Profile{} })

	app.RunWhenOnBrowser()
}
