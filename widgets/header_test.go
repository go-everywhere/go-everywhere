package widgets

import (
	"testing"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func TestHeaderRender(t *testing.T) {
	header := &Header{}

	ui := header.Render()
	if ui == nil {
		t.Error("Header.Render() returned nil")
	}

	headerElem, ok := ui.(app.HTMLHeader)
	if !ok {
		t.Error("Header.Render() should return app.HTMLHeader")
	}
	_ = headerElem
}

func TestHeaderIsCompo(t *testing.T) {
	header := &Header{}
	
	// Verify that Header properly embeds app.Compo
	var _ app.UI = header
}