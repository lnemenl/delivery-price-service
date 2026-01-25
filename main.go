package main

import (
	"log"
	"net/http"
	"time"

	"github.com/lnemenl/delivery-price-service/client"
	"github.com/lnemenl/delivery-price-service/server"
)

func main() {

	cfg := LoadConfig()

	// Initialize the API Client with a 10-second timeout
	apiClient := client.New(cfg.BaseURL, cfg.Timeout)

	// Initialize the handler and inject the API client
	priceHandler := server.New(apiClient)

	// Initialize a new ServeMux to isolate routes
	mux := http.NewServeMux()

	router := server.NewRouter(mux, priceHandler)
	router.Setup()

	// Configure the server with timeouts to ensure reliability
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the server
	log.Printf("Server starting on port %s...", cfg.Port)

	// ListenAndServe blocks forever; if it returns, something went wrong
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
