package domain

// Location represents a point on Earth.
type Location struct {
	Lat float64
	Lon float64
}

// DistanceRange represents one row in the pricing table.
type DistanceRange struct {
	Min int64
	Max int64
	A   int64
	B   float64
}

// DeliveryData holds all the dynamic pricing information
type DeliveryData struct {
	PricingRules      []DistanceRange
	SmallOrderMinimum int64
}
