# Tasks

## 1. Implementation

- [x] 1.1 Create backend/ with go.mod (latest Go version, mongo-driver v2, testify) and cmd/server/main.go. main.go loads the config, connects to MongoDB with a ping, ensures the unique email index, starts the server and shuts down gracefully on SIGINT/SIGTERM.
- [x] 1.2 backend/internal/config/config.go: env-based config with PORT, MONGODB_URI, MONGODB_DB, CORS_ALLOWED_ORIGIN, MAX_BODY_BYTES and LOG_LEVEL. Fail fast if a required value is missing. Add backend/.env.example.
- [x] 1.3 backend/internal/model/user.go: the User struct with bson/json tags, status constants, and the request/response DTOs.
- [x] 1.4 backend/internal/apperror/: typed errors (validation, not found, conflict, unavailable) and mapping to HTTP status. Mongo errors map as follows:
- [x] 1.5 backend/internal/repository/user_repository.go: a UserRepository interface and a Mongo implementation with Create, GetByID, List (BSON filter, regexp.QuoteMeta search, skip/limit, sort by createdAt desc), Update, UpdateStatus, Delete and Stats (counts). Every method takes a ctx. Add EnsureIndexes.
- [x] 1.6 backend/internal/service/user_service.go: validation (required fields, email via net/mail, phone format, length limits, trimmed and lowercased email), ObjectID parsing, status validation, and timestamp handling.
- [x] 1.7 backend/internal/handler/ (users handler, response helpers) and middleware/ (CORS for the single origin, http.MaxBytesReader, request logging with slog, panic recovery). Wire the routes in backend/internal/server/router.go.
- [x] 1.8 Backend tests (see the Test plan).
- [x] 1.9 backend/Dockerfile: multi-stage build, non-root user, with GO_VERSION as a build arg.
- [x] 1.10 Create frontend/ with Vite, React and TS. It contains:
- [x] 1.11 frontend/Dockerfile: a Node build stage, then nginx serving the static build. VITE_API_BASE_URL is passed as a build arg.
- [x] 1.12 Frontend tests with Vitest + RTL (see the Test plan), and a vitest config in vite.config.ts.
- [x] 1.13 docker-compose.yml with mongo, backend and frontend services, a mongo healthcheck, a named volume, and variables read from a root .env.example. Add .gitignore and .dockerignore files.
- [x] 1.14 Update README.md with setup, configuration, the Compose command, running locally, and testing instructions.
