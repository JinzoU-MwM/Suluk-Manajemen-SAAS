# Contributing

## Development

### Frontend

```bash
cd frontend-svelte
npm install
```

### Backend

```bash
go mod download
```

### Infrastructure

```bash
docker compose -f deployments/docker-compose.yml up -d
```

## Running checks

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

## Code style

- Go: format with `gofmt`
- Svelte/JS: follow existing frontend conventions

## Commit messages

Use conventional commit style when practical:

- `feat:`
- `fix:`
- `refactor:`
- `docs:`
- `test:`
- `chore:`
