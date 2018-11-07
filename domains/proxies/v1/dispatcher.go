// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	ctx "context"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/colors/v1"
)

var (
	dispatcherModuleLog zerolog.Logger
)

func initDispatcher() {
	// Dispatcher shouldn't lose his child routines, we need to lock OS thread
	// for stability
	runtime.LockOSThread()

	dispatcherModuleLog = domainLog.With().Str("module", "dispatcher").Logger()
	dispatcherModuleLog.Info().Msg("Initializing proxies dispatcher...")

	go func() {
		awaitColorChanged()
	}()
}

// awaitColorChanged listens to channel which fires up when color actually changes
func awaitColorChanged() {
	// First call of dispatchChange runs at start of the balancer
	for <-colorsv1.ColorChanged {
		dispatchChange()
	}
}

// dispatchChanges restarts the whole bunch of proxies to new color scheme
func dispatchChange() {
	dispatcherModuleLog.Debug().Msgf("Color %s selected. Starting proxies...", colorsv1.GetCurrentColorName())

	if len(httpProxies) > 0 {
		Shutdown()
	}

	for _, proxy := range colorsv1.GetCurrentColorConfiguration().Backends {
		startHTTPProxy(proxy.ListenOn, proxy.Source, proxy.Destinations)
	}
}

// Shutdown shutdowns all proxies (useful on graceful shutdown)
func Shutdown() {
	httpProxiesMutex.Lock()
	defer httpProxiesMutex.Unlock()
	for i, proxy := range httpProxies {
		dispatcherModuleLog.Debug().Msgf("Stopping proxy on %s...", proxy.Addr)
		closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
		defer closedownCancel()
		err := proxy.Shutdown(closedownContext)
		if err != nil {
			dispatcherModuleLog.Error().Err(err).Msg("Failed to shut down proxy")
		}
		delete(httpProxies, i)
	}
}
