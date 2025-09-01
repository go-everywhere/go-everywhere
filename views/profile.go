package views

import (
	"assette/models"
	"assette/widgets"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Profile struct {
	app.Compo
	user models.User
}

func (p *Profile) Render() app.UI {
	return app.Section().Body(
		&widgets.Header{},
		app.H1().Text("Profile page"),
		app.Form().OnSubmit(p.handleSubmit).Body(
			app.Input().
				Type("text").
				Value(p.user.Name).
				Placeholder("Name").
				OnInput(p.ValueTo(&p.user.Name)),
			app.Input().
				Type("email").
				Value(p.user.Email).
				Placeholder("Email").
				OnInput(p.ValueTo(&p.user.Email)),
			app.Button().
				Type("submit").
				Text("Update Profile"),
		),
	)
}

func (p *Profile) handleSubmit(ctx app.Context, e app.Event) {
	e.PreventDefault()

	ctx.Async(func() {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(p.user); err != nil {
			app.Log(err)
			return
		}

		// For demo, we'll create a new user
		// In a real app, you'd track the user ID and use PUT for updates
		resp, err := http.Post("/api/users", "application/json", &buf)
		if err != nil {
			app.Log(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			app.Log("failed to save user profile")
		}
	})
}
