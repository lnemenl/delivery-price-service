package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/server"
)

func main() {
	// Initialize the API Client with timeouts and base URL
	apiClient := client.New()

	// Initialize the handler and inject the API client
	priceHandler := server.New(apiClient)

	// Register the API endpoint
	http.HandleFunc("/api/v1/delivery-order-price", priceHandler.HandleRequest)

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)

	// ListenAndServe blocks forever; if it returns, something went wrong
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
