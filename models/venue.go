package models

// VenueStatic contains location data from the static API endpoint
type VenueStatic struct {
	VenueRaw struct {
		Location struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"location"`
	} `json:"venue_raw"`
}

// VenueDynamic contains delivery specifications from the dynamic API endpoint
type VenueDynamic struct {
	VenueRaw struct {
		DeliverySpecs struct {
			OrderMinimumNoSurcharge int             `json:"order_minimum_no_surcharge"`
			DeliveryPricing         DeliveryPricing `json:"delivery_pricing"`
		} `json:"delivery_specs"`
	} `json:"venue_raw"`
}

// DeliveryPricing contains pricing configuration and distance-based fee rules
type DeliveryPricing struct {
	BasePrice      int             `json:"base_price"`
	DistanceRanges []DistanceRange `json:"distance_ranges"`
}

type DistanceRange struct {
	Min int     `json:"min"`
	Max int     `json:"max"`
	A   int     `json:"a"`
	B   float64 `json:"b"`
}
