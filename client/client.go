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

var ErrVenueNotFound = errors.New("venue not found")

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

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
	errg, groupCtx := errgroup.WithContext(ctx)

	var (
		static  models.VenueStatic
		dynamic models.VenueDynamic
	)

	errg.Go(func() error {
		urlStatic := fmt.Sprintf("%s%s/static", c.BaseURL, slug)
		if err := c.get(groupCtx, urlStatic, &static); err != nil {
			return fmt.Errorf("static data error: %w", err)
		}
		return nil
	})

	errg.Go(func() error {
		urlDynamic := fmt.Sprintf("%s%s/dynamic", c.BaseURL, slug)
		if err := c.get(groupCtx, urlDynamic, &dynamic); err != nil {
			return fmt.Errorf("dynamic data error: %w", err)
		}
		return nil
	})

	// Wait blocks until all goroutines function have returned.
	// It returns the first non-nil error (if any).
	if err := errg.Wait(); err != nil {
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

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("API 404: %w", ErrVenueNotFound)
		}
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
