// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

// IsShuttingDown returns true if LBTDS is shutting down and false if not.
func (c *Context) IsShuttingDown() bool {
	c.inShutdownMutex.Lock()
	defer c.inShutdownMutex.Unlock()
	return c.inShutdown
}

// SetShutdown sets shutdown flag to true.
func (c *Context) SetShutdown() {
	c.inShutdownMutex.Lock()
	c.inShutdown = true
	c.inShutdownMutex.Unlock()
}

// Shutdown shutdowns context-related things.
func (c *Context) Shutdown() {
	c.ShutdownAPIServer()
	c.Logger.Info().Msg("Dropping PID file...")
	c.RemovePIDFile()
	c.Logger.Info().Msgf("LBTDS v. %s gracefully stopped. Have a nice day.", VERSION)
}
