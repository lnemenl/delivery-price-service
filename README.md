# Delivery Order Price Calculator (DOPC) 🚀

A robust, production-ready backend service for calculating delivery order prices. 
Submitted for the Wolt Engineering Internship 2025.

## Quick Start

**Prerequisites:** Go 1.23+

### 1. Build and Run
You can run the service directly:
```bash
go run main.go
```
*The server will start on port `:8000`.*

### 2. Verify with a Request
Open a new terminal and run this command to simulate a delivery in Helsinki:
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=60.17094&user_lon=24.93087"
```

**Expected JSON Output:**
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

### 3. Run the Test Suite
This project includes Unit, Integration, and Mock tests.
```bash
go test -v ./...
```

---

## Design & Maintainability

The assignment emphasizes **correctness** and **maintainability**. Here is how this solution addresses them:

### Architecture (The "Clean" Approach)
I separated the application into three distinct layers to ensure modularity:
1.  **`server/` (HTTP Layer):** Handles parsing, input validation, and status codes. It acts as a guard, ensuring no invalid data reaches the core logic.
2.  **`service/` (Business Logic):** A pure logic layer. It calculates distances and fees using simple inputs. It has no dependencies on HTTP or the network, making it fast and easy to test.
3.  **`client/` (Infrastructure):** Wraps the external API calls. This allows us to use **Dependency Injection** to mock the Wolt API during testing, ensuring our tests don't fail just because the internet is down.

### Key Logic Decisions
* **Distance Calculation:** I implemented the **Haversine Formula**. While simple Euclidean geometry works on small scales, Haversine correctly accounts for the Earth's curvature, providing better accuracy at higher latitudes (like Finland).
* **Floating Point Math:** Money is handled strictly as **Integers** (cents) to avoid floating-point errors. Floats are only used for coordinate geometry.
* **Validation:** The service implements "Defensive Programming." It rejects negative values, impossible coordinates (Lat > 90), and massive request strings immediately with a `400 Bad Request`.

### Error Handling Strategy
* **400 Bad Request:** Client-side errors (e.g., "Distance too long", "Negative Cart Value").
* **404 Not Found:** Resource errors (e.g., "Venue Slug does not exist").
* **500 Internal Error:** System failures (e.g., Upstream API is unreachable).

---

## Project Structure

* `main.go`: Application entry point and dependency wiring.
* `client/`: HTTP client for fetching Static/Dynamic venue data.
* `server/`: HTTP Handlers and Validation logic.
* `service/`: Distance and Price calculators.
* `models/`: Shared JSON structs.

---

## API Reference

**Endpoint:** `GET /api/v1/delivery-order-price`

| Parameter | Type | Description |
| :--- | :--- | :--- |
| `venue_slug` | String | Unique ID of the venue. |
| `cart_value` | Int | Shopping cart value in **cents**. |
| `user_lat` | Float | User's latitude. |
| `user_lon` | Float | User's longitude. |
