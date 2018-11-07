// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"lab.wtfteam.pro/wtfteam/lbtds/context"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/colors/v1"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/proxies/v1"
)

func checkStartupState(goodStartupState bool) {
	if !goodStartupState {
		panic("LBTDS stopped due to unrecoverable errors")
	}
}

func main() {
	// Before any real work - lock to OS thread. We shouldn't leave it until
	// shutdown
	runtime.LockOSThread()

	// And here is the rock'n'roll starts
	c := context.NewContext()
	c.Init()

	checkStartupState(c.InitConfiguration())
	checkStartupState(c.CheckPIDFile())
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
			c.Logger.Info().Msg("Shutting down proxy streams...")
			proxiesv1.Shutdown()
			c.Shutdown()
			shutdownDone <- true
		}
	}()

	<-shutdownDone
	os.Exit(0)
}
