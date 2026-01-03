package api_test

import (
	"encoding/json"
	"testing"

	"github.com/lnemenl/wolt_1/internal/adapter/api"
)

func TestParseVenueJSON(t *testing.T) {

	// 1. Arranging (Fake Wolt response)
	mockJSON := `
	{
		"venue_raw": {
			"location": {
				"coordinates": [24.93, 60.17]
			}
		}
	}
	`

	// 2. Acting
	// CREATE THE EMPTY BUCKET
	var result api.StaticResponse

	// FILL THE BUCKET
	err := json.Unmarshal([]byte(mockJSON), &result)

	// CHECK FOR SPILLS
	if err != nil {
		t.Fatalf("Parser crashed: %v", err)
	}

	// 3. Asserting
	coords := result.VanueRaw.Location.Coordinates

	t.Logf("Set coords %+v", result)

	if len(coords) < 2 {
		t.Errorf("Expected 2 coordinates, but got %d", len(coords))
	}

	if coords[0] != 24.93 {
		t.Errorf("Expected Longitude 24.93, but got %f", coords[0])
	}

	if coords[1] != 60.17 {
		t.Errorf("Expected Latitude 60.17, but got %f", coords[1])
	}

	t.Log("Test Passed!")
}

// go test -v ./...
