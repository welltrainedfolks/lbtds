// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package context

import (
	"math/rand"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
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

	// Random source
	// Needed for picking random exit for proxy
	RandomSource *rand.Rand

	// Are we shutting down?
	inShutdown bool
	// Even bools aren't goroutine-safe!
	inShutdownMutex sync.Mutex
}

// NewContext is an initialization function for Context
func NewContext() *Context {
	c := &Context{}
	return c
}
