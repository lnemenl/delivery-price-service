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
	apiClient := client.New(cfg.BaseURL, cfg.Timeout)
	handlers := &server.Handlers{
		Delivery: server.NewDeliveryHandler(apiClient),
	}
	mux := http.NewServeMux()

	router := server.NewRouter(mux, handlers)
	router.Setup()

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on port %s...", cfg.Port)

	// ListenAndServe blocks forever. If it returns, something went wrong
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
