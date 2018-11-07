// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	ctx "context"
	"net/http"
	"time"
)

// InitAPIServer initializes API server mux
func (c *Context) InitAPIServer() {
	listenAddress := c.Config.API.Address + ":" + c.Config.API.Port
	c.APIServer = &http.Server{
		Addr: listenAddress,
	}
	c.APIServerMux = http.NewServeMux()
}

// StartAPIServer starts API server for listening
func (c *Context) StartAPIServer() {
	listenAddress := c.Config.API.Address + ":" + c.Config.API.Port
	c.Logger.Info().Msg("Starting API server on http://" + listenAddress)

	c.APIServer.Handler = c.APIServerMux
	go func() {
		err := c.APIServer.ListenAndServe()
		// It will always throw an error on graceful shutdown so it's considered
		// as warning
		if err != nil {
			c.Logger.Warn().Err(err).Msgf("API server on http://%s gone down", listenAddress)
		}
	}()

	count := 0
	for {
		if count >= 5 {
			c.Logger.Error().Msg("API server failed to start listening!")
			break
		}

		err := c.checkAPIHealth()
		if err != nil {
			c.Logger.Warn().Msgf("API server is not ready, delaying (error: %s)...", err.Error())
			time.Sleep(time.Second * 1)
			count++
			continue
		}

		c.Logger.Info().Msgf("API server is up")
		c.APIServerUp = true
		break
	}
}

// ShutdownAPIServer stops API server
func (c *Context) ShutdownAPIServer() {
	c.Logger.Info().Msg("Shutting down API server...")
	closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer closedownCancel()
	err := c.APIServer.Shutdown(closedownContext)
	if err != nil {
		c.Logger.Error().Msgf("Failed to shutdown API server: %s", err.Error())
	}
	c.APIServerUp = false
}

// checkAPIHealth sends request to API server
func (c *Context) checkAPIHealth() error {
	listenAddress := c.Config.API.Address + ":" + c.Config.API.Port
	req, err := http.NewRequest("GET", "http://"+listenAddress+"/nonexistent/", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		c.Logger.Error().Msgf("Failed to create request structure: %s", err.Error())
	}

	client := &http.Client{Timeout: time.Second * 1}
	_, err = client.Do(req)
	return err
}
