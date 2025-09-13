// congress_example.go
//
// Simple, well-commented example showing how to call Congress.gov v3 API
// from Go.  It demonstrates:
//   - reading API key from env var
//   - building request URLs and query params
//   - simple retry/error handling
//   - decoding JSON generically so the example works across endpoints
//
// Notes:
//   - Set environment variable CONGRESS_API_KEY to your api.data.gov key before running.
//   - Congress.gov responses include a "request", "pagination", and a data element
//     whose name varies by endpoint (e.g., "bills", "houseVotes", "members", ...).
package main

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

// BaseURL for Congress.gov v3 API.
// We'll append endpoint paths like "/bill" or "/house-vote/118/2/3/members"
const baseURL = "https://api.congress.gov/v3"

// readAPIKey gets the API key from the environment and returns an error if missing.
func readAPIKey() (string, error) {
	key := os.Getenv("CONGRESS_API_KEY")
	if key == "" {
		return "", errors.New("CONGRESS_API_KEY environment variable is not set; get a key at api.data.gov")
	}
	return key, nil
}

// fetchJSON performs a GET to the Congress.gov API path with provided query params.
// The function returns the raw JSON bytes of the response body.
// It adds the required api_key parameter automatically.
func fetchJSON(path string, params map[string]string) ([]byte, error) {
	apiKey, err := readAPIKey()
	if err != nil {
		return nil, err
	}

	// Build URL: base + path + query params (including api_key)
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Use url.Values so params are encoded properly
	q := u.Query()
	q.Set("api_key", apiKey) // Congress.gov uses api.data.gov key param
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Prepare an HTTP client with a timeout
	client := &http.Client{Timeout: 15 * time.Second}

	// Basic retry loop (2 retries) for transient network errors
	var resp *http.Response
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = client.Get(u.String())
		if err != nil {
			// small backoff before retrying
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
			continue
		}
		// got a response — break out
		break
	}
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Expect HTTP 200; otherwise return useful debugging info
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read whole body (safe for reasonably sized API responses)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body failed: %w", err)
	}
	return body, nil
}

// decodeData parses the Congress.gov response and returns the data element (array of items).
func decodeData(raw []byte) ([]map[string]interface{}, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	for k, v := range top {
		if k == "request" || k == "pagination" {
			continue
		}
		var arr []map[string]interface{}
		if err := json.Unmarshal(v, &arr); err == nil {
			return arr, nil
		}
	}
	return nil, errors.New("no array data element found in response")
}

// printData pretty-prints the array of items.
func printData(data []map[string]interface{}) {
	fmt.Printf("Data element (array of %d items):\n", len(data))
	for _, item := range data {
		val, _ := json.MarshalIndent(item, "  ", "  ")
		fmt.Println(string(val))
	}
}

func georgia_members() error {
	// Example: Request a list of members from Georgia (stateCode=GA).
	// Endpoint: GET /v3/member?stateCode=GA
	fmt.Println("=== Example: /v3/member (list) ===")
	memberRaw, err := fetchJSON("/member/GA", map[string]string{
		"format": "json", // optional; API supports JSON
	})
	if err != nil {
		return fmt.Errorf("failed fetching members: %v", err)
	}
	if arr, err := decodeData(memberRaw); err != nil {
		return fmt.Errorf("failed decoding members response: %v", err)
	} else {
		if len(arr) != 0 {
			for _, m := range arr {
				val, _ := json.Marshal(m)
				if strings.Count(string(val), "startYear") > strings.Count(string(val), "endYear") {
					printData(arr)
				}
			}
		}
	}

	return nil
}

func printSponsoredLegislation(raw []byte) error {
	if arr, err := decodeData(raw); err != nil {
		return fmt.Errorf("failed decoding sponsored legislation response: %v", err)
	} else {
		printData(arr)
	}
	return nil
}

func printCosponsoredLegislation(raw []byte) error {
	if arr, err := decodeData(raw); err != nil {
		return fmt.Errorf("failed decoding cosponsored legislation response: %v", err)
	} else {
		printData(arr)
	}
	return nil
}

func specific_member_sponsored(bioguide_id string) error {
	// Example: Request a list of votes for a specific member by bioguide_id.
	// Endpoint: GET /v3/member/{bioguide_id}/votes
	fmt.Printf("\n=== Example: /v3/member/%s/sponsored-legislation ===\n", bioguide_id)
	sponsoredRaw, err := fetchJSON(fmt.Sprintf("/member/%s/sponsored-legislation", bioguide_id), map[string]string{
		"format": "json", // optional; API supports JSON
		"limit":  "20",   // limit to 20 results
	})
	if err != nil {
		return fmt.Errorf("failed fetching member sponsored legislation: %v", err)
	}
	if err := printSponsoredLegislation(sponsoredRaw); err != nil {
		return err
	}
	fmt.Printf("\n=== Example: /v3/member/%s/cosponsored-legislation ===\n", bioguide_id)
	cosponsoredRaw, err := fetchJSON(fmt.Sprintf("/member/%s/cosponsored-legislation", bioguide_id), map[string]string{
		"format": "json", // optional; API supports JSON
		"limit":  "20",   // limit to 20 results
	})
	if err != nil {
		return fmt.Errorf("failed fetching member cosponsored legislation: %v", err)
	}
	if err := printCosponsoredLegislation(cosponsoredRaw); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := georgia_members(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// if err := specific_member_sponsored("G000596"); err != nil { // bioguide_id for Marjorie Taylor Greene
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
}
