// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	"github.com/rs/zerolog"
)

var (
	dispatcherModuleLog zerolog.Logger
)

func initDispatcher() {
	dispatcherModuleLog = domainLog.With().Str("module", "dispatcher").Logger()
	dispatcherModuleLog.Info().Msg("Initializing proxies dispatcher...")
	go func() {
		awaitColorChanged()
	}()
}

// awaitColorChanged listens to channel which fires up when color actually changes
func awaitColorChanged() {
	// First call of dispatchChange runs at start of the balancer
	dispatchChange()
	for <-c.ColorChanged {
		dispatchChange()
	}
}

// dispatchChanges restarts the whole bunch of proxies to new color scheme
func dispatchChange() {
	dispatcherModuleLog.Debug().Msg("Color selected. Starting proxies...")

	for i := range proxies {
		dispatcherModuleLog.Debug().Msgf("Stopping proxy on %s...", proxies[i].Addr)
		proxies[i].Shutdown(nil)
	}

	for _, proxy := range c.Config.Colors[c.GetCurrentColor()] {
		startProxy(proxy.Source, proxy.Destinations)
	}
}
