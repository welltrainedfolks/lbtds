// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package main

import (
	"os"
	"os/signal"
	"syscall"

	"lab.wtfteam.pro/wtfteam/lbtds/context"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/colors/v1"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/proxies/v1"
)

func main() {
	c := context.NewContext()
	c.Init()

	c.InitConfiguration()
	c.InitAPIServer()

	colorsv1.Initialize(c)
	proxiesv1.Initialize(c)

	c.StartAPIServer()

	colorsv1.GetCurrentColor()

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
	os.Exit(0)
}
