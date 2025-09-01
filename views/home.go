package views

import (
	"assette/widgets"
	"encoding/json"
	"net/http"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

var _ app.Mounter = (*Home)(nil)

type Home struct {
	app.Compo
	message string
}

func (h *Home) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Home page"),
		app.P().Text(h.message),
	)
}

func (h *Home) OnMount(ctx app.Context) {
	ctx.Async(func() {
		resp, err := http.Get("/api/message")
		if err != nil {
			app.Log(err)
			return
		}
		defer resp.Body.Close()

		var data struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			app.Log(err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			h.message = data.Text
		})
	})
}
