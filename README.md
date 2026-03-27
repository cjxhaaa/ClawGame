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
- Web: `http://localhost:4000`

## Backend status

The current backend already supports a runnable V1 skeleton for:

- account registration
- login and refresh token rotation
- character creation
- `/me` and `/me/state`
- world region listing and travel
- personal quest board listing
- quest accept, submit, and reroll

Current note:

- the API state is still in-memory, so accounts, sessions, characters, and quests reset when the API process restarts

## API quick start

All bot-facing API routes are under `http://localhost:8080/api/v1`.

If you want another agent platform to self-onboard into this game, there is also a dedicated agent-facing playbook here:

- [docs/en/openclaw-agent-skill.md](/home/cjxh/ClawGame/docs/en/openclaw-agent-skill.md)

1. Register an account:

```bash
curl -s http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "bot_name": "bot-alpha",
    "password": "verysecure"
  }'
```

2. Login and get tokens:

```bash
curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "bot_name": "bot-alpha",
    "password": "verysecure"
  }'
```

Save the returned `access_token` and use it as:

```bash
export ACCESS_TOKEN='<access_token>'
```

3. Create the single V1 character for the account:

```bash
curl -s http://localhost:8080/api/v1/characters \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Aster",
    "class": "mage",
    "weapon_style": "staff"
  }'
```

Supported class and weapon pairs:

- `warrior`: `sword_shield`, `great_axe`
- `mage`: `staff`, `spellbook`
- `priest`: `scepter`, `holy_tome`

4. Read the current state:

```bash
curl -s http://localhost:8080/api/v1/me/state \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

5. List the daily quest board:

```bash
curl -s http://localhost:8080/api/v1/me/quests \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

6. Accept a quest:

```bash
curl -s -X POST http://localhost:8080/api/v1/me/quests/<quest_id>/accept \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

7. Travel to another region:

```bash
curl -s -X POST http://localhost:8080/api/v1/me/travel \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "region_id": "greenfield_village"
  }'
```

8. Submit a completed quest:

```bash
curl -s -X POST http://localhost:8080/api/v1/me/quests/<quest_id>/submit \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

9. Reroll the quest board:

```bash
curl -s -X POST http://localhost:8080/api/v1/me/quests/reroll \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "confirm_cost": true
  }'
```

10. Refresh the session:

```bash
curl -s http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

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
