// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

var (
	proxiesModuleLog zerolog.Logger

	// bunch of http.Server, which represents current proxy list
	proxies []*http.Server
)

// Proxy handles ServeHTTP function for passing data inside proxy
type Proxy struct {
	Source       string
	Destinations []string
}

func initProxies() {
	proxiesModuleLog = domainLog.With().Str("module", "proxies").Logger()
	proxiesModuleLog.Info().Msg("Initializing proxies...")
}

// startProxy starts proxy with desired configuration and adds it to proxies array
func startProxy(src string, dst []string) {
	proxiesModuleLog.Debug().Msgf("Starting proxying on %s to %s...", src, strings.Join(dst, ", "))

	srv := &http.Server{
		Addr:    src,
		Handler: newProxy(src, dst),
	}

	go func() {
		srv.ListenAndServe()
	}()

	proxies = append(proxies, srv)
}

func newProxy(src string, dst []string) *Proxy {
	proxy := Proxy{
		Source:       src,
		Destinations: dst,
	}
	return &proxy
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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
	io.Copy(w, proxyRsp.Body)
	proxyRsp.Body.Close()
}
