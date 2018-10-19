// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"source.hodakov.me/fat0troll/lbtds/internal/config"
)

// VERSION is our current version
const VERSION = "0.0.1"

// Context is the main application context. This struct handles operations
// between all parts of the application
type Context struct {
	Config *config.Struct
	Logger zerolog.Logger

	// Current color
	currentColor      string
	currentColorMutex sync.Mutex

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

	normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
	if err != nil {
		c.Logger.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
	}
	c.Logger.Debug().Msgf("Current color file path: %s", normalizedColorsPath)

	colorsData, err := ioutil.ReadFile(normalizedColorsPath)
	if err != nil {
		c.Logger.Panic().Err(err).Msg("Failed to read current color file")
	}

	c.SetCurrentColor(string(colorsData))
}

// IsShuttingDown returns true if LBTDS is shutting down and false if not.
func (c *Context) IsShuttingDown() bool {
	c.inShutdownMutex.Lock()
	defer c.inShutdownMutex.Unlock()
	return c.inShutdown
}

// GetCurrentColor gets current color for application
func (c *Context) GetCurrentColor() string {
	c.currentColorMutex.Lock()
	defer c.currentColorMutex.Unlock()
	return c.currentColor
}

// SetCurrentColor sets current color for application
func (c *Context) SetCurrentColor(color string) {
	c.currentColorMutex.Lock()
	c.currentColor = color
	c.currentColorMutex.Unlock()
	c.Logger.Info().Msgf("Current color changed to %s", c.currentColor)
}

// SetShutdown sets shutdown flag to true.
func (c *Context) SetShutdown() {
	c.inShutdownMutex.Lock()
	c.inShutdown = true
	c.inShutdownMutex.Unlock()
}

// Shutdown shutdowns context-related things.
func (c *Context) Shutdown() {
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
