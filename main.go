package main

import (
	"log"
	"net/http"
	"time"

	"github.com/lnemenl/delivery-price-service/client"
	"github.com/lnemenl/delivery-price-service/server"
)

func main() {
	// Initialize the API Client with a 10-second timeout
	apiClient := client.New()

	// Initialize the handler and inject the API client
	priceHandler := server.New(apiClient)

	// Initialize a new ServeMux to isolate routes
	mux := http.NewServeMux()

	// Register the API endpoint
	mux.HandleFunc("/api/v1/delivery-order-price", priceHandler.HandleRequest)

	// Configure the server with timeouts to ensure reliability
	port := ":8000"
	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the server
	log.Printf("Server starting on port %s...", port)

	// ListenAndServe blocks forever; if it returns, something went wrong
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
