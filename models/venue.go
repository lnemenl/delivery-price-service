package models

// VenueStatic represents the response from the /static endpoint
type VenueStatic struct {
	VenueRaw VenueRawStatic `json:"venue_raw"`
}

type VenueRawStatic struct {
	Location Location `json:"location"`
}

type Location struct {
	// [longitude, latitude]
	Coordinates []float64 `json:"coordinates"`
}

// VenueDynamic represents the response from the /dynamic endpoint
type VenueDynamic struct {
	VenueRaw VenueRawDynamic `json:"venue_raw"`
}

type VenueRawDynamic struct {
	DeliverySpecs DeliverySpecs `json:"delivery_specs"`
}

type DeliverySpecs struct {
	OrderMinimumNoSurcharge int             `json:"order_minimum_no_surcharge"`
	DeliveryPricing         DeliveryPricing `json:"delivery_pricing"`
}

type DeliveryPricing struct {
	BasePrice      int             `json:"base_price"`
	DistanceRanges []DistanceRange `json:"distance_ranges"`
}

type DistanceRange struct {
	Min  int     `json:"min"`
	Max  int     `json:"max"`
	A    int     `json:"a"`
	B    float64 `json:"b"`
	Flag *string `json:"flag"`
}
