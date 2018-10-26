// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	ctx "context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/config"
)

// VERSION is our current version
const VERSION = "0.1.0"

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

	c.Logger.Info().Msgf("LBTDS v. %s is starting...", VERSION)
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

	c.CheckPIDFile()

	c.GetCurrentColor()
}

// CheckPIDFile checks for existing PID file and creates one if possible
func (c *Context) CheckPIDFile() {
	normalizedPIDFilePath := c.getPIDFilePath()

	c.Logger.Debug().Msgf("PID file path: %s", normalizedPIDFilePath)

	// We need to fail here, not pass
	currentPID, err := ioutil.ReadFile(normalizedPIDFilePath)
	if err != nil {
		// Write PID to new file
		newPIDfile, err := os.OpenFile(normalizedPIDFilePath, os.O_RDWR|os.O_CREATE, 0755)
		defer newPIDfile.Close()
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to create PID file")
		}

		_, err = newPIDfile.Write([]byte(strconv.Itoa(os.Getpid())))
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to write PID file")
		}
	} else {
		// PID file exists
		c.Logger.Panic().Msgf("There is already LBTDS instance with the same configuration running at PID %s. Stop it or remove PID file if instance already stopped.", string(currentPID))
	}
}

// RemovePIDFile removes PID file on stop
func (c *Context) RemovePIDFile() {
	normalizedPIDFilePath := c.getPIDFilePath()

	err := os.Remove(normalizedPIDFilePath)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to remove PID file")
	}
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
					err = c.SetCurrentColor(color)
					if err != nil {
						c.Logger.Warn().Err(err).Msgf("Failed to change color to %s", color)
					}
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
	defer c.currentColorMutex.Unlock()
	if c.Config.Colors[color] != nil {
		c.currentColor = color

		normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
		if err != nil {
			c.Logger.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
		}

		colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, 0755)
		defer colorsFile.Close()
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to open current color file or create one")
		}
		err = colorsFile.Truncate(0)
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to truncate current color file")
		}
		_, err = colorsFile.Write([]byte(color))
		if err != nil {
			c.Logger.Warn().Err(err).Msg("Failed to write current color to file")
		}

		c.Logger.Info().Msgf("Current color changed to %s", c.currentColor)

		c.ColorChanged <- true
	} else {
		c.Logger.Warn().Msgf("There is no such color in configuration: %s", color)
		err = errors.New("Invalid color name")
	}

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
	err := c.APIServer.Shutdown(ctx.TODO())
	if err != nil {
		c.Logger.Error().Msgf("Failed to shutdown API server: %s", err.Error())
	}
	c.Logger.Info().Msg("Shutting down proxy streams...")
	// TODO: Make it shut down proxy streams
	c.Logger.Info().Msg("Dropping PID file...")
	c.RemovePIDFile()
	c.Logger.Info().Msgf("LBTDS v. %s gracefully stopped. Have a nice day.", VERSION)
}

// getMemoryUsage returns memory usage for logger.
func (c *Context) getMemoryUsage(e *zerolog.Event, level zerolog.Level, message string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	e.Str("memalloc", fmt.Sprintf("%dMB", m.Alloc/1024/1024))
	e.Str("memsys", fmt.Sprintf("%dMB", m.Sys/1024/1024))
	e.Str("numgc", fmt.Sprintf("%d", m.NumGC))

}

// getPIDFilePath returns PID file path
func (c *Context) getPIDFilePath() string {
	var pidFile string
	if c.Config.Proxy.PIDFile != "" {
		pidFile = c.Config.Proxy.PIDFile
	} else {
		switch runtime.GOOS {
		case "windows":
			c.Logger.Panic().Msg("LBTDS doesn't support Windows at this time. Please, read CONTRIBUTING.md for adding Windows support if you're interested in it.")
		case "darwin":
			pidFile = "/usr/local/var/run/lbtds.pid"
		case "linux":
			pidFile = "/var/run/lbtds.pid"
		default:
			pidFile = "/var/run/lbtds.pid"
		}
	}

	normalizedPIDFilePath, err := filepath.Abs(pidFile)
	if err != nil {
		c.Logger.Panic().Err(err).Msgf("Failed to normalize PID file path. Path supplied: %s", pidFile)
	}

	return normalizedPIDFilePath
}
