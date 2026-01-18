# Delivery Order Price Calculator (DOPC) 🚀

A backend service that calculates delivery prices based on dynamic venue rules and user location.

## Quick Start
You need **Go 1.23+** installed.

### 1. Build and Run
```bash
go run main.go
```
The server will start on port `:8000`.

### 2. Verify with a Request
Open a new terminal and run this command:
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=60.17094&user_lon=24.93087"
```

**Expected Output:**
```json
{
  "total_price": 1190,
  "small_order_surcharge": 0,
  "cart_value": 1000,
  "delivery": {
    "fee": 190,
    "distance": 177
  }
}
```

### 3. Run Tests
The project includes Unit, Integration, and Mock tests.
```bash
go test ./...
```
*(Tests should complete instantly and pass)*

---

## Features & Implementation
*   **Implementation:** Validated against all requirements in the spec.
*   **Concurrency:** Fetches venue static and dynamic data **in parallel** to minimize latency.
*   **Architecture:** Separated into `server` (HTTP), `service` (Logic), and `client` (External API) layers.
*   **Safety:** Strict validation for inputs (coordinates, negative values) and robust error handling.
*   **Precision:** Uses Haversine formula for accurate distance calculations and integer math for all currency operations.

## Project Structure
*   `main.go`: Entry point.
*   `server/`: Request parsing and HTTP handling.
*   `service/`: Core business logic (Calculator).
*   `client/`: External API interaction (Parallel fetching).
*   `models/`: Data structures.
