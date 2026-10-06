# Warehouse Management

Warehouse Management is a fullstack technical test submission for managing warehouse items and stock receiving. The application includes a Go REST API, PostgreSQL schema, Docker Compose setup, OpenAPI documentation, and a React frontend for item management.

## Architecture

The backend follows a layered structure:

- `handler`: HTTP routing, request parsing, handler-level validation, response envelopes, and error mapping.
- `service`: business rules such as SKU uniqueness, soft delete visibility, stock receive validation, and item/location existence checks.
- `repository`: PostgreSQL access using parameterized SQL queries.
- `model`: shared API/domain data structures and domain errors.
- `middleware`: request logging and CORS middleware.

The frontend uses React routes backed by TanStack Query hooks and small reusable UI components.

## Tech Stack

- Backend: Go, chi, pgx
- Database: PostgreSQL
- Frontend: React, TypeScript, Vite, React Router, TanStack Query, Tailwind CSS, shadcn-style local components
- Tooling: Docker, Docker Compose, OpenAPI 3.0

## Repository Structure

```text
.
├── backend/
│   ├── cmd/api/                 # API entrypoint
│   ├── internal/config/          # Environment configuration
│   ├── internal/handler/         # HTTP handlers and response envelope
│   ├── internal/middleware/      # Logging and CORS middleware
│   ├── internal/model/           # Data models and domain errors
│   ├── internal/repository/      # PostgreSQL queries
│   ├── internal/service/         # Business logic
│   ├── migrations/               # SQL migration
│   └── docs/openapi.yaml         # OpenAPI specification
├── frontend/
│   ├── src/components/           # UI and form components
│   ├── src/hooks/                # TanStack Query and utility hooks
│   ├── src/pages/                # Item page
│   ├── src/services/             # API client
│   └── src/types/                # TypeScript API types
├── docker-compose.yml
└── README.md
```

## Prerequisites

- Docker and Docker Compose
- Go 1.26 or compatible
- Node.js 20 or compatible
- npm

## Local Setup

Backend:

```bash
cd backend
cp .env.example .env
# For local backend outside Docker, set DATABASE_URL to a reachable PostgreSQL host.
# Example when PostgreSQL is exposed by Docker Compose:
# DATABASE_URL=postgres://warehouse:warehouse@localhost:5433/warehouse_management?sslmode=disable
go mod download
set -a; source .env; set +a
go run ./cmd/api
```

Frontend:

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

For local frontend development, `VITE_API_URL` can point directly to the backend, for example `http://localhost:8080`.

## Docker Setup

Run all services:

```bash
docker compose up --build
```

