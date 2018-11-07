// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package proxiesv1

import (
	ctx "context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lab.wtfteam.pro/wtfteam/lbtds/domains/colors/v1"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/testshelpers"
)

/* exported.go */

func TestInitialize(t *testing.T) {
	// Before running this bunch of tests we need to clear temporary files
	err := os.Remove("/tmp/lbtds-test-current")
	if err != nil {
		fmt.Println("Failed to erase files from previous test: " + err.Error())
	}

	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	require.Equal(t, 0, len(httpProxies))

	testshelpers.FlushConfiguration("lbtds-valid")
}

/* dispatcher.go */

func TestDispatchChange(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	require.Equal(t, 0, len(httpProxies))

	colorsv1.GetCurrentColor()
	dispatchChange()

	require.Equal(t, 2, len(httpProxies))

	Shutdown()

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestDispatchChangeTwice(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-different-backends")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	require.Equal(t, 0, len(httpProxies))

	// At first try we're coming to green: in this configuration green has 3 backends
	colorsv1.GetCurrentColor()
	dispatchChange()
	require.Equal(t, 3, len(httpProxies))
	// At second try we shoud communicate via channel
	err := colorsv1.SetCurrentColor("blue")
	require.Nil(t, err)
	// ... and wait some reasonable timeout (we can't shutdown old proxies immediately)
	time.Sleep(3 * time.Second)
	// ...and switch to configuration, where 2 backends
	require.Equal(t, 2, len(httpProxies))

	Shutdown()

	testshelpers.FlushConfiguration("lbtds-different-backends")
}

/* http_proxies.go */

func TestServeHTTPRequestWithoutWorkingDownstream(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	httpProxy := newHTTPProxy("web.host", []string{"127.0.0.1:8123", "127.0.0.1:8124"})
	httpProxyServer := &http.Server{
		Addr:    "127.0.0.1:8100",
		Handler: httpProxy,
	}
	closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer closedownCancel()

	go func() {
		err := httpProxyServer.ListenAndServe()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	replyBody, replyCode := testshelpers.HTTPClearTestRequest(t, "http://127.0.0.1:8100/", "web.host", nil, nil, "GET", httpProxy.ServeHTTP)
	require.NotEmpty(t, replyBody)
	require.Equal(t, 502, replyCode)
	require.Equal(t, "Can't connect to downstream\n", string(replyBody))

	err := httpProxyServer.Shutdown(closedownContext)
	if err != nil {
		fmt.Println(err.Error())
	}
	time.Sleep(5 * time.Second)

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestServeHTTPValidRequest(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	httpProxy := newHTTPProxy("web.host", []string{"127.0.0.1:8123", "127.0.0.1:8124"})
	httpProxyServer := &http.Server{
		Addr:    "127.0.0.1:8100",
		Handler: httpProxy,
	}
	closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer closedownCancel()

	go func() {
		err := httpProxyServer.ListenAndServe()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	c1 := testshelpers.CreateHTTPServer("8123", "web.host", "green", "1")
	c2 := testshelpers.CreateHTTPServer("8124", "web.host", "green", "2")

	// Get some time for test backends to start
	time.Sleep(3 * time.Second)

	replyBody, replyCode := testshelpers.HTTPClearTestRequest(t, "http://127.0.0.1:8100/", "web.host", nil, nil, "GET", httpProxy.ServeHTTP)
	require.NotEmpty(t, replyBody)
	require.Equal(t, 200, replyCode)
	require.Contains(t, string(replyBody), "green")
	require.Contains(t, string(replyBody), "web.host")

	c1 <- true
	c2 <- true
	err := httpProxyServer.Shutdown(closedownContext)
	if err != nil {
		fmt.Println(err.Error())
	}
	time.Sleep(5 * time.Second)

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestServeHTTPRequestWithInvalidDomain(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	colorsv1.Initialize(c)
	Initialize(c)

	httpProxy := newHTTPProxy("web.host", []string{"127.0.0.1:8123", "127.0.0.1:8124"})
	httpProxyServer := &http.Server{
		Addr:    "127.0.0.1:8100",
		Handler: httpProxy,
	}
	closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer closedownCancel()

	go func() {
		err := httpProxyServer.ListenAndServe()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	c1 := testshelpers.CreateHTTPServer("8123", "web.host", "green", "1")
	c2 := testshelpers.CreateHTTPServer("8124", "web.host", "green", "2")

	// Get some time for test backends to start
	time.Sleep(3 * time.Second)

	replyBody, replyCode := testshelpers.HTTPClearTestRequest(t, "http://127.0.0.1:8100/", "invalid.host", nil, nil, "GET", httpProxy.ServeHTTP)
	require.NotEmpty(t, replyBody)
	require.Equal(t, 400, replyCode)
	require.Equal(t, "Invalid domain\n", string(replyBody))

	c1 <- true
	c2 <- true
	err := httpProxyServer.Shutdown(closedownContext)
	if err != nil {
		fmt.Println(err.Error())
	}
	time.Sleep(5 * time.Second)

	testshelpers.FlushConfiguration("lbtds-valid")
}
