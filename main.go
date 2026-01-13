package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/server"
)

func main() {
	// 1. Initialize the API Client
	// This sets up the HTTP client with timeouts and the base URL
	apiClient := client.New()

	// 2. Initialize the Handler
	// We inject the client into the handler
	// It allows the handler to use the client without creating it itself
	priceHandler := server.New(apiClient)

	// 3. Register the Route
	// We map the specific URL path to our handler function
	http.HandleFunc("/api/v1/delivery-order-price", priceHandler.HandleRequest)

	// 4. Start the Server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)

	// ListenAndServe blocks forever. If it returns, something went wrong (like port in use)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
