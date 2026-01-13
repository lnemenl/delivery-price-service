package models

// VenueStatic holds the location data from the static API
// We use inline structs for the wrappers to keep the code concise
type VenueStatic struct {
	VenueRaw struct {
		Location struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"location"`
	} `json:"venue_raw"`
}

// VenueDynamic holds the pricing rules from the dynamic API
type VenueDynamic struct {
	VenueRaw struct {
		DeliverySpecs struct {
			OrderMinimumNoSurcharge int             `json:"order_minimum_no_surcharge"`
			DeliveryPricing         DeliveryPricing `json:"delivery_pricing"`
		} `json:"delivery_specs"`
	} `json:"venue_raw"`
}

// DeliveryPricing defines the base price and the list of distance rules
// We name this struct because we pass it around in the Calculator
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
