// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	proxiesModuleLog zerolog.Logger

	// bunch of http.Server, which represents current proxy list
	httpProxies      []*http.Server
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

	srv := &http.Server{
		Addr:    listenOn,
		Handler: newHTTPProxy(domain, dst),
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			// It will always throw an error on graceful shutdown so it's
			// considered warning
			proxiesModuleLog.Warn().Err(err).Msgf("Proxy server on %s going down", srv.Addr)
		}
	}()

	httpProxies = append(httpProxies, srv)
}

func newHTTPProxy(domain string, dst []string) *HTTPProxy {
	proxy := HTTPProxy{
		Domain:       domain,
		Destinations: dst,
	}
	return &proxy
}

func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	start := time.Now()
	defer proxiesModuleLog.Info().Str("remote", r.RemoteAddr).Str("domain", r.Host).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")

	// Check if we have required domain in received request.
	if r.Host != p.Domain {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

	url := r.URL
	url.Host = p.Destinations[c.RandomSource.Intn(len(p.Destinations))]
	url.Scheme = "http"

	proxiesModuleLog.Debug().Msgf("Proxy request catched. Will go to %s", url.String())

	proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}

	proxyReq.Header.Set("Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	client := &http.Client{}
	proxyRsp, err := client.Do(proxyReq)
	if err != nil {
		proxiesModuleLog.Error().Err(err).Msg("Can't connect to downstream")
		http.Error(w, "Can't connect to downstream", 502)
		return
	}

	for header, values := range proxyRsp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	_, err = io.Copy(w, proxyRsp.Body)
	if err != nil {
		proxiesModuleLog.Error().Err(err).Msg("Can't write response to upstream")
		http.Error(w, "Can't write response to upstream", 502)
		return
	}
	proxyRsp.Body.Close()
}
