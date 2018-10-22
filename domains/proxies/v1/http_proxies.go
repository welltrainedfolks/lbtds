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
	start := time.Now()
	// ToDo: strict or not strict domain forwarding. For now we will
	// forward only domain name, without port.
	domainToForward := strings.Split(r.Host, ":")[0]
	var proxifiedBytesCount int64
	var responseCode int

	// Check if we have required domain in received request.
	if domainToForward != p.Domain {
		proxiesModuleLog.Error().Str("domain", domainToForward).Msg("Invalid domain passed")
		responseCode = http.StatusBadRequest
		http.Error(w, "Invalid domain", responseCode)
		proxiesModuleLog.Info().Str("remote", r.RemoteAddr).Str("domain", domainToForward).Int("code", responseCode).Int64("proxified bytes", proxifiedBytesCount).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")
		return
	}

	url := r.URL
	url.Host = p.Destinations[c.RandomSource.Intn(len(p.Destinations))]
	url.Scheme = "http"

	proxiesModuleLog.Debug().Str("domain", domainToForward).Msgf("Proxy request catched. Will go to %s", url.String())

	proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
	if err != nil {
		proxiesModuleLog.Error().Str("domain", domainToForward).Err(err).Msg("Failed to create new HTTP request to downstream")
		responseCode = http.StatusInternalServerError
		http.Error(w, "Internal error", responseCode)
		proxiesModuleLog.Info().Str("remote", r.RemoteAddr).Str("domain", domainToForward).Int("code", responseCode).Int64("proxified bytes", proxifiedBytesCount).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")
		return
	}

	//proxyReq.Header.Set("Host", domainToForward)
	proxyReq.Host = domainToForward
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	client := &http.Client{}
	proxyRsp, err := client.Do(proxyReq)
	if err != nil {
		proxiesModuleLog.Error().Str("domain", domainToForward).Err(err).Msg("Can't connect to downstream")
		http.Error(w, "Can't connect to downstream", responseCode)
		proxiesModuleLog.Info().Str("remote", r.RemoteAddr).Str("domain", domainToForward).Int("code", responseCode).Int64("proxified bytes", proxifiedBytesCount).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")
		return
	}

	for header, values := range proxyRsp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	proxifiedBytesCount, err = io.Copy(w, proxyRsp.Body)
	if err != nil {
		proxiesModuleLog.Error().Err(err).Msg("Can't write response to upstream")
		http.Error(w, "Can't write response to upstream", 502)
		return
	}

	proxiesModuleLog.Info().Str("remote", r.RemoteAddr).Str("domain", domainToForward).Str("URI", r.URL.String()).Int64("proxified bytes", proxifiedBytesCount).TimeDiff("request time (s)", time.Now(), start).Msg("Received HTTP request")

	proxyRsp.Body.Close()
	r.Body.Close()
}
