# Bremer Abfallkalender API

[![CI](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/ci.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/ci.yml)
[![Build docker and push](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker.yml)
[![Docker hub image](https://img.shields.io/docker/image-size/larmic/abfallkalender_api?label=dockerhub)](https://hub.docker.com/repository/docker/larmic/abfallkalender_api)
![Docker Image Version (latest by date)](https://img.shields.io/docker/v/larmic/abfallkalender_api)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## What is this project?

An HTTP API that acts as a stable proxy in front of Bremen’s official waste collection calendar (Bremer Abfallkalender). The official service has no public, stable API. Instead, it serves data under a dynamic, time‑varying base URL. This project discovers that dynamic URL at runtime and exposes a minimal, predictable API on top.

Key capabilities:
- List all streets known to the official calendar
- List all house numbers for a given street
- Fetch the pickup calendar for a given street and house number (as ICS or CSV)
- Compute and return the next upcoming collection day and its waste types (JSON)
- Serve Prometheus metrics

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
  - Each house number contains minimal links to keep the payload small:
    - `_links.self` → calendar resource (default ICS in browsers; CSV via `Accept: text/csv`)
    - `_links.next` → JSON with the next collection day
  - Example snippet:
    ```json
    {
      "name": "Aachener Straße",
      "houseNumbers": [
        {
          "number": "22",
          "_links": {
            "self": {"href": "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe/number/22"},
            "next": {"href": "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe/number/22/next"}
          }
        }
      ],
      "_links": {"self": {"href": "https://your.host/abfallkalender-api/street/Aachener%20Stra%C3%9Fe"}}
    }
    ```

- GET `/abfallkalender-api/street/{street}/number/{number}`
  - Returns the pickup calendar for the address.
  - Content depends on the `Accept` header:
    - `Accept: text/html` → lightweight HTML preview showing the ICS content inline (no download); includes a link to `…/next`
    - `Accept: text/calendar` → upstream ICS content
    - `Accept: text/csv` → upstream CSV content
    - No `Accept` header → ICS by default (for CLI/cURL compatibility)

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

- GET `/livez` and `/readyz`
  - Kubernetes probes. Return `200 ok` as soon as the process can serve requests.
  - Neither performs an upstream call. See the Kubernetes section below for why.
  - Excluded from the request log so probes do not drown out real traffic.

The full OpenAPI description lives in `open-api-3.yaml`, is embedded into the binary at build time, and is served by the app at `/` and `/abfallkalender-api`.

## Quick start

### Docker

```bash
make docker-build
make docker-run
```

Your API is now available at `http://localhost:8080`. Run `make help` for all targets.

### Go (local)

```bash
make run
# or
go run .
```

The server listens on `:${PORT}` (defaults to `8080`).

## Releasing

Releases are cut from the GitHub Actions UI — no local tagging required:

1. Go to **Actions → Release → Run workflow**
2. Pick the bump level (`patch`, `minor` or `major`)

The workflow computes the next version from the latest tag, creates that tag
plus a GitHub release with generated notes, and pushes the image to Docker Hub
as `larmic/abfallkalender_api`, tagged with the version and `latest`, for
`linux/amd64`, `linux/arm64` and `linux/arm/v7`.

Tags carry no `v` prefix (`0.0.20`, not `v0.0.20`).

Releasing touches Docker Hub only. The AWS Lambda deployment is a separate,
deliberately manual path — see `infra/README.md`.

### Images

The `Dockerfile` cross-compiles: the builder stage always runs natively on the
build platform (`FROM --platform=$BUILDPLATFORM`) and Go targets the requested
architecture via `GOOS`/`GOARCH`. No QEMU emulation is involved, which keeps the
multi-arch build fast. It exposes two targets:

- `runner-standard` — plain image for K8s, Docker or a Raspberry Pi
- `runner-lambda` — adds the AWS Lambda Web Adapter as a Lambda extension

## Kubernetes

The published image runs as UID `65534` (nobody) and handles `SIGTERM` by
draining in-flight requests, so it works under a restricted Pod Security
Standard and survives rolling updates without dropping connections.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: abfallkalender-api
spec:
  replicas: 2
  selector:
    matchLabels:
      app: abfallkalender-api
  template:
    metadata:
      labels:
        app: abfallkalender-api
    spec:
      # Must stay above the application's own 15s shutdown timeout.
      terminationGracePeriodSeconds: 30
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: api
          image: larmic/abfallkalender_api:0.0.20
          ports:
            - name: http
              containerPort: 8080
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
          livenessProbe:
            httpGet:
              path: /livez
              port: http
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            periodSeconds: 10
          resources:
            requests:
              cpu: 10m
              memory: 32Mi
            limits:
              memory: 128Mi
---
apiVersion: v1
kind: Service
metadata:
  name: abfallkalender-api
spec:
  selector:
    app: abfallkalender-api
  ports:
    - name: http
      port: 80
      targetPort: http
```

Notes:

- **Pin the image tag.** `latest` exists but gives you no rollback target.
- **`readOnlyRootFilesystem: true` works** — the binary writes nothing to disk;
  the OpenAPI spec is embedded and the response cache lives in memory.
- **Readiness does not check the upstream.** This service is a caching proxy in
  front of `web.c-trace.de`. If the upstream fails, an upstream-coupled probe
  would take every replica out of the service simultaneously, even though cached
  responses (24 h TTL) could still be served. Liveness and readiness therefore
  report the same thing.
- **The cache is per pod.** Each replica keeps its own in-memory cache, so more
  replicas mean proportionally more upstream requests while caches warm up.
- **Scraping metrics**: `/metrics` is served on the same port.

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