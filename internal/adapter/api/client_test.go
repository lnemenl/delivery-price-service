package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchVenueLocation(t *testing.T) {
	// 1. Aranging
	// Create a server. The function inside decides what the server says back
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Safety Check: Did our client ask for the right URL?
		// We expect the client to ask for: /home-assignment-api/v1/venues/.../static
		expectedPath := "/home-assignment-api/v1/venues/test-venue/static"

		if r.URL.Path != expectedPath {
			t.Errorf("Client asked for wrong path. Expected %s, got %s", expectedPath, r.URL.Path)
		}

		// Return 200 OK
		w.WriteHeader(http.StatusOK)

		// Return the JSON (The same simple one we verified in parser_test.go)
		w.Write([]byte(`{
            "venue_raw": {
                "location": {
                    "coordinates": [24.93, 60.17]
                }
            }
        }`))
	}))
	// Close the server when the test finishes, or it keeps running
	defer mockServer.Close()

}
