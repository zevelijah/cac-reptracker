package main

// main.go
//
// Simple backend for "states -> representatives" demo.
// - Serves /states and /representatives?state=XX
// - Mock data mode (no external API keys required)
// - Adds basic CORS so a Flutter app running on your machine can call it.
//
// To run:
//   go run main.go
//
// By default this runs on :8080. Use MODE=real and configure API keys
// if you later want to fetch live data from an external API.

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Basic types for the JSON payloads returned to the Flutter app

// State is the JSON shape returned by /states
type State struct {
	Code string `json:"code"` // e.g., "NY"
	Name string `json:"name"` // e.g., "New York"
}

// Representative is the JSON shape returned by /representatives
type Representative struct {
	ID        string `json:"id"`        // unique id you define for the member (e.g., "rep-ny-01")
	FirstName string `json:"firstName"` // first name
	LastName  string `json:"lastName"`  // last name
	Party     string `json:"party"`     // party shorthand e.g., "D", "R", "I"
	District  string `json:"district"`  // district or at-large info
	Title     string `json:"title"`     // e.g., "Representative" or "Senator"
}

// Simple helper to write JSON with status code and content-type
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ") // pretty-printed for developer readability
	if err := enc.Encode(v); err != nil {
		log.Printf("failed encoding json: %v", err)
	}
}

// Allow simple CORS for local development (NOT production-ready)
func enableCORS(w http.ResponseWriter) {
	// Allow everything for local dev. Lock this down in production (origins, methods).
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Handler: GET /states
func statesHandler(w http.ResponseWriter, r *http.Request) {
	// Quick CORS preflight support
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In this example we return a fixed list of states + DC. In a real app
	// you might store this in DB or pull from a canonical source.
	states := getStateList()
	writeJSON(w, http.StatusOK, states)
}

// Handler: GET /representatives?state=XX
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

	// Validate query parameter
	state := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("state")))
	if state == "" {
		http.Error(w, "missing state query parameter (e.g. ?state=NY)", http.StatusBadRequest)
		return
	}

	// Choose mode from env var (default: mock)
	mode := strings.ToLower(os.Getenv("MODE")) // "mock" or "real"
	var reps []Representative
	var err error
	if mode == "real" {
		// Placeholder: call the real API integration here
		// Implement getRepresentativesFromRealAPI to call Congress.gov, ProPublica, etc.
		reps, err = getRepresentativesFromRealAPI(state)
	} else {
		// Mock data (fast and simple for development)
		reps, err = getRepresentativesMock(state)
	}
	if err != nil {
		log.Printf("error getting representatives for %s: %v", state, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, reps)
}

func main() {
	// Simple router
	http.HandleFunc("/states", statesHandler)
	http.HandleFunc("/representatives", representativesHandler)

	// Server with timeout settings to avoid Leaky Go routines in real deployments
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting server on :8080 (MODE=", os.Getenv("MODE"), ")")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

//
// ---------- Mock data + placeholders below ----------
//

// getStateList returns standard US states + DC.
// Keep codes two-letter for client convenience.
func getStateList() []State {
	return []State{
		{"AL", "Alabama"}, {"AK", "Alaska"}, {"AZ", "Arizona"}, {"AR", "Arkansas"},
		{"CA", "California"}, {"CO", "Colorado"}, {"CT", "Connecticut"}, {"DE", "Delaware"},
		{"FL", "Florida"}, {"GA", "Georgia"}, {"HI", "Hawaii"}, {"ID", "Idaho"},
		{"IL", "Illinois"}, {"IN", "Indiana"}, {"IA", "Iowa"}, {"KS", "Kansas"},
		{"KY", "Kentucky"}, {"LA", "Louisiana"}, {"ME", "Maine"}, {"MD", "Maryland"},
		{"MA", "Massachusetts"}, {"MI", "Michigan"}, {"MN", "Minnesota"}, {"MS", "Mississippi"},
		{"MO", "Missouri"}, {"MT", "Montana"}, {"NE", "Nebraska"}, {"NV", "Nevada"},
		{"NH", "New Hampshire"}, {"NJ", "New Jersey"}, {"NM", "New Mexico"}, {"NY", "New York"},
		{"NC", "North Carolina"}, {"ND", "North Dakota"}, {"OH", "Ohio"}, {"OK", "Oklahoma"},
		{"OR", "Oregon"}, {"PA", "Pennsylvania"}, {"RI", "Rhode Island"}, {"SC", "South Carolina"},
		{"SD", "South Dakota"}, {"TN", "Tennessee"}, {"TX", "Texas"}, {"UT", "Utah"},
		{"VT", "Vermont"}, {"VA", "Virginia"}, {"WA", "Washington"}, {"WV", "West Virginia"},
		{"WI", "Wisconsin"}, {"WY", "Wyoming"}, {"DC", "District of Columbia"},
	}
}

// getRepresentativesMock returns a small set of fake representatives for demo.
// IMPORTANT: This is mock data and not real representatives.
func getRepresentativesMock(state string) ([]Representative, error) {
	// Minimal example mapping; expand as you like.
	mockDB := map[string][]Representative{
		"NY": {
			{ID: "rep-ny-1", FirstName: "Alex", LastName: "Johnson", Party: "D", District: "1", Title: "Representative"},
			{ID: "rep-ny-2", FirstName: "Riley", LastName: "Martinez", Party: "R", District: "2", Title: "Representative"},
		},
		"CA": {
			{ID: "rep-ca-12", FirstName: "Morgan", LastName: "Lee", Party: "D", District: "12", Title: "Representative"},
			{ID: "rep-ca-14", FirstName: "Taylor", LastName: "Nguyen", Party: "D", District: "14", Title: "Representative"},
		},
		"TX": {
			{ID: "rep-tx-7", FirstName: "Sam", LastName: "Williams", Party: "R", District: "7", Title: "Representative"},
		},
		"DC": {
			{ID: "del-dc", FirstName: "Jamie", LastName: "Green", Party: "I", District: "At-Large", Title: "Delegate"},
		},
	}

	// Return empty slice if unknown (client handles empty list).
	if reps, ok := mockDB[state]; ok {
		return reps, nil
	}
	return []Representative{}, nil
}

// getRepresentativesFromRealAPI is a placeholder to show where live calls would go.
// For a real implementation:
//   - call your chosen API (Congress.gov, ProPublica, OpenStates, etc.)
//   - normalize the response into []Representative
//   - handle rate limits, caching, retries
func getRepresentativesFromRealAPI(state string) ([]Representative, error) {
	// TODO: implement using your preferred authoritative API.
	// Example steps:
	// 1) Lookup members for 'state' in the real API (may require API key).
	// 2) Map fields to Representative struct.
	// 3) Return list.
	//
	// For the demo we just return an empty list.
	return []Representative{}, nil
}
