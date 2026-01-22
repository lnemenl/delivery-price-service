package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lnemenl/delivery-price-service/client"
	"github.com/lnemenl/delivery-price-service/service"
)

// PriceHandler holds the dependencies needed to process a request
type PriceHandler struct {
	client *client.APIClient
}

// New creates a new handler instance
func New(c *client.APIClient) *PriceHandler {
	return &PriceHandler{client: c}
}

// HandleRequest is the main entry point for the HTTP traffic
func (h *PriceHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate input parameters
	venueSlug, deliveryInput, err := parseInput(r)
	if err != nil {
		// Return 400 Bad Request for invalid input
		http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch venue data from the API
	staticData, dynamicData, err := h.client.FetchVenueData(venueSlug)
	if err != nil {
		// Map 404 errors to Not Found response
		if strings.Contains(err.Error(), "404") {
			http.Error(w, "Venue not found", http.StatusNotFound)
			return
		}

		// Return 500 for other API errors
		http.Error(w, fmt.Sprintf("Failed to fetch venue data: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate delivery price
	priceResponse, err := service.CalculatePrice(deliveryInput, staticData, dynamicData)
	if err != nil {
		// Return 400 Bad Request for calculation errors
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusBadRequest)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	// Encode the response struct to JSON
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

	// Verify venue slug is provided and not too long
	if slug == "" {
		return "", service.DeliveryInput{}, fmt.Errorf("missing venue_slug")
	}

	if len(slug) > 100 {
		return "", service.DeliveryInput{}, fmt.Errorf("venue_slug too long")
	}

	// Verify cart value is a valid non-negative integer
	if cartValStr == "" {
		return "", service.DeliveryInput{}, fmt.Errorf("missing cart_value")
	}
	cartValue, err := strconv.Atoi(cartValStr)
	if err != nil {
		return "", service.DeliveryInput{}, fmt.Errorf("invalid cart_value")
	}
	if cartValue < 0 {
		return "", service.DeliveryInput{}, fmt.Errorf("cart_value cannot be negative")
	}

	// Verify latitude is within valid range (-90 to 90)
	if latStr == "" {
		return "", service.DeliveryInput{}, fmt.Errorf("missing user_lat")
	}
	userLat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return "", service.DeliveryInput{}, fmt.Errorf("invalid user_lat")
	}
	if userLat < -90 || userLat > 90 {
		return "", service.DeliveryInput{}, fmt.Errorf("user_lat must be between -90 and 90")
	}

	// Verify longitude is within valid range (-180 to 180)
	if lonStr == "" {
		return "", service.DeliveryInput{}, fmt.Errorf("missing user_lon")
	}
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
