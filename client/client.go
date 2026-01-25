package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/lnemenl/delivery-price-service/models"
)

// When the external API responds with 404
var ErrVenueNotFound = errors.New("venue not found")

// APIClient holds the configuration for connecting to Wolt
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New creates an API client with a 10-second timeout
func New(baseURL string, timeout time.Duration) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchVenueData retrieves both static and dynamic venue data from the API
func (c *APIClient) FetchVenueData(ctx context.Context, slug string) (models.VenueStatic, models.VenueDynamic, error) {
	// Create an errgroup derived from the parent context.
	// If one goroutine returns an error, 'groupCtx' will be canceled immediately.
	g, groupCtx := errgroup.WithContext(ctx)

	var (
		static  models.VenueStatic
		dynamic models.VenueDynamic
	)

	// Fetch Static Data
	g.Go(func() error {
		// Use groupCtx so this request is canceled if the other one fails
		urlStatic := fmt.Sprintf("%s%s/static", c.BaseURL, slug)
		if err := c.get(groupCtx, urlStatic, &static); err != nil {
			return fmt.Errorf("static data error: %w", err)
		}
		return nil
	})

	// Fetch Dynamic Data
	g.Go(func() error {
		urlDynamic := fmt.Sprintf("%s%s/dynamic", c.BaseURL, slug)
		if err := c.get(groupCtx, urlDynamic, &dynamic); err != nil {
			return fmt.Errorf("dynamic data error: %w", err)
		}
		return nil
	})

	// Wait blocks until all goroutines function have returned.
	// It returns the first non-nil error (if any).
	if err := g.Wait(); err != nil {
		return models.VenueStatic{}, models.VenueDynamic{}, err
	}

	return static, dynamic, nil
}

// get is a private helper that performs the HTTP Request with Context
func (c *APIClient) get(ctx context.Context, url string, target any) error {
	// Use NewRequestWithContext to enable cancellation
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		if resp.StatusCode == http.StatusNotFound {
			// Wrap error
			return fmt.Errorf("API 404: %w", ErrVenueNotFound)
		}
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	// Decode JSON into the target struct
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
