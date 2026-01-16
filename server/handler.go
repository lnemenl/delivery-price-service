package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/service"
)

// PriceHandler holds the dependencies needed to process a request
type PriceHandler struct {
	client *client.APIClient
}

// New creates a new handler instance
func New(c *client.APIClient) *PriceHandler {
	return &PriceHandler{
		client: c,
	}
}

// HandleRequest is the main entry point for the HTTP traffic
func (h *PriceHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 1. Method Check
	// We only allow GET requests. Any other is rejected
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse and Validate Inputs
	venueSlug, deliveryInput, err := parseInput(r)
	if err != nil {
		// If inputs are wrong (e.g., text instead of numbers), return 400 Bad Request
		http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
		return
	}

	// 3. Fetch Data (The Courier)
	// We use the injected client to get the raw data from Wolt
	staticData, dynamicData, err := h.client.FetchVenueData(venueSlug)
	if err != nil {
		// If the external API fails, we return 500 Internal Server Error
		http.Error(w, fmt.Sprintf("Failed to fetch venue data: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Calculate Price (The Brain)
	// We pass the clean inputs and the venue data to the calculator
	priceResponse, err := service.CalculatePrice(deliveryInput, staticData, dynamicData)
	if err != nil {
		// If calculation fails (e.g., distance too far), return 400 Bad Request
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusBadRequest)
		return
	}

	// 5. Send Response
	// We set the header so the browser knows it's JSON
	w.Header().Set("Content-Type", "application/json")
	// We encode the Go struct into JSON and write it to the stream
	if err := json.NewEncoder(w).Encode(priceResponse); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// parseInput extracts and validates query parameters from the URL
// It returns the venue slug and the standardized DeliveryInput struct
func parseInput(r *http.Request) (string, service.DeliveryInput, error) {
	q := r.URL.Query()

	slug := q.Get("venue_slug")
	cartValStr := q.Get("cart_value")
	latStr := q.Get("user_lat")
	lonStr := q.Get("user_lon")

	// 1. Check Slug
	if slug == "" {
		return "", service.DeliveryInput{}, fmt.Errorf("missing venue_slug")
	}

	if len(slug) > 100 {
		return "", service.DeliveryInput{}, fmt.Errorf("venue_slug too long")
	}

	// 2. Check Cart Value (Must be number AND non-negative)
	cartValue, err := strconv.Atoi(cartValStr)
	if err != nil {
		return "", service.DeliveryInput{}, fmt.Errorf("invalid cart_value")
	}
	if cartValue < 0 {
		return "", service.DeliveryInput{}, fmt.Errorf("cart_value cannot be negative")
	}

	// 3. Check Latitude (-90 to 90)
	userLat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return "", service.DeliveryInput{}, fmt.Errorf("invalid user_lat")
	}
	if userLat < -90 || userLat > 90 {
		return "", service.DeliveryInput{}, fmt.Errorf("user_lat must be between -90 and 90")
	}

	// 4. Check Longitude (-180 to 180)
	userLon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return "", service.DeliveryInput{}, fmt.Errorf("invalid user_lon")
	}
	if userLon < -180 || userLon > 180 {
		return "", service.DeliveryInput{}, fmt.Errorf("user_lon must be between -180 and 180")
	}

	return slug, service.DeliveryInput{
		CartValue: cartValue,
		UserLat:   userLat,
		UserLon:   userLon,
	}, nil
}
