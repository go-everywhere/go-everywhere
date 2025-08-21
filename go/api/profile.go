//go:build !js

package api

import (
	"net/http"

	"github.com/olric-data/olric"
)

func UpdateProfile(db *olric.EmbeddedClient) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	}
}
