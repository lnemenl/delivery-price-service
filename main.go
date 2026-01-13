package main

import (
	"log"
	"net/http"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/server"
)

func main() {
	// 1. WIRE THE DEPENDENCIES
	// Create the API Client (The Tool)
	apiClient := client.New()

	// Create the Handler (The Worker) and give it the tool
	// This is "Dependency Injection" in its simplest form.
	priceHandler := server.New(apiClient)

	// 2. DEFINE THE ROUTES
	// We tell Go's default router:
	// "When a user hits '/api/v1/delivery-order-price', call priceHandler.HandleRequest"
	http.HandleFunc("/api/v1/delivery-order-price", priceHandler.HandleRequest)

	// 3. START THE SERVER
	// This is an infinite loop. It blocks here and listens for traffic.
	// ":8000" means "Listen on Port 8000 on this machine".
	log.Println("--- Wolt DOPC Service is running on http://localhost:8000 ---")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("Server crashed: ", err)
	}
}
