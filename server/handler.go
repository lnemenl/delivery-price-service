package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/service"
)

type PriceHandler struct {
	client *client.APIClient // I renamed this slot to 'client' to be clearer
}

// New creates the handler.
// We accept 'c' (the client) and put it into our struct.
func New(c *client.APIClient) *PriceHandler {
	return &PriceHandler{
		client: c,
	}
}

// HandleRequest is the main brain. It runs from top to bottom.
func (h *PriceHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {

	// ---------------------------------------------------------
	// 1. CHECK METHOD
	// ---------------------------------------------------------
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ---------------------------------------------------------
	// 2. READ INPUTS (The Query Parameters)
	// ---------------------------------------------------------
	// We read strings from the URL
	q := r.URL.Query()
	slugStr := q.Get("venue_slug")
	cartStr := q.Get("cart_value")
	latStr := q.Get("user_lat")
	lonStr := q.Get("user_lon")

	// ---------------------------------------------------------
	// 3. CONVERT INPUTS (String -> Number)
	// ---------------------------------------------------------
	if slugStr == "" {
		http.Error(w, "Missing venue_slug", http.StatusBadRequest)
		return
	}

	// Convert "1000" -> 1000
	cartValue, err := strconv.Atoi(cartStr)
	if err != nil {
		http.Error(w, "Invalid cart_value: must be an integer", http.StatusBadRequest)
		return
	}

	// Convert "60.17" -> 60.17
	userLat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid user_lat: must be a number", http.StatusBadRequest)
		return
	}

	userLon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid user_lon: must be a number", http.StatusBadRequest)
		return
	}

	// ---------------------------------------------------------
	// 4. FETCH DATA (Call the Client)
	// ---------------------------------------------------------
	// h.client refers to the pointer we stored in the struct
	staticData, dynamicData, err := h.client.FetchVenueData(slugStr)
	if err != nil {
		// Log error for us, send 500 to user
		fmt.Printf("API Error: %v\n", err)
		http.Error(w, "Failed to fetch venue data", http.StatusInternalServerError)
		return
	}

	// ---------------------------------------------------------
	// 5. CALCULATE (Call the Service)
	// ---------------------------------------------------------
	input := service.DeliveryInput{
		CartValue: cartValue,
		UserLat:   userLat,
		UserLon:   userLon,
	}

	priceResponse, err := service.CalculatePrice(input, staticData, dynamicData)
	if err != nil {
		// If math fails (too far), send 400
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---------------------------------------------------------
	// 6. SEND RESPONSE
	// ---------------------------------------------------------
	// Tell browser: "This is JSON data"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	// Write the JSON
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(priceResponse); err != nil {
		fmt.Printf("Encoding failed: %v\n", err)
	}
}
