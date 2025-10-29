# Country Currency API

A Go (Gin) service that fetches country metadata and exchange rates, stores them in a local SQLite database, and serves a REST API for querying countries and a generated summary image.

## Features

- **Refresh countries** from external APIs (countries + USD exchange rates)
- **Upsert** normalized country data into SQLite
- **Filter and sort** countries by region, currency, population, name, and estimated GDP
- **Lookup** country by name and **delete** by name
- **Status** endpoint exposing total countries and last refresh time
- **Generated summary image** (top countries by estimated GDP)

## Project Structure

- `cmd/server/` — application entrypoint (defines routes and server wiring)
- `internal/config/` — configuration loading and validation
- `internal/database/` — DB connection, schema/migrations, repository queries
- `internal/handlers/` — HTTP handlers (Gin)
- `internal/models/` — data models and API response structs
- `internal/services/` — external API client, country domain service, image generation
- `data/` — default location for the SQLite database file (gitignored)

## Tech Stack

- **Language:** Go 1.24+
- **Framework:** Gin
- **Database:** MySQL 8.0+
- **Driver:** github.com/go-sql-driver/mysql

## Getting Started

### Prerequisites

- Go 1.24+
- MySQL 8.0+ (running and accessible)
- MySQL client (for database creation)

### Database Setup

1. **Create MySQL database:**

```sql
CREATE DATABASE country_currency_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. **Create MySQL user (optional):**

```sql
CREATE USER 'country_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON country_currency_db.* TO 'country_user'@'localhost';
FLUSH PRIVILEGES;
```

### Install dependencies

```bash
go mod tidy
```

### Environment Variables

Create a `.env` file in the repository root (see `.env.example` for reference):

**Required:**
- `DB_HOST` — MySQL host (default: `localhost`)
- `DB_PORT` — MySQL port (default: `3306`)
- `DB_USER` — MySQL username (default: `root`)
- `DB_PASSWORD` — MySQL password (required)
- `DB_NAME` — MySQL database name (default: `country_currency_db`)

**Optional:**
- `PORT` — server port (default: `8080`)
- `COUNTRIES_API_URL` — countries API (default provided)
- `EXCHANGE_API_URL` — exchange rates API (default provided)

### Run

```bash
# Run the server directly (routes are defined in cmd/server/main.go)
go run ./cmd/server
```

**Quick start with nginx reverse proxy:**

```bash
# One-command setup (installs nginx config)
./setup-nginx.sh

# Then start your app
go run ./cmd/server

# Test through nginx on port 80
curl http://localhost/health
```

### Build

```bash
go build -o country-currency-api ./cmd/server
```

The first run will initialize the database and schema automatically.

## Database

- **MySQL** is used for data persistence
- Connection is created in `internal/database/db.go` via `Connect(user, password, host, port, dbName)`
- **Migrations** are executed automatically at startup:
  - Creates `countries` table with indices
  - Creates `metadata` table for tracking refresh timestamps
  - Seeds initial metadata
- Schema defined in `internal/database/schema.go`
- All queries are MySQL-compatible (using `ON DUPLICATE KEY UPDATE`, `NOW()`, etc.)

## External APIs

- Countries: `restcountries.com`
- Exchange rates: `open.er-api.com`

Requests are made via `internal/services/api_client.go` with a 30s timeout and JSON decoding.

## Handlers and Capabilities

Routes are declared in `cmd/server/main.go`. The following handlers expose functionality:

- `handlers.CountryHandler.RefreshCountries()`
  - Triggers refresh workflow via `services.CountryService.RefreshCountries()`
- `handlers.CountryHandler.GetAllCountries()`
  - Supports optional query params: `region`, `currency`, `sort`
  - Sorting options implemented in repository: `gdp_desc`, `gdp_asc`, `population_desc`, `population_asc`, `name_asc`, `name_desc`
- `handlers.CountryHandler.GetCountryByName()`
  - Retrieves a single country by name
- `handlers.CountryHandler.DeleteCountryByName()`
  - Deletes a country by name
- `handlers.CountryHandler.GetStatus()`
  - Returns total count and last refreshed time
- `handlers.CountryHandler.GetSummaryImage()`
  - Serves the generated PNG summary image from the image cache path

Refer to `cmd/server/main.go` to confirm exact route paths (e.g., `/countries`, `/countries/:name`, `/status`, `/summary-image`).

## API Endpoints

### 1. POST `/countries/refresh`
**Description:** Fetch all countries and exchange rates, then cache them in the database

```bash
curl -X POST http://localhost:8080/countries/refresh
```

**Success Response (200 OK):**
```json
{
  "message": "Countries refreshed successfully"
}
```

**Error Response (503 Service Unavailable):**
```json
{
  "error": "External data source unavailable",
  "details": "Could not fetch data from countries API: ..."
}
```

---

### 2. GET `/countries`
**Description:** Get all countries from database with optional filters and sorting

**Query Parameters:**
- `region` — Filter by region (e.g., `Africa`, `Europe`, `Asia`)
- `currency` — Filter by currency code (e.g., `NGN`, `USD`, `GBP`)
- `sort` — Sort order: `gdp_desc`, `gdp_asc`, `population_desc`, `population_asc`, `name_asc`, `name_desc`

**Examples:**

```bash
# Get all countries
curl http://localhost:8080/countries

# Get all African countries
curl "http://localhost:8080/countries?region=Africa"

# Get all countries using NGN currency
curl "http://localhost:8080/countries?currency=NGN"

# Get African countries sorted by GDP (descending)
curl "http://localhost:8080/countries?region=Africa&sort=gdp_desc"

