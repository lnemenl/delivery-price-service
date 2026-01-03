package api_test

import (
	"testing"

	"github.com/lnemenl/wolt_1/internal/adapter/api"
)

func TestIntegration_FetchRealWoltData(t *testing.T) {
	realBaseURL := "https://consumer-api.development.dev.woltapi.com"
	realSlug := "home-assignment-venue-helsinki"

	loc, err := api.FetchVenueLocation(realBaseURL, realSlug)

	if err != nil {
		t.Fatalf("Failed to call Real Wolt API: %v", err)
	}

	t.Logf("Latitude:  %f", loc.Lat)
	t.Logf("Longitude: %f", loc.Lon)

}
