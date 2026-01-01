package api

type VenueCoordinates struct {
	Coordinates []float64 `json:"coordinates"`
}

type VenueLocation struct {
	Location VenueCoordinates `json:"location"`
}

type VenueInfo struct {
	VanueRaw VenueLocation `json:"venue_raw"`
}
