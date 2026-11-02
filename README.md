# User Management

Full-stack user management app: React (TypeScript) frontend, Go REST backend (Handler → Service → Repository), MongoDB storage via the official MongoDB Go Driver.

```
backend/    Go API (cmd/server, internal/{config,model,apperror,repository,service,handler,middleware,server})
frontend/   React + TypeScript (Vite) SPA
docker-compose.yml
```

## Run with Docker Compose

```
cp .env.example .env     # adjust if needed
docker compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080 (`/healthz`, `/api/v1/users`)
- MongoDB runs in the `mongo` service with a named volume.

`GO_VERSION` in `.env` selects the Go base image; keep it equal to the `go` directive in `backend/go.mod`.

## Configuration (environment variables only)

Backend (`backend/.env.example`): `PORT`, `MONGODB_URI` (required), `MONGODB_DB` (required), `CORS_ALLOWED_ORIGIN` (required, the single allowed origin), `MAX_BODY_BYTES`, `LOG_LEVEL`.
Frontend (`frontend/.env.example`): `VITE_API_BASE_URL` (baked in at build time).
No credentials or connection details are hardcoded.

## Run locally

Backend (needs Go and a reachable MongoDB):
```
cd backend
export MONGODB_URI=mongodb://localhost:27017 MONGODB_DB=usermanagement CORS_ALLOWED_ORIGIN=http://localhost:5173
go run ./cmd/server
```
Frontend:
```
cd frontend
npm ci
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```

## API

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/users` | Create user (201; 409 duplicate email; 400 validation) |
| GET | `/api/v1/users?page&pageSize&search&status` | Paginated list, search (name/email), status filter |
| GET | `/api/v1/users/stats` | Dashboard counts (total/active/inactive) |
| GET | `/api/v1/users/{id}` | Details (400 malformed id, 404 missing) |
| PUT | `/api/v1/users/{id}` | Update (refreshes `updatedAt`) |
| PATCH | `/api/v1/users/{id}/status` | `{"status":"ACTIVE"|"INACTIVE"}` |
| DELETE | `/api/v1/users/{id}` | Permanent delete |

Responses use `{success, data, meta?, error?{code,message,details}}`. If MongoDB is unavailable the API returns 503 with a generic message.

## Testing

```
cd backend && go build ./... && go vet ./... && go test -count=1 ./...
cd frontend && npm ci && npm test && npm run build
```

Service, handler and middleware tests use an in-memory repository double. Repository and full-lifecycle integration tests run against a real MongoDB and are skipped unless `MONGODB_TEST_URI` is set, e.g.:
```
MONGODB_TEST_URI=mongodb://localhost:27017 go test -count=1 ./...
```
