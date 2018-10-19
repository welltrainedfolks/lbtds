// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package main

import (
	"os"
	"os/signal"
	"syscall"

	"source.hodakov.me/fat0troll/lbtds/context"
	"source.hodakov.me/fat0troll/lbtds/domains/colors/v1"
	"source.hodakov.me/fat0troll/lbtds/domains/proxies/v1"
)

func main() {
	c := context.NewContext()
	c.Init()

	c.Logger.Info().Msgf("Starting LBTDS version %s", context.VERSION)
	c.InitConfiguration()
	c.InitAPIServer()

	colorsv1.Initialize(c)
	proxiesv1.Initialize(c)

	c.StartAPIServer()

	// CTRL+C handler.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)
	shutdownDone := make(chan bool, 1)
	go func() {
		signalThing := <-interrupt
		if signalThing == syscall.SIGTERM || signalThing == syscall.SIGINT {
			c.Logger.Info().Msg("Got " + signalThing.String() + " signal, shutting down...")
			c.SetShutdown()

			// TODO: actually shutdown proxy streams here

			c.Shutdown()
			shutdownDone <- true
		}
	}()

	<-shutdownDone
	c.Logger.Info().Msg("LBTDS instance shutted down")
	os.Exit(0)

}
