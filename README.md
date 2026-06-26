# 🚗 Roadlog

A self-hosted vehicle expense tracking system. Track fuel fill-ups, other expenses, and monitor consumption across all your vehicles.

## Features

- **Multi-vehicle support** — petrol, diesel, LPG, CNG, electric, hybrid
- **Smart fill-up form** — pre-fills last odometer, price, station (configurable)
- **Saved stations** — per fuel type, with autocomplete suggestions
- **Other expenses** — insurance, service, repair, tires, parking, tolls, etc.
- **Dashboard** — monthly spending charts (per vehicle, color-coded), period filters
- **Per-vehicle charts** — consumption trends, price trends
- **Multi-user** — shared vehicles, user management
- **CSV import** — with smart column mapping and auto-guess
- **CSV export & full backup/restore**
- **Dark/light theme**
- **Configurable currency** — EUR, USD, GBP, CZK, PLN, CHF, SEK, NOK, DKK, HUF
- **Dynamic units** — L, kWh, kg based on fuel type
- **Single container** — Go backend + embedded frontend, SQLite database

## Quick Start

```yaml
services:
  roadlog:
    image: stomko11/roadlog:latest
    container_name: roadlog
    ports:
      - "3000:3000"
    volumes:
      - ./data:/data
    restart: unless-stopped
```

```bash
docker compose up -d
```

Open http://localhost:3000

**Default credentials:** `admin@roadlog.local` / `roadlog`

> ⚠️ Change the default password after first login.

### Alternative registries

- Docker Hub: `stomko11/roadlog:latest`
- GitHub Container Registry: `ghcr.io/stomko11/roadlog:latest`

## Environment Variables

| Name | Description | Default |
|------|-------------|---------|
| `JWT_SECRET` | Secret for signing auth tokens | `roadlog-change-me-in-production` |
| `PORT` | Server port | `3000` |
| `DATA_DIR` | Directory for SQLite database | `/data` |

## Build from Source

Requirements: Go 1.22+

```bash
cd backend
go mod tidy
go build -o roadlog .
DATA_DIR=. ./roadlog
```

Open http://localhost:3000

## Tech Stack

- **Backend:** Go, Gin, GORM, SQLite
- **Frontend:** Vanilla JS (embedded single HTML file)
- **Container:** Alpine Linux (~20MB)

## Development

```bash
cd backend
DATA_DIR=. go run main.go
```

The frontend is embedded in `backend/static/index.html`. Changes require rebuilding the Go binary.

## License

MIT
