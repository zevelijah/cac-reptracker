package main

// helpers.go
// Contains utility functions for HTTP responses and middleware.

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON is a helper to write a JSON response with a given status code.
// It sets the Content-Type header and pretty-prints the JSON for readability.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ") // pretty-printed for developer readability
	if err := enc.Encode(v); err != nil {
		log.Printf("failed encoding json: %v", err)
	}
}

// enableCORS sets permissive CORS headers for local development.
// WARNING: This is not suitable for production environments.
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
