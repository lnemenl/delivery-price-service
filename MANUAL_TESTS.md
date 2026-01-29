# Manual Test Scenarios

These scenarios assume the server is running locally on port 8000.
Run the server first: `go run main.go`

## 1. Happy Path: Helsinki (Expected: 200 OK)
**Description:** Standard order in Helsinki, close to the venue, sufficient cart value.
- Venue: `home-assignment-venue-helsinki`
- Cart: €10.00 (1000 cents) -> Meets Minimum
- Distance: Short (~177m)
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=60.17094&user_lon=24.93087"
```
**Expected Response:**
```json
{"total_price":1190,"small_order_surcharge":0,"cart_value":1000,"delivery":{"fee":190,"distance":177}}
```

## 2. Small Order Surcharge (Expected: 200 OK)
**Description:** Cart value is below the minimum order value (€10.00). Surcharge should apply.
- Cart: €8.00 (800 cents)
- Surcharge: €2.00 (200 cents)
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=800&user_lat=60.17094&user_lon=24.93087"
```
**Expected Response:** `total_price` should be higher due to surcharge.
```json
{"total_price":1190,"small_order_surcharge":200,"cart_value":800,"delivery":{"fee":190,"distance":177}}
```

## 3. Distance Too Long (Expected: 400 Bad Request)
**Description:** User is very far from the venue (fake coordinates 0,0).
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=0&user_lon=0"
```
**Expected Response:** HTTP 400
`Delivery not possible: delivery distance too long: 7026040 meters` (Note: exact distance may vary slightly)

## 4. Venue Not Found (Expected: 404 Not Found)
**Description:** Requesting a venue that does not exist.
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=does-not-exist&cart_value=1000&user_lat=60.17&user_lon=24.93"
```
**Expected Response:** HTTP 404
`Venue not found`

## 5. Invalid Input (Expected: 400 Bad Request)
**Description:** Missing parameters (e.g., missing `cart_value`).
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&user_lat=60.17&user_lon=24.93"
```
**Expected Response:** HTTP 400
`Invalid input: missing cart_value`
