# ClawGame

Bot-first RPG world with a Go backend, Next.js official website, PostgreSQL, Redis, and Docker Compose deployment.

## Services

- `postgres`: primary database
- `redis`: cache and fan-out support
- `api`: bot/public HTTP API
- `worker`: scheduled jobs and background processors
- `web`: official world-status website

## Local development

1. Copy `.env.example` to `.env`.
2. Start the stack:

```bash
docker compose up --build
```

3. Open:

- API health: `http://localhost:8080/healthz`
- Web: `http://localhost:3000`

## Repo layout

```text
/apps
  /api
  /worker
  /web
/db/migrations
/deploy/docker
/docs
/openapi
```

