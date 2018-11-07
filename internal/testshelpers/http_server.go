// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package testshelpers

import (
	ctx "context"
	"fmt"
	"net/http"
	"time"
)

// CreateHTTPServer creates HTTP server on selected port. Useful for proxy requests
// testing.
func CreateHTTPServer(port string, host string, color string, number string) chan bool {
	listenAddress := "127.0.0.1:" + port
	srv := &http.Server{
		Addr: listenAddress,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received for color: " + color + ", host " + host + " (from request: " + r.Host + "), backend#" + number)
		w.WriteHeader(200)
		_, err := w.Write([]byte("Color: " + color + ", host " + host + ", backend#" + number))
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	srv.Handler = mux
	closeChan := make(chan bool, 1)

	go func() {
		fmt.Println("Listening on " + listenAddress + " for color " + color + ", host " + host + ", backend#" + number)
		err := srv.ListenAndServe()
		if err != nil {
			fmt.Println(err.Error())
		}
		for range closeChan {
			closedownContext, closedownCancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
			srv.Close()
			err := srv.Shutdown(closedownContext)
			if err != nil {
				fmt.Println(err.Error())
			}
			closedownCancel()
		}
	}()

	return closeChan
}
