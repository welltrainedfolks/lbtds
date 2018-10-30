// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	ctx "context"
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

	"github.com/pztrn/flagger"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/config"
)

// VERSION is our current version
const VERSION = "0.1.0"

// Context is the main application context. This struct handles operations
// between all parts of the application
type Context struct {
	Config  *config.Struct
	Flagger *flagger.Flagger
	Logger  zerolog.Logger

	// API server
	APIServer    *http.Server
	APIServerMux *http.ServeMux
	APIServerUp  bool

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

	c.RandomSource = rand.New(rand.NewSource(time.Now().Unix()))

	c.Flagger = flagger.New(nil)
	c.Flagger.Initialize()
	err := c.Flagger.AddFlag(&flagger.Flag{
		Name:         "config",
		Description:  "Path to configuration file, including filename. Can be overrided with flag --config.",
		Type:         "string",
		DefaultValue: "./lbtds.yaml",
	})
	if err != nil {
		c.Logger.Panic().Err(err).Msg("Failed to add flag to parse")
	}
	c.Flagger.Parse()

	c.Logger.Info().Msgf("LBTDS v. %s is starting...", VERSION)
}

// InitConfiguration reads configuration from YAML and parses it in
// config.Struct.
func (c *Context) InitConfiguration() {
	c.Logger.Info().Msg("Loading configuration files...")

	configPath, err := c.Flagger.GetStringValue("config")
	if err != nil {
		c.Logger.Panic().Msg("Failed to read config file path from flag")
	}
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
		if err != nil {
			c.Logger.Panic().Err(err).Msg("Failed to create PID file")
		}
		defer newPIDfile.Close()

		_, err = newPIDfile.Write([]byte(strconv.Itoa(os.Getppid())))
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

	parentPID, err := ioutil.ReadFile(normalizedPIDFilePath)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to read PID file")
	}

	if string(parentPID) != strconv.Itoa(os.Getppid()) {
		c.Logger.Error().Err(err).Msgf("PID file contains wrong PID: expected %d, but got %s", os.Getppid(), parentPID)
	}

	err = os.Remove(normalizedPIDFilePath)
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
			pidFile = "/usr/local/var/run/lbtds.lock"
		case "linux":
			pidFile = "/var/run/lbtds.lock"
		default:
			pidFile = "/var/run/lbtds.lock"
		}
	}

	normalizedPIDFilePath, err := filepath.Abs(pidFile)
	if err != nil {
		c.Logger.Panic().Err(err).Msgf("Failed to normalize PID file path. Path supplied: %s", pidFile)
	}

	return normalizedPIDFilePath
}
