package main

import (
	// stdlib
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var httpServersCloseChannels []chan bool

func main() {
	// Start 4 servers.
	c1 := createHTTPServer("8123", "web.host", "green", "1")
	c2 := createHTTPServer("8124", "web.host", "green", "2")
	c3 := createHTTPServer("8223", "web2.host", "green", "1")
	c4 := createHTTPServer("8224", "web2.host", "green", "2")

	c5 := createHTTPServer("9123", "web.host", "blue", "1")
	c6 := createHTTPServer("9124", "web.host", "blue", "2")
	c7 := createHTTPServer("9223", "web2.host", "blue", "1")
	c8 := createHTTPServer("9224", "web2.host", "blue", "2")

	// CTRL+C handler.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)
	shutdownDone := make(chan bool, 1)
	go func() {
		signalThing := <-interrupt
		if signalThing == syscall.SIGTERM || signalThing == syscall.SIGINT {

			c1 <- true
			c2 <- true
			c3 <- true
			c4 <- true
			c5 <- true
			c6 <- true
			c7 <- true
			c8 <- true
			shutdownDone <- true
		}
	}()

	<-shutdownDone
	fmt.Println("Shutted down")
	os.Exit(0)
}

func createHTTPServer(port string, host string, color string, number string) chan bool {
	listenAddress := "127.0.0.1:" + port
	srv := &http.Server{
		Addr: listenAddress,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received for color: " + color + ", host " + host + ", backend#" + number)
		w.WriteHeader(200)
		w.Write([]byte("Color: " + color + ", host " + host + ", backend#" + number))
	})

	srv.Handler = mux
	closeChan := make(chan bool, 1)

	go func() {
		fmt.Println("Listening on " + listenAddress + " for color " + color + ", host " + host + ", backend#" + number)
		srv.ListenAndServe()
		for {
			select {
			case <-closeChan:
				srv.Close()
			}
		}
	}()

	return closeChan
}
