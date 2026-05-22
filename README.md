# Suluk Manajemen SAAS

Standalone Go microservices and Svelte frontend for travel operational management.

## Stack

- Backend: Go services in `cmd/` and `internal/`
- Frontend: Svelte app in `frontend-svelte/`
- Infra: Postgres, Redis, NATS, MinIO via `deployments/docker-compose.yml`

## Main Services

- `cmd/api-gateway`
- `cmd/auth-service`
- `cmd/package-service`
- `cmd/jamaah-service`
- `cmd/invoice-service`
- `cmd/finance-service`
- `cmd/vendor-service`
- `cmd/contract-service`
- `cmd/ai-ocr-service`

## Local setup

### 1. Infrastructure

```bash
docker compose -f deployments/docker-compose.yml up -d
```

### 2. Frontend

```bash
cd frontend-svelte
npm install
npm run dev
```

### 3. Backend services

Run each service in separate terminals as needed:

```bash
go run ./cmd/api-gateway
go run ./cmd/auth-service
go run ./cmd/package-service
go run ./cmd/jamaah-service
go run ./cmd/invoice-service
go run ./cmd/finance-service
go run ./cmd/vendor-service
go run ./cmd/contract-service
go run ./cmd/ai-ocr-service
```

Or use the helper targets in `Makefile`.

## Migrations

Apply service migrations with:

```bash
go run ./cmd/migration/main.go -service all -direction up
```

Supported services include:

- `auth`
- `package`
- `jamaah`
- `invoice`
- `finance`
- `aiocr`
- `vendor`
- `contract`

## Testing

### Go

```bash
go test ./...
```

### Frontend

```bash
cd frontend-svelte
npm test
npm run check
```

## Repository Notes

- The old Python/FastAPI backend has been removed from this standalone repository.
- Use `deployments/docker-compose.yml` instead of the deleted legacy root compose file.
