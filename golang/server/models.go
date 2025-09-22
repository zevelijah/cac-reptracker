package main

// models.go
// Contains data structures for JSON payloads and API responses.

// State is the JSON shape returned by /states
type State struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Member is the JSON shape returned to the Flutter app for /representatives.
// This structure is tailored for the client's needs.
type Member struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Party     string `json:"party"`
	District  string `json:"district"`
}

// ApiMember is structured to match the member object from the Congress.gov API.
type ApiMember struct {
	BioguideID string   `json:"bioguideId"`
	Name       string   `json:"name"`
	Terms      ApiTerms `json:"terms"`
	State      string   `json:"state"`
	District   int      `json:"district"`
	Party      string   `json:"partyName"`
}

// ApiTerms represents the nested "terms" object in the API response.
type ApiTerms struct {
	Item []ApiTermItem `json:"item"`
}

// ApiTermItem represents a single term in the "terms.item" array.
type ApiTermItem struct {
	Chamber   string `json:"chamber"` // "House" or "Senate"
	StartYear int    `json:"startYear"`
	EndYear   *int   `json:"endYear,omitempty"`
}
