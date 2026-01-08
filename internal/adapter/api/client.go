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
	coords := result.VenueRaw.Location.Coordinates

	if len(coords) < 2 {
		return domain.Location{}, fmt.Errorf("missing coordinates")
	}

	return domain.Location{
		Lat: coords[1],
		Lon: coords[0],
	}, nil
}

// FetchDeliveryData gets the pricing rules (Dynamic endpoint)
func FetchDeliveryData(baseURL string, slug string) (domain.DeliveryData, error) {

	// Build the URL
	// Notice the change: /static -> /dynamic
	fullURL := fmt.Sprintf("%s/home-assignment-api/v1/venues/%s/dynamic", baseURL, slug)

	// Request
	resp, err := http.Get(fullURL)
	if err != nil {
		return domain.DeliveryData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.DeliveryData{}, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	// Decode
	// We use the new DynamicResponse struct
	var result DynamicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.DeliveryData{}, fmt.Errorf("invalid json: %v", err)
	}

	// 4. Extract the Raw Rules
	// The hierarchy: VenueRaw -> DeliverySpecs -> DeliveryPricing -> DistanceRanges
	rawSpecs := result.VenueRaw.DeliverySpecs
	rawRanges := rawSpecs.DeliveryPricing.DistanceRanges

	// 5. Convert Raw Rules to Domain Rules
	// We create an empty slice to hold the clean rules
	var rules []domain.DistanceRange

	// The Loop
	for _, r := range rawRanges {
		rules = append(rules, domain.DistanceRange{
			Min: r.Min,
			Max: r.Max,
			A:   r.A,
			B:   r.B,
		})
	}

	return domain.DeliveryData{
		PricingRules:      rules,
		SmallOrderMinimum: rawSpecs.OrderMinimumNoSurcharge,
	}, nil
}
