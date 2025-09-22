package main

// main.go
//
// Simple backend for "states -> representatives" demo.
// This file contains the main function to start the web server.
//
// To run:
//   go run .
//
// By default this runs on :8080. Use MODE=real and configure API keys
// if you want to fetch live data from an external API.

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Register handlers for the API endpoints.
	http.HandleFunc("/states", statesHandler)
	http.HandleFunc("/representatives", representativesHandler)

	// Configure the server with timeouts to enhance stability.
	port := "8080"
	addr := ":" + port
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "mock" // Default to mock mode if not set
	}

	log.Printf("Starting server on %s (MODE=%s)", addr, mode)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
