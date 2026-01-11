package main

import (
	"fmt"
	"log"

	"github.com/lnemenl/wolt_1/client"
)

func main() {
	fmt.Println("--- Wolt Client Sanity Check ---")

	// 1. Create the Client (The Remote Control)
	api := client.New()

	// 2. Define a real venue slug from the README
	venueSlug := "home-assignment-venue-helsinki"

	// 3. Make the Call
	fmt.Printf("Fetching data for: %s...\n", venueSlug)
	staticData, dynamicData, err := api.FetchVenueData(venueSlug)

	if err != nil {
		log.Fatalf("CRASH! Failed to fetch data: %v", err)
	}

	// 4. Print the results to prove we got them
	fmt.Println("SUCCESS! Data received.")

	// Print coordinates from Static Data
	coords := staticData.VenueRaw.Location.Coordinates
	fmt.Printf("Venue Coordinates: %v\n", coords)

	// Print pricing rules from Dynamic Data
	pricing := dynamicData.VenueRaw.DeliverySpecs.DeliveryPricing
	fmt.Printf("Base Price: %d\n", pricing.BasePrice)
	fmt.Printf("Number of Distance Ranges: %d\n", len(pricing.DistanceRanges))
}
