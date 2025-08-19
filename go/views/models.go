package views

import (
	"assette/widgets"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Models struct {
	app.Compo
}

func (h *Models) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Models page"),
	)
}
