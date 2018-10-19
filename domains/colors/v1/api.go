// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

var (
	apiModuleLog zerolog.Logger
)

func initAPI() {
	apiModuleLog = domainLog.With().Str("module", "api").Logger()
	apiModuleLog.Info().Msg("Initializing API...")

	c.APIServerMux.HandleFunc("/api/v1/color", ChangeColor)
}

// ChangeColor handles color changing for application context
func ChangeColor(w http.ResponseWriter, r *http.Request) {
	apiModuleLog.Debug().Msg("Received color change request")
	switch r.Method {
	case http.MethodPost:
		fmt.Println(r.Body)
	default:
		w.WriteHeader(502)
	}
}
