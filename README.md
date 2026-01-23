# Delivery Price Service

A Go microservice that calculates delivery fees based on user location and cart value.

## Features

- **Concurrent Data Fetching**: Retrieves static (location) and dynamic (pricing) data in parallel using `errgroup` for efficient error handling and request cancellation.
- **Geodetic Accuracy**: Uses the Haversine formula for precise distance calculations.
- **Robust Error Handling**: Distinguishes between client input errors (400) and system failures (500).
- **Production Ready**: Configured with strict timeouts and clean separation of concerns.

## Design Decisions

### Concurrency
I chose `errgroup` over `sync.WaitGroup` to orchestrate parallel API requests. This ensures that if the Static API call fails, the Dynamic API call is automatically canceled to save resources, and the first error is effectively propagated up the stack.

### Testing
The project uses `httptest.NewServer` to mock external dependencies. This allows for:
- Offline testing
- Deterministic results
- Simulation of edge cases (e.g., API returning 404 or malformed JSON)

## Running the Service

### Prerequisites
- Go 1.25 or higher

### Start Server
```bash
go run main.go
```
The server starts on port `8000`.

### API Usage

**Endpoint:** `GET /api/v1/delivery-order-price`

| Parameter    | Description                                                  |
|--------------|--------------------------------------------------------------|
| `venue_slug` | The ID of the venue                                          |
| `cart_value` | Value of the cart in cents                                   |
| `user_lat`   | User's latitude                                              |
| `user_lon`   | User's longitude                                             |

**Example Request:**
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=60.17094&user_lon=24.93087"
```

## Testing

### Unit Tests
Run the comprehensive test suite (including mocked API calls):
```bash
go test ./...
```

### Manual Verification
See [MANUAL_TESTS.md](MANUAL_TESTS.md) for a list of curl commands to test happy paths, edge cases, and error handling against a running server.

## Project Structure

- `client/`: External API interaction.
- `models/`: Shared data structures.
- `server/`: HTTP handler and routing.
- `service/`: Core logic (distance math, price calculation).
