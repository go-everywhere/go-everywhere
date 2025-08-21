package views

import (
	"assette/widgets"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/olric-data/olric"
)

type Profile struct {
	app.Compo
	DB *olric.EmbeddedClient
}

func (h *Profile) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Profile page"),
	)
}
