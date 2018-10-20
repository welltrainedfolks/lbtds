// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package proxiesv1

import (
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

var (
	proxiesModuleLog zerolog.Logger

	// bunch of http.Server, which represents current proxy list
	httpProxies      []*fasthttp.Server
	httpProxiesMutex sync.Mutex
)

// HTTPProxy handles ServeHTTP function for passing data inside proxy
type HTTPProxy struct {
	Domain       string
	Destinations []string
}

func initProxies() {
	proxiesModuleLog = domainLog.With().Str("module", "proxies").Logger()
	proxiesModuleLog.Info().Msg("Initializing proxies...")
}

// startHTTPProxy starts proxy with desired configuration and adds it to proxies array
func startHTTPProxy(listenOn string, domain string, dst []string) {
	proxiesModuleLog.Debug().Msgf("Starting proxying on %s for domain %s to %s...", listenOn, domain, strings.Join(dst, ", "))

	srvHandler := newHTTPProxy(domain, dst)

	srv := &fasthttp.Server{
		Handler: srvHandler.HandleFastHTTP,
	}

	go func(addr string) {
		err := srv.ListenAndServe(listenOn)
		if err != nil {
			// It will always throw an error on graceful shutdown so it's
			// considered warning
			proxiesModuleLog.Warn().Err(err).Msgf("Proxy server on %s going down", addr)
		}
	}(listenOn)

	httpProxies = append(httpProxies, srv)
}

func newHTTPProxy(domain string, dst []string) *HTTPProxy {
	proxy := HTTPProxy{
		Domain:       domain,
		Destinations: dst,
	}
	return &proxy
}

// HandleFastHTTP handles requests for HTTPProxy
func (p *HTTPProxy) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	defer proxiesModuleLog.Info().Str("remote", ctx.RemoteAddr().String()).Str("domain", string(ctx.Host())).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")

	// Check if we have required domain in received request.
	if string(ctx.Host()) != p.Domain {
		ctx.Error("Invalid domain", fasthttp.StatusNotFound)
		return
	}

	destinationHost := p.Destinations[c.RandomSource.Intn(len(p.Destinations))]
	url := ctx.URI()
	url.SetHost(destinationHost)

	proxiesModuleLog.Debug().Msgf("Proxy request catched. Will go to %s", url.String())

	proxyRequest := ctx.Request
	proxyRequest.SetRequestURI(url.String())
	proxyRequest.Header.Set("Host", destinationHost)
	proxyRequest.Header.Set("X-Forwarded-For", ctx.RemoteAddr().String())

	proxyClient := &fasthttp.Client{}
	err := proxyClient.Do(&proxyRequest, &ctx.Response)
	if err != nil {
		proxiesModuleLog.Error().Err(err).Msg("Failed to deliver request")
		ctx.Error("Failed to deliver request", fasthttp.StatusBadGateway)
		return
	}
}
