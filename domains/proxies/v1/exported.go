// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	"github.com/rs/zerolog"
	"lab.wtfteam.pro/wtfteam/lbtds/context"
)

var (
	c *context.Context

	// Package-wide logger, with "domain" parameter defined
	domainLog zerolog.Logger
)

// Initialize initializes package
func Initialize(cc *context.Context) {
	c = cc
	domainLog = c.Logger.With().Str("domain", "proxies").Int("version", 1).Logger()

	initProxies()
	initDispatcher()

	domainLog.Info().Msg("Domain «proxies» initialized")
}
