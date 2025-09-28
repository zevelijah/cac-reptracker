package main

import (
	"fmt"
	"strings"
)

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
	ID         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Party      string `json:"party"`
	District   string `json:"district"`
	InitalYear int    `json:"initialYear"`
	ImageURL   string `json:"imageUrl"`
}

// ApiMember is structured to match the member object from the Congress.gov API.
type ApiMember struct {
	BioguideID string `json:"bioguideId"`
	Name       string `json:"name"`
	Terms      struct {
		Item []struct {
			Chamber   string `json:"chamber"` // "House" or "Senate"
			StartYear int    `json:"startYear"`
			EndYear   *int   `json:"endYear,omitempty"`
		} `json:"item"`
	} `json:"terms"`
	State     string `json:"state"`
	District  int    `json:"district"`
	Party     string `json:"partyName"`
	Depiction struct {
		ImageURL *string `json:"imageUrl,omitempty"`
	} `json:"depiction"`
}

// apiMemberToMember converts an ApiMember to a client-facing Member, for current members only.
// It returns the member and a boolean indicating if the conversion was successful (i.e., is a current member).
func apiMemberToMember(apiM ApiMember) (Member, bool) {
	if len(apiM.Terms.Item) == 0 {
		return Member{}, false
	}

	lastTerm := apiM.Terms.Item[len(apiM.Terms.Item)-1]
	firstTerm := apiM.Terms.Item[0]

	// Only include current members (those without an end year on their last term)
	if lastTerm.EndYear != nil {
		return Member{}, false
	}

	// Use the provided image URL if available; otherwise, construct a default one.
	var imageURL string
	if apiM.Depiction.ImageURL != nil && *apiM.Depiction.ImageURL != "" {
		imageURL = *apiM.Depiction.ImageURL
	} else {
		imageURL = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/6f/2023_United_States_Capitol_118th_Congress%2C_sunrise_%28Cropped%29.jpg/640px-2023_United_States_Capitol_118th_Congress%2C_sunrise_%28Cropped%29.jpg"
	}

	// Parse name in "Last, First Middle" format
	var firstName, lastName string
	parts := strings.SplitN(apiM.Name, ",", 2)
	if len(parts) == 2 {
		lastName = strings.TrimSpace(parts[0])
		firstMiddle := strings.TrimSpace(parts[1])
		// Take just the first name from "First Middle"
		firstName = strings.Split(firstMiddle, " ")[0]
	} else {
		// Fallback for names not in "Last, First" format
		lastName = strings.TrimSpace(apiM.Name)
		firstName = ""
	}

	var partyDisplay string
	switch apiM.Party {
	case "Democratic":
		partyDisplay = " (D)"
	case "Republican":
		partyDisplay = " (R)"
	case "Independent":
		partyDisplay = " (I)"
	case "Libertarian":
		partyDisplay = " (L)"
	case "Green":
		partyDisplay = " (G)"
	default:
		// For other parties, just use the initial if available.
		if len(apiM.Party) > 0 {
			partyDisplay = fmt.Sprintf(" (%s)", apiM.Party)
		}
	}

	var districtDisplay string

	if apiM.District == 0 {
		// Find the current term (the one without EndYear)
		if strings.EqualFold(lastTerm.Chamber, "Senate") {
			districtDisplay = "Senator"
		} else {
			districtDisplay = "At-Large Rep."
		}
	} else {
		districtDisplay = fmt.Sprintf("District %d Rep.", apiM.District)
	}

	return Member{
		ID:         apiM.BioguideID,
		FirstName:  firstName,
		LastName:   lastName,
		Party:      partyDisplay,
		District:   districtDisplay,
		InitalYear: firstTerm.StartYear,
		ImageURL:   imageURL,
	}, true
}
