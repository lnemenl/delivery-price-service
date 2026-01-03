package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lnemenl/wolt_1/internal/core/domain"
)

func FetchVenueLocation(baseURL string, slug string) (domain.Location, error) {
	fullURL := fmt.Sprintf("%s/home-assignment-api/v1/venues/%s/static", baseURL, slug)

	resp, err := http.Get(fullURL)

	if err != nil {
		return domain.Location{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Location{}, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result StaticResponse

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return domain.Location{}, fmt.Errorf("invalid json: %v", err)
	}
	coords := result.VanueRaw.Location.Coordinates

	if len(coords) < 2 {
		return domain.Location{}, fmt.Errorf("missing coordinates")
	}

	return domain.Location{
		Lat: coords[1],
		Lon: coords[0],
	}, nil
}
