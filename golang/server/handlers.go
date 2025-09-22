package main

// handlers.go
// Contains HTTP handlers for the API endpoints.

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// statesHandler handles requests for GET /states.
// It returns a static list of US states.
func statesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	states := getStateList()
	writeJSON(w, http.StatusOK, states)
}

// representativesHandler handles requests for GET /representatives?state=XX.
// It fetches data from either a mock source or a live API based on the MODE env var.
func representativesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("state")))
	if state == "" {
		http.Error(w, "missing required 'state' query parameter (e.g. ?state=NY)", http.StatusBadRequest)
		return
	}

	mode := strings.ToLower(os.Getenv("MODE"))
	var reps []Member
	var err error

	if mode == "real" {
		reps, err = getMembers(state)
	} else {
		reps, err = getMembersMock(state)
	}

	if err != nil {
		log.Printf("error getting representatives for state %s: %v", state, err)
		http.Error(w, "internal server error while fetching representatives", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, reps)
}
