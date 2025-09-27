package main

// api.go
// Contains logic for interacting with the external Congress.gov API.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const baseURL = "https://api.congress.gov/v3"

var memberCache = NewCache()

// getMembers fetches member data for a given state from the Congress.gov API.
// It handles API key reading, request building, retries, and response parsing.
func getMembers(state string) ([]Member, error) {
	// Check cache first
	if cachedMembers, found := memberCache.Get(state); found {
		return cachedMembers, nil
	}

	// Create a map of state codes (e.g., "AR") to full state names (e.g., "Arkansas")
	// to filter the API results, which use the full name.
	stateCodeToNameMap := make(map[string]string)
	for _, s := range getStateList() {
		stateCodeToNameMap[s.Code] = s.Name
	}
	stateFullName, ok := stateCodeToNameMap[state]
	if !ok {
		// Return an empty slice for an invalid state code; the client can handle it.
		return []Member{}, nil
	}

	path := fmt.Sprintf("/member/%s", state)
	rawJSON, err := fetchJSON(path, map[string]string{
		"format": "json",
		"limit":  "75", // Fetch all members of Congress to filter locally
	})
	if err != nil {
		return nil, fmt.Errorf("failed fetching all members: %w", err)
	}

	allApiMembers, err := decodeData(rawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed decoding all members response: %w", err)
	}

	var members []Member
	// Iterate through all members and filter for the requested state.
	for _, apiM := range allApiMembers {
		if apiM.State == stateFullName {
			if member, ok := apiMemberToMember(apiM); ok {
				members = append(members, member)
			}
		}
	}

	// Cache the filtered result for this specific state
	memberCache.Set(state, members, 1*time.Hour)

	return members, nil
}

// apiMemberToMember converts an ApiMember to a client-facing Member, for current members only.
// It returns the member and a boolean indicating if the conversion was successful (i.e., is a current member).
func apiMemberToMember(apiM ApiMember) (Member, bool) {
	if len(apiM.Terms.Item) == 0 {
		return Member{}, false
	}

	lastTerm := apiM.Terms.Item[len(apiM.Terms.Item)-1]

	// Only include current members (those without an end year on their last term)
	if lastTerm.EndYear != nil {
		return Member{}, false
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
		ID:        apiM.BioguideID,
		FirstName: firstName,
		LastName:  lastName,
		Party:     partyDisplay,
		District:  districtDisplay,
	}, true
}

// readAPIKey retrieves the Congress.gov API key from environment variables.
func readAPIKey() (string, error) {
	key := os.Getenv("LOC_API_KEY")
	if key == "" {
		return "", errors.New("LOC_API_KEY environment variable is not set; get a key at api.data.gov")
	}
	return key, nil
}

// decodeData extracts the list of members from the nested JSON response structure.
// The Congress.gov API wraps the main data array in a dynamic key (e.g., "members").
func decodeData(raw []byte) ([]ApiMember, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("invalid json structure: %w", err)
	}

	for key, value := range top {
		if key == "request" || key == "pagination" {
			continue
		}
		var apiMembers []ApiMember
		if err := json.Unmarshal(value, &apiMembers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal api members from key '%s': %w", key, err)
		}
		return apiMembers, nil
	}
	return nil, errors.New("no member data array found in API response")
}

// fetchJSON performs a GET request to a specified path of the Congress.gov API.
// It automatically adds the API key and includes a simple retry mechanism.
func fetchJSON(path string, params map[string]string) ([]byte, error) {
	apiKey, err := readAPIKey()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid api path: %w", err)
	}

	q := u.Query()
	q.Set("api_key", apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}

	var resp *http.Response
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = client.Get(u.String())
		if err == nil {
			break // Success
		}
		if attempt == maxRetries {
			return nil, fmt.Errorf("http request failed after %d attempts: %w", maxRetries, err)
		}
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("api returned non-200 status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}
	return body, nil
}
