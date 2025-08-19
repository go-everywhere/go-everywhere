package views

import (
	"assette/widgets"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Generate struct {
	app.Compo
}

func (h *Generate) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Generate page"),
	)
}
