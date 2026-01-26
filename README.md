# Delivery Price Service

A Go microservice that calculates delivery fees. It takes a user's location and cart value, asks an external API for the specific venue's rules, and returns the total price.

## Prerequisites

- **Go 1.25** or later

## Getting Started

### 1. Run the Service
You can run the server directly. It starts on port `8000`.

```bash
go run .
```

### 2. Configuration
You can change settings using **command-line flags** or **environment variables**.
(Command-line flags always overwrite environment variables).
**Command-line Flags > Environment Variables > Default Values**

```bash
# Example: Change the port and set a 5-second timeout
go run . -port :9090 -timeout 5s
```

| Flag | Description |
|:---|:---|
| `-port` | Which port to listen on |
| `-base-url` | The external API address to get venue data |
| `-timeout` | How long to wait for the external API before giving up |

## Design & Architecture

### Why I organized the code this way
I split the code into specific layers so that every file has exactly one job.

* **The Server Layer (`server/`)**: This is the "Front Desk". It handles the internet traffic (HTTP) and checks if the user sent valid data (like ensuring the cart value is positive). It **does not** do any complex calculations.
* **The Service Layer (`service/`)**: This is the "Calculator". It contains the math logic. It takes inputs (distance, price) and returns the result.
* **The Client Layer (`client/`)**: This is the "Messenger". It knows how to talk to the external API.
* **The Models Layer (`models/`)**: Defines the data structures.

### Safe to add new features
Inside the `server/` directory, I separated the **Routing** (`endpoints.go`) from the **Logic** (`handler_delivery.go`).

**Why I did this:**
If I want to add a new feature later (like `GET /something-else`):
1.  I create a **new** handler file (e.g., `handler_time.go`).
2.  I wire it up in the configuration files (`handlers.go`, `endpoints.go`, `main.go`).
3.  **Crucially, I do not open or touch the existing `handler_delivery.go` file.**
This guarantees that adding a new feature cannot accidentally break the existing delivery price logic.

### Handling high traffic
This service is designed to be **Stateless**. This simply means the server has no "memory" of previous requests.

* **How it works:** When a request comes in, the server calculates the price and immediately forgets everything. It doesn't save user data in a variable or a file.
* **Why it helps:** Because the server doesn't need to remember anything, you can simply run multiple copies of this program at the same time to handle more traffic. Since they don't share memory, they won't confuse one user's order with another's.

### Concurrency
To get the price, I need two pieces of information from the external API: **Location** and **Pricing Rules**.
Instead of fetching them one by one which is slow, I used `errgroup` to fetch them **both at the same time**.

**Why it's safer:** `errgroup` connects these two tasks. If the "Location" request fails, the code automatically cancels the "Pricing" request immediately to save resources.

## Testing & Safety

### How to Run Tests
The project includes unit tests for business logic and integration tests that mock the external API to ensure reliability without network dependencies.
```bash
go test ./...
```

For manual testing scenarios, see [MANUAL_TESTS.md](MANUAL_TESTS.md).


## API Reference

### Get Delivery Price
**Endpoint:** `GET /api/v1/delivery-order-price`

**Example Request:**
```bash
curl "http://localhost:8000/api/v1/delivery-order-price?venue_slug=home-assignment-venue-helsinki&cart_value=1000&user_lat=60.17094&user_lon=24.93087"
```

**Status Codes**
- `200 OK`: Successful calculation.
- `400 Bad Request`: Invalid input (e.g., negative cart value) or delivery distance too long.
- `404 Not Found`: Venue slug does not exist.
- `502 Bad Gateway`: External API returned invalid or incomplete venue data.
- `500 Internal Server Error`: External API failure or data parsing error.

### Project Folders
* `client/`: Talks to the outside world (API).
* `models/`: Defines what the data looks like (JSON).
* `server/`: Handles the incoming HTTP requests.
* `service/`: Calculations.
