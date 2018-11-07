// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"encoding/json"
	"net/http"
	"time"

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
	start := time.Now()
	defer apiModuleLog.Info().Str("remote", r.RemoteAddr).TimeDiff("request time (s)", time.Now(), start).Msg("Received color switch HTTP request")
	switch r.Method {
	case http.MethodPost:
		var requestParams colorRequestParams
		err := json.NewDecoder(r.Body).Decode(&requestParams)
		if err != nil {
			apiModuleLog.Error().Err(err).Msg("Failed to unmarshal POST data")
			http.Error(w, "Invalid request body", 400)
			return
		}
		err = SetCurrentColor(requestParams.Color)
		if err != nil {
			http.Error(w, "Invalid color", 404)
		} else {
			http.Error(w, "Color changed", 200)
		}
	default:
		http.Error(w, "404 page not found", 404)
	}
}