Service URLs:

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:5173`
- PostgreSQL: `localhost:5433`

The frontend container runs Vite with a proxy to the backend service.

## Environment Variables

Backend variables are documented in `backend/.env.example`:

```text
APP_ENV=development
PORT=8080
DATABASE_URL=postgres://warehouse:warehouse@postgres:5432/warehouse_management?sslmode=disable
CORS_ALLOWED_ORIGIN=http://localhost:5173
READ_HEADER_TIMEOUT_SECONDS=5
SHUTDOWN_TIMEOUT_SECONDS=10
```

Frontend variables are documented in `frontend/.env.example`:

```text
VITE_API_URL=http://localhost:8080
```

Docker Compose sets service-specific values directly in `docker-compose.yml`.

## Database Migration Setup

The initial migration is located at:

```text
backend/migrations/000001_init.sql
```

It creates:

- `items`
- `locations`
- `stock`

The schema includes primary keys, foreign keys, timestamps, active SKU uniqueness, soft delete support with `items.deleted_at`, stock uniqueness per item/location, and a nonnegative stock quantity constraint.

When using Docker Compose with an empty PostgreSQL volume, the migration is mounted into `/docker-entrypoint-initdb.d` and runs during database initialization. If the existing `postgres_data` volume already exists, recreate it to rerun initialization from scratch.

## API Documentation

OpenAPI 3.0 documentation is available at:

```text
backend/docs/openapi.yaml
```

## Main API Endpoints

All API responses use the standard envelope:

```json
{
  "success": true,
  "message": "Message",
  "data": {},
  "meta": {}
}
```

Implemented endpoints:

- `GET /health`
- `POST /api/v1/items`
- `GET /api/v1/items?category=&page=&limit=`
- `GET /api/v1/items/{id}`
- `PUT /api/v1/items/{id}`
- `DELETE /api/v1/items/{id}`
- `POST /api/v1/stock/receive`
- `GET /api/v1/stock/{item_id}`

Item deletion is a soft delete. Deleted items are excluded from normal list and detail queries.

## Frontend Setup

The frontend item management UI is available at:

```text
http://localhost:5173/items
```

Routes:

- `/items`
- `/items/new`
- `/items/:id/edit`

The UI supports item listing, category filtering, debounced name/SKU search, client-side pagination for fetched item data, create/edit modals, delete confirmation, loading states, and success/error notifications.

## Technical Decisions And Trade-Offs

1. **Layered backend architecture**: handlers, services, and repositories are separated to keep request validation, business rules, and SQL access in the right layers. This adds a small amount of structure but makes the code easier to review and test.

2. **Soft delete for items**: items use `deleted_at` instead of physical deletion. This keeps existing stock records and historical references intact while normal item list/detail queries only return active items.

3. **Minimal stock/location design**: the schema includes only `items`, `locations`, and `stock`. Stock is unique by `(item_id, location_id)`, and receiving stock increments the current quantity. Stock mutation logging is done at application level as required, without adding a separate transaction table.

4. **Validation and error handling**: request shape validation is done in handlers, while business rules such as duplicate SKU and stock receive checks live in services. Errors are mapped into a consistent response envelope with appropriate HTTP status codes.

5. **Frontend search and pagination**: the backend supports category filtering and pagination. The frontend fetches up to 100 items for item browsing, then applies debounced SKU/name search and page slicing client-side to keep the frontend simple for the technical-test scope.

## Manual Testing Instructions

Start the application:

```bash
docker compose up --build
```

Check health:

```bash
curl http://localhost:8080/health
```

Item flow:

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-001","name":"Keyboard","category":"Electronics","unit":"pcs"}'

curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-001","name":"Duplicate Keyboard","category":"Electronics","unit":"pcs"}'

curl "http://localhost:8080/api/v1/items?page=1&limit=10"
curl http://localhost:8080/api/v1/items/1

curl -X PUT http://localhost:8080/api/v1/items/1 \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-001","name":"Mechanical Keyboard","category":"Electronics","unit":"pcs"}'

curl -X DELETE http://localhost:8080/api/v1/items/1
curl "http://localhost:8080/api/v1/items?page=1&limit=10"
```

Stock flow requires an existing location. With the Docker database running, insert a sample location:

```bash
docker compose exec postgres psql -U warehouse -d warehouse_management \
  -c "INSERT INTO locations (code, zone, type) VALUES ('A-01', 'A', 'RACK') ON CONFLICT (code) DO NOTHING;"
```

Then test stock receiving:

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-002","name":"Mouse","category":"Electronics","unit":"pcs"}'

# Replace item_id with the id returned by the create item response if it differs.
curl -X POST http://localhost:8080/api/v1/stock/receive \
  -H "Content-Type: application/json" \
  -d '{"item_id":2,"location_id":1,"qty":10}'

# Replace the path id with the same item id used above if it differs.
curl http://localhost:8080/api/v1/stock/2
```

Expected checks:

- Duplicate SKU returns `409`.
- Invalid pagination such as `limit=101` returns `400`.
- Deleted items no longer appear in normal list/detail queries.
- Stock receive rejects nonpositive quantity.
- Stock receive logs the mutation at application level.

Frontend checks:

- Open `http://localhost:5173/items`.
- Create, edit, and delete an item.
- Verify category filter, debounced search, pagination, table/card layout toggle, loading states, and toast notifications.

## AI Usage Declaration

AI coding tools were used extensively during the development of this technical test, as permitted by the assessment guidelines.

### Tools Used

- ChatGPT
- Codex
- AI-assisted coding tools in the development environment

### AI-Assisted Areas

AI was used to assist with:

- Backend project scaffolding and Go implementation
- Handler, service, and repository implementation
- PostgreSQL schema and migration
- REST API implementation
- Middleware and response handling
- OpenAPI documentation
- Docker and Docker Compose configuration
- React and TypeScript implementation
- TanStack Query hooks and API integration
- UI components and forms
- Debugging and troubleshooting
- Testing and code review
- README documentation

### Candidate's Contribution

The candidate defined the implementation scope based on the technical test requirements, reviewed the generated implementation, validated business rules and API behavior, tested the application locally and with Docker Compose, and made the final technical decisions.

AI-generated code was reviewed, modified where necessary, and verified before submission. The candidate is responsible for the final implementation and its behavior.

## Known Limitations

- The application does not include authentication or authorization.
- Locations can be seeded through SQL, but there is no frontend or public CRUD API for managing locations.
- Stock mutation history is logged at application level as required; there is no stock transaction table.
- Frontend SKU/name search is client-side over fetched item data rather than a backend search endpoint.
