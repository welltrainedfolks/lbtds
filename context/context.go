// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"source.hodakov.me/fat0troll/lbtds/internal/config"
)

// VERSION is our current version
const VERSION = "0.0.2"

// Context is the main application context. This struct handles operations
// between all parts of the application
type Context struct {
	Config *config.Struct
	Logger zerolog.Logger

	// API server
	APIServer    *http.Server
	APIServerMux *http.ServeMux
	APIServerUp  bool

	// Current color
	currentColor      string
	currentColorMutex sync.Mutex

	// Color-changing channel
	// There will be signal on each color change
	ColorChanged chan bool

	// Random source
	// Needed for picking random exit for proxy
	RandomSource *rand.Rand

	// Are we shutting down?
	inShutdown bool
	// Even bools aren't goroutine-safe!
	inShutdownMutex sync.Mutex
}

// Init is an initialization function for core context
// Without these parts of the application we can't start at all
func (c *Context) Init() {
	c.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	c.Logger = c.Logger.Hook(zerolog.HookFunc(c.getMemoryUsage))

	c.inShutdownMutex.Lock()
	c.inShutdown = false
	c.inShutdownMutex.Unlock()

	c.ColorChanged = make(chan bool)
	c.RandomSource = rand.New(rand.NewSource(time.Now().Unix()))
}

// InitConfiguration reads configuration from YAML and parses it in
// config.Struct.
func (c *Context) InitConfiguration() {
	c.Logger.Info().Msg("Loading configuration files...")

	// TODO: make it flaggable
	configPath := "./lbtds.yaml"
	normalizedConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		c.Logger.Panic().Msgf("Failed to normalize configuration path. Path supplied: '%s'", configPath)
	}
	c.Logger.Debug().Msgf("Configuration file path: %s", normalizedConfigPath)

	// Read configuration file into []byte.
	fileData, err := ioutil.ReadFile(normalizedConfigPath)
	if err != nil {
		c.Logger.Panic().Err(err).Msg("Failed to read configuration file")
	}

	c.Config = &config.Struct{}
	err = yaml.Unmarshal(fileData, c.Config)
	if err != nil {
		c.Logger.Panic().Err(err).Msg("Failed to parse configuration file")
	}

	c.Logger.Info().Msg("Configuration file parsed successfully")

	if len(c.Config.Colors) == 0 {
		c.Logger.Panic().Err(err).Msg("There is no colors in configuration")
	}

	c.GetCurrentColor()
}

// InitAPIServer initializes API server mux
func (c *Context) InitAPIServer() {
	listenAddress := c.Config.API.Address + ":" + c.Config.API.Port
	c.APIServer = &http.Server{
		Addr: listenAddress,
	}
	c.APIServerMux = http.NewServeMux()
}

// IsShuttingDown returns true if LBTDS is shutting down and false if not.
func (c *Context) IsShuttingDown() bool {
	c.inShutdownMutex.Lock()
	defer c.inShutdownMutex.Unlock()
	return c.inShutdown
}

// GetCurrentColor gets current color for application
func (c *Context) GetCurrentColor() string {
	if c.currentColor == "" {
		normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
		if err != nil {
			c.Logger.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
		}
		c.Logger.Debug().Msgf("Current color file path: %s", normalizedColorsPath)

		colorsData, err := ioutil.ReadFile(normalizedColorsPath)
		if err != nil {
			idx := 0
			for color := range c.Config.Colors {
				if idx == 0 {
					c.SetCurrentColor(color)
				}
				idx++
			}
		} else {
			c.currentColor = string(colorsData)
		}
	}
	return c.currentColor
}

// SetCurrentColor sets current color for application
func (c *Context) SetCurrentColor(color string) error {
	var err error
	c.currentColorMutex.Lock()
	if c.Config.Colors[color] != nil {
		c.currentColor = color

		normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
		if err != nil {
			c.Logger.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
		}

		colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to open current color file or create one")
		}
		colorsFile.Truncate(0)
		colorsFile.Write([]byte(color))
		colorsFile.Close()

		c.Logger.Info().Msgf("Current color changed to %s", c.currentColor)

		c.ColorChanged <- true
	} else {
		c.Logger.Warn().Msgf("There is no such color in configuration: %s", color)
		err = errors.New("Invalid color name")
	}

	c.currentColorMutex.Unlock()
	return err
}

// SetShutdown sets shutdown flag to true.
func (c *Context) SetShutdown() {
	c.inShutdownMutex.Lock()
	c.inShutdown = true
	c.inShutdownMutex.Unlock()
}

// StartAPIServer starts API server for listening
func (c *Context) StartAPIServer() {
	listenAddress := c.Config.API.Address + ":" + c.Config.API.Port
	c.Logger.Info().Msg("Starting API server on http://" + listenAddress)

	c.APIServer.Handler = c.APIServerMux
	go func() {
		c.APIServer.ListenAndServe()
	}()

	count := 0
	for {
		if count >= 5 {
			c.Logger.Error().Msg("API server failed to start listening!")
			break
		}

		req, err := http.NewRequest("GET", "http://"+listenAddress+"/notexistant/", nil)
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		if err != nil {
			c.Logger.Error().Msgf("Failed to create request structure: %s", err.Error())
		}

		client := &http.Client{Timeout: time.Second * 1}
		_, err = client.Do(req)
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

// Shutdown shutdowns context-related things.
func (c *Context) Shutdown() {
	c.Logger.Info().Msg("Shutting down API server...")
	err := c.APIServer.Shutdown(nil)
	if err != nil {
		c.Logger.Error().Msgf("Failed to shutdown API server: %s", err.Error())
	}
	c.Logger.Info().Msg("Shutting down proxy streams...")
	// TODO: Make it shut down proxy streams
}

// getMemoryUsage returns memory usage for logger.
func (c *Context) getMemoryUsage(e *zerolog.Event, level zerolog.Level, message string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	e.Str("memalloc", fmt.Sprintf("%dMB", m.Alloc/1024/1024))
	e.Str("memsys", fmt.Sprintf("%dMB", m.Sys/1024/1024))
	e.Str("numgc", fmt.Sprintf("%d", m.NumGC))

}
