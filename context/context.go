// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/config"
)

// Init is an initialization function for core context
// Without these parts of the application we can't start at all
func (c *Context) Init() {
	c.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	c.Logger = c.Logger.Hook(zerolog.HookFunc(c.getMemoryUsage))

	c.inShutdownMutex.Lock()
	c.inShutdown = false
	c.inShutdownMutex.Unlock()

	c.RandomSource = rand.New(rand.NewSource(time.Now().Unix()))

	c.Logger.Info().Msgf("LBTDS v. %s is starting...", VERSION)
}

// InitConfiguration reads configuration from YAML and parses it in
// config.Struct.
func (c *Context) InitConfiguration() bool {
	c.Logger.Info().Msg("Loading configuration files...")

	configPath := os.Getenv("LBTDS_CONFIG")
	if configPath == "" {
		configPath = "./lbtds.yaml"
	}
	normalizedConfigPath, _ := filepath.Abs(configPath)
	c.Logger.Debug().Msgf("Configuration file path: %s", normalizedConfigPath)

	// Read configuration file into []byte.
	fileData, err := ioutil.ReadFile(normalizedConfigPath)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to read configuration file")
		return false
	}

	c.Config = &config.Struct{}
	err = yaml.Unmarshal(fileData, c.Config)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to parse configuration file")
		return false
	}

	c.Logger.Info().Msg("Configuration file parsed successfully")

	if len(c.Config.Colors) == 0 {
		c.Logger.Error().Err(err).Msg("There is no colors in configuration")
		return false
	}

	return true
}

// getMemoryUsage returns memory usage for logger.
func (c *Context) getMemoryUsage(e *zerolog.Event, level zerolog.Level, message string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	e.Str("memalloc", fmt.Sprintf("%dMB", m.Alloc/1024/1024))
	e.Str("memsys", fmt.Sprintf("%dMB", m.Sys/1024/1024))
	e.Str("numgc", fmt.Sprintf("%d", m.NumGC))

}
