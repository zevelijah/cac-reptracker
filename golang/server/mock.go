package main

// mock.go
// Provides mock data for development and testing.

// getMembersMock returns a small set of fake representatives for demo.
// IMPORTANT: This is mock data and not real representatives.
func getMembersMock(state string) ([]Member, error) {
	// Minimal example mapping; expand as you like.
	mockDB := map[string][]Member{
		"NY": {
			{ID: "rep-ny-1", FirstName: "Alex", LastName: "Johnson", Party: " (D)", District: "1"},
			{ID: "rep-ny-2", FirstName: "Riley", LastName: "Martinez", Party: " (R)", District: "2"},
		},
		"CA": {
			{ID: "rep-ca-12", FirstName: "Morgan", LastName: "Lee", Party: " (D)", District: "12"},
			{ID: "rep-ca-14", FirstName: "Taylor", LastName: "Nguyen", Party: " (D)", District: "14"},
		},
		"TX": {
			{ID: "rep-tx-7", FirstName: "Sam", LastName: "Williams", Party: " (R)", District: "7"},
		},
	}

	// Return empty slice if unknown (client handles empty list).
	if reps, ok := mockDB[state]; ok {
		return reps, nil
	}
	return []Member{}, nil
}

// getStateList returns standard US states + DC.
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
		{"WI", "Wisconsin"}, {"WY", "Wyoming"},
	}
}