# Get all countries sorted by population (ascending)
curl "http://localhost:8080/countries?sort=population_asc"

# Get European countries with EUR currency, sorted by name
curl "http://localhost:8080/countries?region=Europe&currency=EUR&sort=name_asc"
```

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Nigeria",
    "capital": "Abuja",
    "region": "Africa",
    "population": 206139589,
    "currency_code": "NGN",
    "exchange_rate": 1600.23,
    "estimated_gdp": 257674481.25,
    "flag_url": "https://flagcdn.com/ng.svg",
    "last_refreshed_at": "2025-10-22T18:00:00Z"
  },
  {
    "id": 2,
    "name": "Ghana",
    "capital": "Accra",
    "region": "Africa",
    "population": 31072940,
    "currency_code": "GHS",
    "exchange_rate": 15.34,
    "estimated_gdp": 30298345.21,
    "flag_url": "https://flagcdn.com/gh.svg",
    "last_refreshed_at": "2025-10-22T18:00:00Z"
  },
  {
    "id": 3,
    "name": "Antarctica",
    "capital": null,
    "region": "Polar",
    "population": 1000,
    "currency_code": null,
    "exchange_rate": null,
    "estimated_gdp": null,
    "flag_url": "https://flagcdn.com/aq.svg",
    "last_refreshed_at": "2025-10-22T18:00:00Z"
  }
]
```

**Note:** Fields like `capital`, `currency_code`, `exchange_rate`, `estimated_gdp`, and `flag_url` can be `null` if:
- Country has no capital
- Country has no currencies in the external API
- Currency code not found in exchange rates API
- Country with multiple currencies (only first is used)

---

### 3. GET `/countries/:name`
**Description:** Get a single country by name (case-insensitive)

```bash
# Get Nigeria
curl http://localhost:8080/countries/Nigeria

# Case-insensitive (works the same)
curl http://localhost:8080/countries/nigeria

# Get United States
curl http://localhost:8080/countries/"United%20States"
```

**Success Response (200 OK):**
```json
{
  "id": 1,
  "name": "Nigeria",
  "capital": "Abuja",
  "region": "Africa",
  "population": 206139589,
  "currency_code": "NGN",
  "exchange_rate": 1600.23,
  "estimated_gdp": 257674481.25,
  "flag_url": "https://flagcdn.com/ng.svg",
  "last_refreshed_at": "2025-10-22T18:00:00Z"
}
```

**Example with null values:**
```json
{
  "id": 15,
  "name": "Antarctica",
  "capital": null,
  "region": "Polar",
  "population": 1000,
  "currency_code": null,
  "exchange_rate": null,
  "estimated_gdp": null,
  "flag_url": "https://flagcdn.com/aq.svg",
  "last_refreshed_at": "2025-10-22T18:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Country not found"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation failed",
  "details": {
    "name": "is required"
  }
}
```

---

### 4. DELETE `/countries/:name`
**Description:** Delete a country record by name

```bash
# Delete Nigeria
curl -X DELETE http://localhost:8080/countries/Nigeria

# Delete with verbose output
curl -X DELETE -v http://localhost:8080/countries/Ghana
```

**Success Response (204 No Content):**
```
(Empty response body)
```

**Error Response (404 Not Found):**
```json
{
  "error": "Country not found"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation failed",
  "details": {
    "name": "is required"
  }
}
```

---

### 5. GET `/status`
**Description:** Get total countries count and last refresh timestamp

```bash
curl http://localhost:8080/status

# Pretty print with jq
curl http://localhost:8080/status | jq
```

**Success Response (200 OK):**
```json
{
  "total_countries": 250,
  "last_refreshed_at": "2025-10-22T18:00:00Z"
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "error": "Internal server error",
  "details": "..."
}
```

---

### 6. GET `/countries/image`
**Description:** Serve the generated summary image (PNG)

The image contains:
- Total number of countries
- Top 5 countries by estimated GDP
- Timestamp of last refresh

```bash
# Download the image
curl -O http://localhost:8080/countries/image

# Download with custom name
curl -o summary.png http://localhost:8080/countries/image

# View image info
curl -I http://localhost:8080/countries/image
```

**Success Response (200 OK):**
```
Content-Type: image/png
(Binary PNG data)
```

**Error Response (404 Not Found):**
```json
{
  "error": "Summary image not found"
}
```

---

## Complete Workflow Example

```bash
# 1. Check if server is running
curl http://localhost:8080/health

# 2. Refresh data from external APIs
curl -X POST http://localhost:8080/countries/refresh

# 3. Check status
curl http://localhost:8080/status

# 4. Get all African countries sorted by GDP
curl "http://localhost:8080/countries?region=Africa&sort=gdp_desc" | jq

# 5. Get specific country
curl http://localhost:8080/countries/Nigeria | jq

# 6. Download summary image
curl -O http://localhost:8080/countries/image

# 7. Get countries with specific currency
curl "http://localhost:8080/countries?currency=USD" | jq

# 8. Delete a country (testing)
curl -X DELETE http://localhost:8080/countries/TestCountry
```

## Notes and Implementation Details

- **Estimated GDP:** Calculated as `population × random(1000-2000) ÷ exchange_rate` (refresh generates new random multiplier each time)
- **Currency Handling:** Takes first currency from array; sets NULL if no currency or rate not found
- **Upsert Logic:** Case-insensitive name matching; updates existing records or inserts new ones
- **Image Generation:** Auto-generated after refresh showing top 5 countries by GDP
- **MySQL-specific:** All queries use MySQL syntax (`ON DUPLICATE KEY UPDATE`, `NOW()`, etc.)


