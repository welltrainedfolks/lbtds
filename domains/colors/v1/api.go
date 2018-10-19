// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

var (
	apiModuleLog zerolog.Logger
)

type colorRequestParams struct {
	Color string `json:"color"`
}

func initAPI() {
	apiModuleLog = domainLog.With().Str("module", "api").Logger()
	apiModuleLog.Info().Msg("Initializing API...")

	c.APIServerMux.HandleFunc("/api/v1/color/", ChangeColor)
}

// ChangeColor handles color changing for application context
func ChangeColor(w http.ResponseWriter, r *http.Request) {
	apiModuleLog.Debug().Msg("Received color change request")
	switch r.Method {
	case http.MethodPost:
		if r.Body == nil {
			http.Error(w, "Request body missing", 400)
			return
		}
		var requestParams colorRequestParams
		err := json.NewDecoder(r.Body).Decode(&requestParams)
		if err != nil {
			apiModuleLog.Error().Err(err).Msg("Failed to unmarshal POST data")
			http.Error(w, "Invalid request body", 400)
		}
		err = c.SetCurrentColor(requestParams.Color)
		if err != nil {
			http.Error(w, "Invalid color", 404)
		} else {
			http.Error(w, "Color changed", 200)
		}
	default:
		http.Error(w, "404 page not found", 404)
	}
}
