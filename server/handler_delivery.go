package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lnemenl/delivery-price-service/client"
	"github.com/lnemenl/delivery-price-service/service"
)

// DeliveryHandler holds the dependencies needed to process a request
type DeliveryHandler struct {
	client *client.APIClient
}

// NewDeliveryHandler creates a new handler instance
func NewDeliveryHandler(c *client.APIClient) *DeliveryHandler {
	return &DeliveryHandler{client: c}
}

// HandleRequest is the main entry point for the HTTP traffic
func (h *DeliveryHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	venueSlug, deliveryInput, err := parseInput(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
		return
	}

	staticData, dynamicData, err := h.client.FetchVenueData(r.Context(), venueSlug)
	if err != nil {
		if errors.Is(err, client.ErrVenueNotFound) {
			http.Error(w, "Venue not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to fetch venue data: %v", err), http.StatusInternalServerError)
		return
	}

	venueInfo, err := service.MergeToVenueInfo(staticData, dynamicData)
	if err != nil {
		// If merging fails, it's an issue with the upstream data
		http.Error(w, fmt.Sprintf("Upstream data error: %v", err), http.StatusBadGateway)
		return
	}

	priceResponse, err := service.CalculatePrice(deliveryInput, venueInfo)
	if err != nil {
		if errors.Is(err, service.ErrDistanceTooLong) || errors.Is(err, service.ErrNoRangeFound) {
			http.Error(w, fmt.Sprintf("Delivery not possible: %v", err), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusBadRequest)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
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
