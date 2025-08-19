package widgets

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Header struct {
	app.Compo
}

func (h *Header) Render() app.UI {
	return app.Header().Body(
		app.A().Href("/").Text("Home"),
		app.A().Href("/generate").Text("Generate"),
		app.A().Href("/models").Text("Models"),
	)
}
