package client
// Declares that this file belongs to the 'client' package.

import (
    "encoding/json" // Library for converting JSON <-> Go Structs
    "fmt"           // Library for formatting strings/errors
    "net/http"      // Library for making network requests

    "github.com/lnemenl/wolt_1/models" // Importing our own models
)

// APIClient struct definition
type APIClient struct {
    BaseURL    string       // A simple string to hold "https://..."
    HTTPClient *http.Client // A pointer to a standard Go HTTP client engine.
                            // We use a pointer to share the connection pool.
}

// New is the constructor
func New() *APIClient {
    // Returns a pointer to a newly allocated APIClient
    return &APIClient{
        BaseURL: "https://consumer-api.development.dev.woltapi.com/home-assignment-api/v1/venues/",
        HTTPClient: &http.Client{
            Timeout: 0, // Explicitly set timeout to 0 (infinity) for now.
        },
    }
}

// FetchVenueData is a Public Method.
// Receiver: (c *APIClient). This means it can read 'c.BaseURL' and 'c.HTTPClient'.
// Arguments: slug (string).
// Returns: VenueStatic value, VenueDynamic value, error.
func (c *APIClient) FetchVenueData(slug string) (models.VenueStatic, models.VenueDynamic, error) {
    
    // Call the private helper method 'getStatic'
    // 'c.getStatic' works because 'getStatic' is also attached to *APIClient
    staticData, err := c.getStatic(slug)
    if err != nil {
        // If it fails, return empty structs and the error.
        return models.VenueStatic{}, models.VenueDynamic{}, err
    }

    // Call the private helper method 'getDynamic'
    dynamicData, err := c.getDynamic(slug)
    if err != nil {
        return models.VenueStatic{}, models.VenueDynamic{}, err
    }

    // Return both results successfully
    return staticData, dynamicData, nil
}

// getStatic is a Private Method (lowercase 'g').
// It is specific to VenueStatic.
func (c *APIClient) getStatic(slug string) (models.VenueStatic, error) {
    // Concatenate strings to build the full URL
    url := c.BaseURL + slug + "/static"
    
    // Create an empty variable of type VenueStatic
    var response models.VenueStatic
    
    // Call the generic helper 'makeRequest'.
    // We pass '&response' (address of our empty struct).
    // The makeRequest function will follow this address and fill the memory with data.
    err := c.makeRequest(url, &response)
    
    // Return the filled struct and any error
    return response, err
}

func (c *APIClient) getDynamic(slug string) (models.VenueDynamic, error) {
    url := c.BaseURL + slug + "/dynamic"
    var response models.VenueDynamic
    err := c.makeRequest(url, &response)
    return response, err
}

// makeRequest is the Generic Helper.
// Argument 'target': interface{}. This accepts ANY value.
// We expect the caller to pass a Pointer to a struct.
func (c *APIClient) makeRequest(url string, target interface{}) error {
    
    // 1. Execute the HTTP GET request using the client stored in struct 'c'
    resp, err := c.HTTPClient.Get(url)
    if err != nil {
        // If network fails (DNS, offline), wrap the error and return
        // %w wraps the original error so we can check it later
        return fmt.Errorf("failed to call API: %w", err)
    }
    
    // 2. Schedule the cleanup.
    // 'defer' pushes this function call onto a stack.
    // It executes LATER, exactly when 'makeRequest' returns.
    // resp.Body is a network stream. We MUST close it to free the TCP port.
    defer resp.Body.Close()

    // 3. Check HTTP Status
    if resp.StatusCode != http.StatusOK { // http.StatusOK is constant 200
        return fmt.Errorf("API returned status: %d", resp.StatusCode)
    }

    // 4. Decode JSON Stream
    // json.NewDecoder(resp.Body): Creates a stream reader attached to the network response.
    // .Decode(target): Reads the stream, parses JSON, and uses Reflection to write
    //                  data into the memory address pointed to by 'target'.
    if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
        return fmt.Errorf("failed to decode JSON: %w", err)
    }

    return nil
}
