# Flight Price Monitor

Full-stack application that monitors flight prices and sends alerts when they drop below a target threshold. Built with Go, Next.js, PostgreSQL, and a complete observability stack.

## Tech Stack

| Layer         | Technology                                      |
|---------------|------------------------------------------------|
| Backend       | Go 1.26, net/http, PostgreSQL 16               |
| Frontend      | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| Flight Data   | SerpAPI (Google Flights)                        |
| Auth          | JWT                                             |
| Observability | Grafana, Loki, Tempo, Mimir (LGTM)             |
| Infra         | Docker Compose, Nginx, Grafana Alloy            |
| CI/CD         | GitHub Actions                                  |

## Architecture

```
                  ┌──────────┐
                  │  Nginx   │ :80
                  └────┬─────┘
                 ┌─────┴─────┐
           ┌─────┴──┐   ┌───┴────┐
           │  API   │   │  Web   │
           │ (Go)   │   │(Next.js)│
           │  :8080 │   │  :3000 │
           └───┬────┘   └────────┘
               │
        ┌──────┴──────┐
   ┌────┴───┐   ┌─────┴─────┐
   │Postgres│   │  SerpAPI  │
   │  :5432 │   │ (external)│
   └────────┘   └───────────┘

   ┌──────────────────────────────┐
   │     Observability (LGTM)     │
   │ Alloy → Loki / Tempo / Mimir│
   │         Grafana :3001        │
   └──────────────────────────────┘
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Or for local development: Go 1.26+, Node.js 22+, PostgreSQL 16

### Quick Start (Docker Compose)

```bash
cp .env.example .env
# Edit .env with your values (at minimum POSTGRES_PASSWORD and JWT_SECRET)

docker compose up --build
```

The app will be available at `http://localhost` (Nginx proxy), Grafana at `http://localhost:3001`.

### Local Development

**Backend:**

```bash
cd api
cp ../.env.example ../.env
# Set DATABASE_URL pointing to your local PostgreSQL

make run
# API starts on :8080
```

**Frontend:**

```bash
cd web
npm install
npm run dev
# App starts on :3000
```

### Environment Variables

| Variable          | Required | Default                  | Description                    |
|-------------------|----------|--------------------------|--------------------------------|
| `POSTGRES_USER`   | No       | `flightscanner`          | PostgreSQL username            |
| `POSTGRES_PASSWORD` | Yes    | -                        | PostgreSQL password            |
| `POSTGRES_DB`     | No       | `flightscanner`          | PostgreSQL database name       |
| `JWT_SECRET`      | Yes      | -                        | Secret for JWT signing         |
| `SERPAPI_KEY`     | No       | -                        | SerpAPI key for Google Flights |
| `SERVER_PORT`     | No       | `8080`                   | API server port                |
| `ENV`             | No       | `development`            | Environment (development/production) |
| `FRONTEND_URL`    | No       | `http://localhost:3000`  | Allowed CORS origin            |

## API Endpoints

### Authentication

| Method | Path              | Description          |
|--------|-------------------|----------------------|
| POST   | `/auth/register`  | Create user account  |
| POST   | `/auth/login`     | Login, returns JWT   |

### Routes (requires auth)

| Method | Path                      | Description             |
|--------|---------------------------|-------------------------|
| GET    | `/api/routes`             | List monitored routes   |
| POST   | `/api/routes`             | Create route to monitor |
| PUT    | `/api/routes/{id}`        | Update route            |
| DELETE | `/api/routes/{id}`        | Delete route            |
| PATCH  | `/api/routes/{id}/pause`  | Pause monitoring        |
| PATCH  | `/api/routes/{id}/resume` | Resume monitoring       |

### Search

| Method | Path                   | Description                  |
|--------|------------------------|------------------------------|
| POST   | `/api/search/flights`  | Search flights by route/date |

### History

| Method | Path                              | Description                   |
|--------|-----------------------------------|-------------------------------|
| GET    | `/api/routes/{id}/history`        | Price history (`?days=30`)    |
| GET    | `/api/routes/{id}/history/export` | Export as CSV/JSON            |

### Alerts

| Method | Path                          | Description        |
|--------|-------------------------------|--------------------|
| GET    | `/api/alerts`                 | List price alerts  |
| PATCH  | `/api/alerts/{id}/mark-read`  | Mark alert as read |

### Health

| Method | Path      | Description                     |
|--------|-----------|---------------------------------|
| GET    | `/health` | Health check + monitor count    |

## Project Structure

```
flight-scanner/
├── api/
│   ├── cmd/server/          # Entry point, graceful shutdown
│   ├── internal/
│   │   ├── auth/            # JWT service
│   │   ├── config/          # Env-based configuration
│   │   ├── database/        # PostgreSQL connection & migrations
│   │   ├── flightapi/       # SerpAPI client (Google Flights)
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # CORS, auth middleware
│   │   ├── models/          # Data models
│   │   ├── monitor/         # Background price monitoring workers
│   │   ├── repository/      # Data access layer
│   │   └── telemetry/       # OpenTelemetry setup
│   ├── Dockerfile
│   └── Makefile
├── web/
│   ├── src/
│   │   ├── app/             # Next.js app router
│   │   ├── components/      # UI component library
│   │   ├── lib/             # API client, types, utilities
│   │   └── modules/         # Feature modules (search, routes, alerts, history, auth)
│   └── Dockerfile
├── infra/
│   ├── nginx.conf           # Reverse proxy config
│   ├── grafana/             # Dashboards & provisioning
│   ├── alloy-config.alloy   # Telemetry collector
│   ├── loki-config.yaml     # Log aggregation
│   ├── tempo-config.yaml    # Distributed tracing
│   └── mimir-config.yaml    # Metrics storage
├── docker-compose.yml
└── .env.example
```

## Development

### Makefile Commands (api/)

```bash
make build          # Compile Go binary
make run            # Run server locally
make test           # Run tests with race detector
make lint           # Run golangci-lint
make fmt            # Format code
make tidy           # Tidy go.mod
make clean          # Remove build artifacts
make docker-build   # Build Docker image
```

### Frontend Commands (web/)

```bash
npm run dev         # Start dev server
npm run build       # Production build
npm run lint        # ESLint
npm run test        # Run Vitest tests
```

## Deployment

The project uses GitHub Actions for CI/CD:

1. **CI** (`.github/workflows/ci.yml`) — Runs Go and Node tests on push/PR to main
2. **Deploy** (`.github/workflows/deploy.yml`) — After CI passes, deploys via Docker Compose on a self-hosted runner

All services include health checks and auto-restart policies.

## Observability

The full LGTM stack is included in Docker Compose:

- **Grafana** (`localhost:3001`) — Pre-provisioned dashboards
- **Loki** — Log aggregation from all containers via Alloy
- **Tempo** — Distributed tracing with OpenTelemetry
- **Mimir** — Metrics storage and querying
- **Alloy** — Telemetry collector (logs, traces, metrics)

The API emits traces and metrics via OTLP to Alloy, which routes them to the appropriate backends.
