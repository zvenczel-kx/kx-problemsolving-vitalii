# KX Problem Solution

## Overview
This solution consists of two main services:

1. **Gateway Service** – serves data to clients and shows the availability of Storage Services (0–3 may be available).  

   **Public API** (port 8080, exposed externally via Docker Compose):  
   * `GET http://localhost:8080/status` – returns the status of each Storage Service  
   * `GET http://localhost:8080/data` – retrieves dummy data from a Storage Service using round-robin and returns JSON  

   **Internal API** (port 9090, accessible via Docker DNS in the internal network):  
   * `POST /register` – register a Storage Service in the service registry  
   * `POST /heartbeat` – invoke periodically by each Storage Service to indicate it is alive  
   * `GET /healthz` – returns service health status  

2. **Storage Service** – registers itself with the Gateway Service on startup `/register` and periodically sends `/heartbeat` to indicate it is alive.  
   It stores dummy data in memory, accessible via `GET /data`, and exposes `GET /healthz` for health checks.  
   **Environment variable:** `STORAGE_NAME` – defines the name of the service for registration and dummy data.  

## Running the Services
docker compose -f docker-compose.yaml up -d --build

## Running unit tests for the Gateway Service
Inside of 'gateway' folder run next command: go test -v ./... 

## Behavior When No Storage Services Are Running

The Gateway Service responds with HTTP 503 'Service Unavailable'. Detailed information can be found in container logs.

## Future Improvements

* Add TLS to the Gateway Service (Public API) for HTTPS support. Internal API and Storage Services remain HTTP.

* Implement JWT authentication for the Gateway Service (Public API).

* Add a rate limiter to protect against DDoS attacks.

* Set up observability: logs, traces (Jaeger), metrics (Prometheus), alerts, SLO / SLI / SLA

* Add more unit tests with coverage, end-to-end (E2E) tests, and performance tests (K6).

* Create a Makefile to make project builds and running tests more convenient.