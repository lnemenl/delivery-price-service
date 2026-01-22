package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lnemenl/delivery-price-service/models"
)

// APIClient holds the configuration for connecting to Wolt
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New creates an API client with a 10-second timeout
func New() *APIClient {
	return &APIClient{
		BaseURL: "https://consumer-api.development.dev.woltapi.com/home-assignment-api/v1/venues/",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchVenueData retrieves both static and dynamic venue data from the API
func (c *APIClient) FetchVenueData(slug string) (models.VenueStatic, models.VenueDynamic, error) {
	var (
		static  models.VenueStatic
		dynamic models.VenueDynamic
		wg      sync.WaitGroup
		errChan = make(chan error, 2)
	)

	wg.Add(2)

	// Fetch static data (venue location)
	go func() {
		defer wg.Done()
		urlStatic := c.BaseURL + slug + "/static"
		if err := c.get(urlStatic, &static); err != nil {
			errChan <- fmt.Errorf("static data error: %w", err)
		}
	}()

	// Fetch dynamic data (pricing and delivery specifications)
	go func() {
		defer wg.Done()
		urlDynamic := c.BaseURL + slug + "/dynamic"
		if err := c.get(urlDynamic, &dynamic); err != nil {
			errChan <- fmt.Errorf("dynamic data error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// If we encountered any errors, return the first one
	if err := <-errChan; err != nil {
		return models.VenueStatic{}, models.VenueDynamic{}, err
	}

	return static, dynamic, nil
}

// get is a private helper that performs the HTTP Request and Decodes the JSON
// target interface{} allows to pass ANY struct (Static or Dynamic) to be filled
func (c *APIClient) get(url string, target interface{}) error {
	// Create Request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Execute Request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check Status Code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	// Decode JSON into the target struct
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
