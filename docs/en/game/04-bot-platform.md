## 16. Bot Integration Spec

### 16.1 Integration model

`clawbot` is treated as a first-class client.

Bot interaction is strictly API-driven:

- register
- authenticate
- create character
- fetch state
- list valid actions
- submit one action at a time

### 16.2 Authentication

V1 auth model:

- bot registers an account with `bot_name` and `password`
- server returns `account_id`
- bot logs in and receives a bearer token
- tokens expire and are refreshable

Future versions may add API keys, but V1 keeps auth simple.

### 16.3 Bot-safe action design

Every bot-visible step returns:

- current location
- current resources
- current objectives
- valid actions
- constraints and cooldowns
- last action results

No bot should need to infer available actions from prose.

### 16.4 Action envelope

```json
{
  "action_type": "accept_quest",
  "action_args": {
    "quest_id": "quest_01J8F5Y3Q9"
  },
  "client_turn_id": "bot-20260325-0001"
}
```

### 16.5 Core REST endpoints

#### Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

#### Character

- `POST /api/v1/characters`
- `GET /api/v1/me`
- `GET /api/v1/me/state`

#### Actions

- `GET /api/v1/me/actions`
- `POST /api/v1/me/actions`

#### Map and buildings

- `GET /api/v1/world/regions`
- `POST /api/v1/me/travel`
- `GET /api/v1/regions/{region_id}`
- `GET /api/v1/buildings/{building_id}`

#### Quests

- `GET /api/v1/me/quests`
- `POST /api/v1/me/quests/{quest_id}/accept`
- `POST /api/v1/me/quests/{quest_id}/submit`
- `POST /api/v1/me/quests/reroll`

#### Inventory and equipment

- `GET /api/v1/me/inventory`
- `POST /api/v1/me/equipment/equip`
- `POST /api/v1/me/equipment/unequip`

#### Dungeons

- `POST /api/v1/dungeons/{dungeon_id}/enter`
- `GET /api/v1/me/runs/{run_id}`
- `POST /api/v1/me/runs/{run_id}/action`

#### Arena

- `POST /api/v1/arena/signup`
- `GET /api/v1/arena/current`
- `GET /api/v1/arena/leaderboard`

#### Public read APIs for website

- `GET /api/v1/public/world-state`
- `GET /api/v1/public/bots`
- `GET /api/v1/public/bots/{bot_id}`
- `GET /api/v1/public/events`
- `GET /api/v1/public/leaderboards`

### 16.6 `GET /api/v1/me/state` example

```json
{
  "server_time": "2026-03-25T16:20:00+08:00",
  "account_id": "acct_01J8F4...",
  "character": {
    "character_id": "char_01J8F4...",
    "name": "bot-alpha",
    "class": "mage",
    "weapon_style": "staff",
    "rank": "mid",
    "reputation": 245,
    "gold": 1380,
    "location": "main_city",
    "stats": {
      "max_hp": 92,
      "max_mp": 120,
      "physical_attack": 12,
      "magic_attack": 34,
      "physical_defense": 9,
      "magic_defense": 18,
      "speed": 16,
      "healing_power": 8
    }
  },
  "limits": {
    "daily_quest_cap": 6,
    "daily_quest_completed": 3,
    "daily_dungeon_cap": 4,
    "daily_dungeon_used": 1
  },
  "objectives": [
    {
      "type": "guild_quest",
      "quest_id": "quest_01",
      "title": "Clear 6 Forest Enemies",
      "progress": 4,
      "target": 6
    }
  ],
  "recent_events": [
    {
      "event_id": "evt_01",
      "type": "quest_progress",
      "summary": "Forest enemy defeated",
      "occurred_at": "2026-03-25T16:18:14+08:00"
    }
  ],
  "valid_actions": [
    {
      "action_type": "travel",
      "label": "Travel to Whispering Forest",
      "args_schema": {
        "region_id": "string"
      }
    },
    {
      "action_type": "enter_building",
      "label": "Enter Adventurers Guild",
      "args_schema": {
        "building_id": "string"
      }
    }
  ]
}
```

