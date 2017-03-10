// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

// This program provides a sample web service that implements a
// RESTFul CRUD API against a MongoDB database.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ardanlabs/gotraining/starter-kits/http/cmd/apid/routes"
)

// init is called before main. We are using init to customize logging output.
func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

// main is the entry point for the application.
func main() {
	log.Println("main : Started")

	// Check the environment for a configured port value.
	host := os.Getenv("HOST")
	if host == "" {
		host = ":3000"
	}

	// Create a new server and set timeout values.
	server := http.Server{
		Addr:           host,
		Handler:        routes.API(),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start listening for requests.
	go func() {
		log.Println("listener : Listening on " + host)
		err := server.ListenAndServe()
		log.Printf("listener : %v", err)
	}()

	// Block until there's an interrupt then shut the server down. The main
	// func must not return before this process is complete or in-flight
	// requests will be aborted.
	shutdownOnInterrupt(&server)

	log.Println("main : Completed")
}

func shutdownOnInterrupt(server *http.Server) {

	// Set up channel to receive interrupt signals.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	log.Println("shutdown : Waiting for a shutdown signal")
	<-c

	log.Println("shutdown : Signal received. Attempting graceful shut down...")

	// Create a context with a 5 second timeout. If the server can't
	// gracefully shut down in that time we'll kill it.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Try a graceful shutdown.
	if err := server.Shutdown(ctx); err != nil {

		// Couldn't shut down gracefully. Try a forceful shutdown.
		log.Printf("shutdown : Graceful shutdown did not complete in %v : %v", timeout, err)
		log.Print("shutdown : Killing server.")

		if err := server.Close(); err != nil {
			log.Printf("shutdown : Error killing server : %v", err)
		}
	}
}
