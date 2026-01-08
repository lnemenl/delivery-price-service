package api

// STATIC DATA (Location)
// JSON Path: venue_raw -> location -> coordinates

// Top Level for Static Endpoint
type StaticResponse struct {
	VenueRaw VenueRawStatic `json:"venue_raw"`
}

type VenueRawStatic struct {
	Location Location `json:"location"`
}

type Location struct {
	Coordinates []float64 `json:"coordinates"`
}

// DYNAMIC DATA (Pricing Rules)
// JSON Path: venue_raw -> delivery_specs -> delivery_pricing -> distance_ranges

// Top Level for Dynamic Endpoint
type DynamicResponse struct {
	VenueRaw VenueRawDynamic `json:"venue_raw"`
}

type VenueRawDynamic struct {
	DeliverySpecs DeliverySpecs `json:"delivery_specs"`
}

type DeliverySpecs struct {
	OrderMinimumNoSurcharge int64           `json:"order_minimum_no_surcharge"`
	DeliveryPricing         DeliveryPricing `json:"delivery_pricing"`
}

type DeliveryPricing struct {
	DistanceRanges []DistanceRangeRaw `json:"distance_ranges"`
}

// The Gold Nugget: One single rule
type DistanceRangeRaw struct {
	Min int64   `json:"min"`
	Max int64   `json:"max"`
	A   int64   `json:"a"`
	B   float64 `json:"b"`
}
