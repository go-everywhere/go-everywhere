package views

import (
	"assette/widgets"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/olric-data/olric"
)

var _ app.Mounter = (*Home)(nil)

type Home struct {
	app.Compo
	DB *olric.EmbeddedClient
}

func (h *Home) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Home page"),
	)
}

func (h *Home) OnMount(ctx app.Context) {

}