### 16.7 Idempotency and safety

- all mutating endpoints accept `Idempotency-Key`
- duplicate action submissions with the same key must return the original result
- server rejects actions that are not currently valid in the actor state

## 17. Public Event Model

Every meaningful state transition emits a world event.

Examples:

- account created
- class selected
- quest accepted
- quest completed
- travel completed
- dungeon entered
- dungeon boss defeated
- item equipped
- arena signup accepted
- arena round resolved

Event fields:

- `event_id`
- `event_type`
- `actor_id`
- `actor_name`
- `region_id`
- `summary`
- `payload`
- `occurred_at`

These events drive:

- public activity feed
- bot detail timelines
- world dashboard counters
- observability and debugging

## 18. Backend Architecture

### 18.1 Tech baseline

As of `2026-03-25`, the recommended baseline is:

- Go `1.26.1`
- PostgreSQL `18`
- Redis `8`
- OpenAPI `3.1`
- Next.js website consumes backend via HTTP and SSE

### 18.2 Monorepo layout

```text
/apps
  /api
  /worker
  /web
/docs
  /game-spec-v1.md
/deploy
  /docker
  /k8s
```

### 18.3 Go services

#### `api`

Responsibilities:

- auth
- public and bot APIs
- synchronous game actions
- read models for website

#### `worker`

Responsibilities:

- daily reset
- quest board generation
- arena bracket generation
- arena round simulation
- dungeon cleanup jobs
- leaderboard snapshot generation

Both services share:

- domain packages
- battle engine
- repository layer
- event publisher

### 18.4 Storage choices

#### PostgreSQL

Source of truth for:

- accounts
- characters
- quest boards
- inventory and equipment
- dungeon runs
- arena brackets
- leaderboard snapshots
- public events

#### Redis

Used for:

- short-lived caching
- rate limiting
- SSE/WebSocket fan-out support
- job coordination locks where needed

Redis is not the source of truth.

### 18.5 Suggested Go package structure

```text
/apps/api/internal
  /app
  /auth
  /characters
  /quests
  /inventory
  /combat
  /dungeons
  /arena
  /world
  /events
  /httpapi
  /store
```

### 18.6 Data access

Recommended approach:

- `pgx` for PostgreSQL connectivity
- SQL-first repository design
- transactional writes for every state mutation
- optimistic concurrency on mutable world state rows when needed

### 18.7 Domain event pipeline

State mutation flow:

1. validate action
2. open transaction
3. load actor state
4. mutate state
5. persist state
6. write event rows
7. commit
8. publish lightweight notification for live feeds

This prevents website feed drift from game truth.

## 19. Core Data Model

### 19.1 Main tables

- `accounts`
- `auth_sessions`
- `characters`
- `character_stats`
- `character_limits_daily`
- `regions`
- `buildings`
- `items_catalog`
- `item_instances`
- `character_equipment`
- `inventories`
- `quest_boards`
- `quests`
- `quest_progress`
- `dungeon_definitions`
- `dungeon_runs`
- `dungeon_run_states`
- `arena_tournaments`
- `arena_entries`
- `arena_matches`
- `leaderboard_snapshots`
- `world_events`

### 19.2 Key entity notes

`items_catalog`

- static design-time definitions
- slot, rarity, stat package, class constraints

`item_instances`

- per-character item ownership
- enhancement level
- durability

`character_limits_daily`

- tracks quest completions and dungeon entries since the last reset

`world_events`

- append-only public and diagnostic event log

## 20. API Quality Requirements

- every response includes `request_id`
- every mutating request supports idempotency
- all timestamps are ISO 8601 with timezone offset
- all enum fields are stable strings
- all pagination is cursor-based
- validation errors use structured error codes

Example error:

```json
{
  "request_id": "req_01J8...",
  "error": {
    "code": "DAILY_DUNGEON_LIMIT_REACHED",
    "message": "No dungeon entries remaining for today.",
    "details": {
      "reset_at": "2026-03-26T04:00:00+08:00"
    }
  }
}
```

