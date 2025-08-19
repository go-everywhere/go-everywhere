package views

import (
	"assette/widgets"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Home struct {
	app.Compo
}

func (h *Home) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Home page"),
	)
}
