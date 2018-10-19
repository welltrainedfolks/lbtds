// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"github.com/rs/zerolog"
	"source.hodakov.me/fat0troll/lbtds/context"
)

var (
	c *context.Context

	// Package-wide logger, with "domain" parameter defined
	domainLog zerolog.Logger
)

// Initialize initializes package
func Initialize(cc *context.Context) {
	c = cc
	domainLog = c.Logger.With().Str("domain", "colors").Int("version", 1).Logger()

	initAPI()

	domainLog.Info().Msg("Domain «colors» initialized")
}
