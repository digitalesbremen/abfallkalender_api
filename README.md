# Bremer Abfallkalender API

[![Build backend](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/backend.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/backend.yml)
[![Build frontend](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/frontend.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/frontend.yml)
[![Build docker and push](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This project is still alpha and in active development.

---

## What is this project?

An HTTP API and tiny web component that act as a stable proxy in front of Bremen’s official waste collection calendar (Bremer Abfallkalender). The official service has no public, stable API. Instead, it serves data under a dynamic, time‑varying base URL. This project discovers that dynamic URL at runtime and exposes a minimal, predictable API on top.

Key capabilities:
- List all streets known to the official calendar
- List all house numbers for a given street
- Fetch the pickup calendar for a given street and house number (as ICS or CSV)
- Compute and return the next upcoming collection day and its waste types (JSON)
- Serve Prometheus metrics
- Serve a lightweight frontend web component (`kalender.js`)

Use cases:
- Integrate Bremen waste pickup schedules into home automation (Home Assistant, Node‑RED, etc.)
- Generate personal reminders (calendar subscriptions, notifications)
- Build dashboards without depending on hidden or unstable upstream URLs

## Why does this app exist?

The official Bremen waste calendar is implemented as a web app without a stable public API. Its base path changes via a redirect mechanism. This app stabilizes access by:
1. Discovering the current dynamic base URL of the official service
2. Calling the upstream JSON/ICS/CSV endpoints
3. Normalizing and returning responses via a small, well‑documented API

This avoids hard‑coding brittle upstream URLs in clients while enabling clean integrations.

## How it works (high level)

1. On incoming requests, the backend asks the official service for the current base URL (a HEAD request reveals a `Location` header). See `/misc/example/example-official-requests.http` for details.
2. Using that discovered base URL, the app queries upstream endpoints for streets, house numbers, and calendar files (ICS/CSV).
3. It returns structured data (HAL+JSON) or passes through calendar files depending on the route and the `Accept` header.
4. For the "next" endpoint, it parses the upstream CSV, finds the nearest future date, and classifies the waste types for that day.

No data is stored persistently; responses reflect current upstream content.

## API endpoints

Base path: your deployment domain. Examples below assume `https://your.host`.

- GET `/` and `/abfallkalender-api`
  - Returns the OpenAPI 3 specification (YAML) of this service.

- GET `/abfallkalender-api/streets`
  - Lists all available streets in Bremen.
  - Response: `application/json` (HAL style)
  - Example snippet:
    ```json
    {
      "_embedded": {
        "streets": [
          {
            "name": "Aachener Straße",
            "_links": {"self": {"href": "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe"}}
          }
        ]
      }
    }
    ```

- GET `/abfallkalender-api/street/{street}`
  - Returns the street and all available house numbers.
  - Response: `application/json` (HAL style)
  - Path parameter `street` must match the official spelling (URL‑encode umlauts/ß).

- GET `/abfallkalender-api/street/{street}/number/{number}`
  - Returns the pickup calendar for the address.
  - Content depends on the `Accept` header:
    - `Accept: text/calendar` → upstream ICS content
    - `Accept: text/csv` → upstream CSV content
    - No `Accept` header → ICS by default

- GET `/abfallkalender-api/street/{street}/number/{number}/next`
  - Returns the next upcoming collection day and the detected waste types.
  - Response (JSON):
    ```json
    {
      "day_of_collection": "2025-01-15",
      "garbage_types": ["yellow", "blue"]
    }
    ```
  - Possible waste types: `yellow`, `blue`, `brown`, `black`, `christmas`.

- GET `/metrics`
  - Exposes Prometheus metrics (`http_requests_total`, `http_request_duration_seconds`).

- GET `/kalender.js` and `/kalender.js.map`
  - Serves a small web component that can render a calendar widget in the browser.

The full OpenAPI description lives in `open-api-3.yaml` and is served by the app at `/` and `/abfallkalender-api`.

## Quick start

### Docker

```bash
docker build -t abfallkalender-api .
docker run --rm -p 8080:8080 -e PORT=8080 abfallkalender-api
```

Your API is now available at `http://localhost:8080`.

### Go (local)

```bash
make run
# or
go run ./...
```

The server listens on `:${PORT}` (defaults to `8080`).

## Examples (curl)

```bash
# All streets
curl -s https://your.host/abfallkalender-api/streets | jq .

# Street details incl. house numbers
curl -s "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe" | jq .

# ICS calendar
curl -s -H "Accept: text/calendar" \
  "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe/number/22"

# CSV calendar
curl -s -H "Accept: text/csv" \
  "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe/number/22"

# Next pickup
curl -s "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe/number/22/next" | jq .
```

Developer‑friendly HTTP files for IDE clients are available under `misc/example`.

## Metrics

Prometheus metrics are exposed at `/metrics` and already instrumented with request count and latency histograms. Add your Prometheus scrape config accordingly.

## Limitations and notes

- Upstream dependency: The app depends on Bremen’s official service being available. If the upstream format changes, this proxy may require updates.
- Exact spelling: Street and house number must match the upstream data. Use URL encoding for special characters (e.g., `Straße` → `Stra%C3%9Fe`).
- City scope: This project focuses on the city of Bremen.
- No persistence: Data is not cached permanently; every request reflects upstream responses.
- Formats: ICS and CSV are passed through from the official service; JSON is produced by this proxy for directory endpoints and the `next` computation.

## License

Apache License 2.0 — see `LICENSE`.