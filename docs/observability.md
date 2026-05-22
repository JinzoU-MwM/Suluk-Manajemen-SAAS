# Observability

This repository now uses the standalone Go service stack only.

## Scope

- Gateway and service health endpoints
- Structured application logs
- Service-level metrics

## Current expectations

- Every service exposes `/health`
- The API gateway is the main entry point for frontend traffic
- Logs should include request path, method, status, and duration when applicable

## Local infrastructure

Use:

```bash
docker compose -f deployments/docker-compose.yml up -d
```

This starts the supporting infrastructure for the Go services:

- PostgreSQL
- Redis
- NATS
- MinIO
