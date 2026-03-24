# URL Shortener

A production-grade URL shortening service built in Go. This project was built as a system design learning exercise, implementing real architectural patterns used at scale.

## What It Does

Users submit a long URL and receive a short code. When that short code is visited, the user is redirected to the original URL.

Two core actions:
- `POST /shorten` — convert a long URL into a short code
- `GET /{code}` — look up a short code and redirect the user
- `GET /health` — health check endpoint

## Architecture

```
                        ┌─────────────────┐
                        │   HTTP Server   │
                        │   (Go/net http) │
                        └────────┬────────┘
                                 │
               ┌─────────────────┼─────────────────┐
               │                 │                 │
        POST /shorten       GET /{code}       GET /health
               │                 │
               ▼                 ▼
    ┌──────────────────┐  ┌─────────────┐
    │  Rate Limiter    │  │ Check Redis │
    │  (Redis)         │  │   Cache     │
    └────────┬─────────┘  └──────┬──────┘
             │               Hit │  Miss
             ▼                   │    │
    ┌──────────────────┐         │    ▼
    │  Code Generator  │         │ ┌──────────┐
    │  (Random 6 char) │         │ │ Postgres │
    └────────┬─────────┘         │ └──────┬───┘
             │                   │        │
             ▼                   │        ▼
    ┌──────────────────┐         │  ┌──────────┐
    │    Postgres      │         │  │  Store   │
    │  (Persistent     │         │  │ in Redis │
    │    Storage)      │         │  └──────────┘
    └──────────────────┘         │
                                 ▼
                           Redirect User
```

## System Design Concepts Covered

### Caching (Cache-Aside Pattern)
Redis sits in front of Postgres for redirect lookups. On a cache miss the value is fetched from Postgres and stored in Redis with a 24 hour TTL. Subsequent requests for the same short code never touch Postgres. This pattern is used at scale to reduce database load on hot data.

### Distributed Rate Limiting
Rate limiting is implemented via Redis rather than in-memory. This means the limit is enforced correctly across multiple server instances behind a load balancer. Each IP address is allowed 10 requests per 60 second window. Uses Redis `INCR` which is atomic, avoiding race conditions.

### Persistent Storage vs Cache
Postgres is the source of truth. Redis is ephemeral — it can be wiped and rebuilt from Postgres at any time. This distinction between persistent and cache storage is fundamental to distributed system design.

### Database Schema Design
The `urls` table uses `BIGSERIAL` for the primary key (64-bit auto-increment), `VARCHAR(10)` for short codes with a `UNIQUE` constraint, and `TIMESTAMPTZ` for timezone-aware timestamps. The unique constraint automatically creates a B-tree index on `short_code`, making redirect lookups O(log n).

### Auto-Migrations
Schema migrations run automatically on startup using `golang-migrate`. Migrations are versioned, ordered, and reversible — every environment (local, staging, production) always has the correct schema without manual intervention.

### 301 vs 302 Redirects
The service uses `302 Found` (temporary redirect) intentionally. A `301` (permanent redirect) would be cached by browsers forever, preventing click tracking and making it impossible to update where a short code points. This is a product decision with real technical consequences.

### Infrastructure as Code
All infrastructure is defined in `docker-compose.yml`. Postgres and Redis are containerized with persistent volumes, meaning data survives container restarts. No manual installation required.

### Secrets Management
All credentials and configuration are loaded from environment variables at startup. Hardcoded credentials are a security vulnerability. In production these would be injected by a secrets manager like AWS Secrets Manager or HashiCorp Vault.

### Fail Fast on Misconfiguration
If required environment variables are missing the app refuses to start with a clear error message. A misconfigured app that starts anyway causes mysterious runtime bugs. Failing fast surfaces problems immediately.

### Single Responsibility Principle
Every package has one job:
- `internal/api` — HTTP handlers
- `internal/db` — database connection and migrations
- `internal/cache` — Redis operations
- `internal/shortener` — URL shortening business logic
- `internal/config` — configuration loading

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Database | PostgreSQL 15 |
| Cache | Redis 7 |
| Containerization | Docker / Docker Compose |
| DB Driver | pgx v5 |
| Redis Client | go-redis v9 |
| Migrations | golang-migrate |

## Getting Started

### Prerequisites
- Go 1.21+
- Docker and Docker Compose

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/url-shortener
cd url-shortener
```

2. Create your `.env` file:
```bash
cp .env.example .env
```

3. Start the infrastructure:
```bash
docker compose up -d
```

4. Run the server:
```bash
go run cmd/api/main.go
```

The server starts on the port defined in your `.env` file. Migrations run automatically on startup.

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | Postgres connection string | `postgres://admin:password@localhost:5432/url_shortener` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `PORT` | Server port | `8080` |

## API Reference

### Create Short URL
```
POST /shorten
Content-Type: application/json

{
  "url": "https://www.example.com/very/long/url"
}
```

Response:
```json
{
  "code": "aB3kQ2"
}
```

### Redirect
```
GET /{code}
```

Redirects to the original URL with `302 Found`.

### Health Check
```
GET /health
```

Returns `ok`.

## What Breaks At Scale

| Problem | Cause | Fix |
|---------|-------|-----|
| Single server overwhelmed | Too many requests for one machine | Multiple servers behind a load balancer |
| Postgres overwhelmed | Every redirect hits the DB | Redis cache (already implemented) |
| Redis goes down | All requests fall back to Postgres | Redis replication |
| Postgres goes down | Data loss | Replication and automated backups |
| Rate limit bypass | In-memory counters don't share state | Distributed rate limiting via Redis (already implemented) |

## Project Structure

```
url_shortener/
├── cmd/
│   └── api/
│       └── main.go               # Entry point — wires everything together
├── internal/
│   ├── api/
│   │   └── handlers.go           # HTTP handlers
│   ├── cache/
│   │   ├── redis.go              # Redis connection
│   │   ├── urls.go               # URL cache operations
│   │   └── ratelimit.go          # Distributed rate limiting
│   ├── config/
│   │   └── config.go             # Environment variable loading
│   ├── db/
│   │   ├── db.go                 # Postgres connection
│   │   ├── migrate.go            # Migration runner
│   │   └── migrations/
│   │       ├── 000001_create_urls_table.up.sql
│   │       └── 000001_create_urls_table.down.sql
│   └── shortener/
│       ├── generator.go          # Short code generation
│       └── store.go              # URL persistence
├── .env                          # Local environment variables (never commit)
├── .env.example                  # Template for environment variables
├── .gitignore
├── docker-compose.yml
├── go.mod
└── go.sum
```