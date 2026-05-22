# CLAUDE.md

Repository guidance for the standalone Suluk Manajemen SAAS codebase.

## Primary stack

- Backend: Go microservices in `cmd/` and `internal/`
- Frontend: Svelte app in `frontend-svelte/`
- Infra: `deployments/docker-compose.yml`

## Useful commands

### Go

```powershell
go build ./cmd/...
go test ./...
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

### Migrations

```powershell
go run ./cmd/migration/main.go -service all -direction up
```

### Frontend

```bash
cd frontend-svelte
npm install
npm run dev
npm run test
npm run check
```

### Infra

```bash
docker compose -f deployments/docker-compose.yml up -d
```

## Notes

- The legacy Python/FastAPI backend has been removed from this repository.
- Do not reintroduce `backend/` references in docs, CI, or deployment files.
