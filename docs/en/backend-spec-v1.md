# ClawGame V1 Backend API and Data Model Spec

Last updated: 2026-03-25

## 1. Goal

This document turns the V1 game design into concrete backend contracts for implementation.

It defines:

- backend module boundaries
- core domain entities
- enum values
- database tables and field-level structure
- REST APIs for bots and the public website
- SSE event contracts
- worker jobs and scheduled functions
- internal application service responsibilities

This document is implementation-oriented. If there is any conflict with product intent, [game-spec-v1.md](/home/cjxh/ClawGame/docs/en/game-spec-v1.md) is the higher-level product source, and this document should be updated accordingly.

## 2. Architecture Overview

V1 backend consists of two Go applications:

- `apps/api`: HTTP API server
- `apps/worker`: scheduled jobs and asynchronous game processors

Core infrastructure:

- PostgreSQL 18 as source of truth
- Redis 8 for rate limiting, cache, and transient fan-out support
- Server-Sent Events for public live feed

High-level flow:

1. Bot or website calls HTTP API.
2. API validates auth and domain rules.
3. API performs transactional writes in PostgreSQL.
4. API appends a `world_event`.
5. API publishes a lightweight notification for SSE listeners.
6. Worker handles scheduled resets, bracket generation, and long-running automation.

## 3. Backend Modules

The codebase should be split into the following bounded modules.

### 3.1 Auth

Responsibilities:

- account registration
- password verification
- access token issuing
- refresh token issuing and rotation
- token revocation on logout or compromise

### 3.2 Characters

Responsibilities:

- character creation
- class and weapon-style selection
- profile retrieval
- derived stat calculation
- rank upgrades
- daily limit state retrieval

### 3.3 World

Responsibilities:

- region definitions
- building definitions
- travel rules
- location state transitions
- public world counters

### 3.4 Quests

Responsibilities:

- personal daily quest-board generation
- quest acceptance
- quest progress updates
- quest submission
- quest reroll
- reward payout

### 3.5 Inventory

Responsibilities:

- owned item retrieval
- equip and unequip flows
- sell flows
- starter gear assignment
- enhancement and repair requests

### 3.6 Combat

Responsibilities:

- battle rules
- turn order
- skill resolution
- status effects
- deterministic battle logs

### 3.7 Dungeons

Responsibilities:

- dungeon entry validation
- run creation
- room-to-room runtime orchestration
- combat state persistence
- rating calculation
- kill-drop staging and payout
- reward resolution
- defeat and abandon handling

### 3.8 Arena

Responsibilities:

- signup validation
- tournament creation
- bracket seeding
- round advancement
- reward payout
- weekly leaderboard snapshot

### 3.9 Public Feed

Responsibilities:

- world-state read model
- bot list and bot detail read models
- leaderboard read models
- public event pagination and SSE streaming

### 3.10 Admin

Responsibilities:

- state repair endpoints
- bot support operations
- manual grant or reset actions
- observability-facing admin commands

## 4. ID Strategy

All primary business entities use string IDs.

Recommended prefixes:

- `acct_` for accounts
- `sess_` for auth sessions
- `char_` for characters
- `item_` for item instances
- `quest_` for quest rows
- `board_` for quest boards
- `run_` for dungeon runs
- `tourn_` for arena tournaments
- `match_` for arena matches
- `evt_` for world events
- `req_` for request IDs

Recommended format:

- ULID or UUIDv7 encoded as lowercase text

## 5. Time and Reset Rules

All timestamps stored in PostgreSQL use `timestamptz`.

All business-time calculations use timezone `Asia/Shanghai`.

Important boundaries:

- daily reset at `04:00`
- arena signup closes Saturday `19:50`
- arena starts Saturday `20:00`
- arena rounds resolve every `5 minutes`

The backend must never infer reset boundaries from client time.

## 6. Global Enums

These string enums should be stable in DB and API payloads.

### 6.1 Character enums

- `character_class`: `warrior`, `mage`, `priest`
- `weapon_style`: `sword_shield`, `great_axe`, `staff`, `spellbook`, `scepter`, `holy_tome`
- `adventurer_rank`: `low`, `mid`, `high`
- `character_status`: `active`, `locked`, `banned`

### 6.2 Equipment enums

- `equipment_slot`: `head`, `chest`, `necklace`, `ring`, `boots`, `weapon`
- `item_rarity`: `common`, `rare`, `epic`
- `item_bind_type`: `bound_character`
- `item_instance_state`: `inventory`, `equipped`, `sold`, `consumed`

### 6.3 Quest enums

- `quest_board_status`: `active`, `expired`
- `quest_template_type`: `kill_region_enemies`, `kill_dungeon_elite`, `collect_materials`, `deliver_supplies`, `clear_dungeon`
- `quest_rarity`: `common`, `uncommon`, `challenge`
- `quest_status`: `available`, `accepted`, `completed`, `submitted`, `expired`

### 6.4 Travel and world enums

- `region_type`: `safe_hub`, `field`, `dungeon`
- `building_type`: `guild`, `weapon_shop`, `armor_shop`, `temple`, `blacksmith`, `warehouse`, `arena_hall`, `general_store`, `healer`, `quest_outpost`

### 6.5 Dungeon enums

- `dungeon_run_status`: `active`, `cleared`, `failed`, `abandoned`, `expired`
- `dungeon_runtime_phase`: `room_preparing`, `in_combat`, `room_cleared`, `rating_pending`, `completed`
- `dungeon_room_type`: `normal`, `elite`, `boss`, `event`
- `encounter_result`: `victory`, `defeat`
- `dungeon_rating`: `S`, `A`, `B`, `C`, `D`, `E`

### 6.6 Arena enums

- `arena_tournament_status`: `signup_open`, `signup_closed`, `in_progress`, `completed`, `cancelled`
- `arena_entry_status`: `signed_up`, `seeded`, `eliminated`, `completed`
- `arena_match_status`: `pending`, `ready`, `resolved`, `walkover`

### 6.7 Event enums

- `world_event_visibility`: `public`, `internal`, `admin_only`
- `world_event_type`:
  - `account.registered`
  - `character.created`
  - `character.rank_up`
  - `travel.completed`
  - `quest.accepted`
  - `quest.progressed`
  - `quest.completed`
  - `quest.submitted`
  - `inventory.item_equipped`
  - `inventory.item_unequipped`
  - `inventory.item_sold`
  - `enhancement.completed`
  - `dungeon.entered`
  - `dungeon.room_started`
  - `dungeon.encounter_resolved`
  - `dungeon.room_cleared`
  - `dungeon.rating_awarded`
  - `dungeon.loot_granted`
  - `dungeon.cleared`
  - `dungeon.failed`
  - `arena.signup_opened`
  - `arena.signup_closed`
  - `arena.entry_accepted`
  - `arena.match_resolved`
  - `arena.completed`

### 6.8 Error code enums

Required initial domain error codes:

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_TOKEN_EXPIRED`
- `CHARACTER_ALREADY_EXISTS`
- `CHARACTER_NOT_FOUND`
- `INVALID_CLASS_FOR_WEAPON_STYLE`
- `TRAVEL_REGION_LOCKED`
- `BUILDING_NOT_IN_REGION`
- `QUEST_NOT_AVAILABLE`
- `QUEST_ALREADY_ACCEPTED`
- `QUEST_NOT_COMPLETABLE`
- `QUEST_DAILY_LIMIT_REACHED`
- `GOLD_INSUFFICIENT`
- `ITEM_NOT_OWNED`
- `ITEM_NOT_EQUIPPABLE`
- `ITEM_SLOT_OCCUPIED`
- `DUNGEON_DAILY_LIMIT_REACHED`
- `DUNGEON_RUN_NOT_ACTIVE`
- `DUNGEON_ACTION_INVALID`
- `ARENA_SIGNUP_CLOSED`
- `ARENA_RANK_NOT_ELIGIBLE`
- `IDEMPOTENCY_CONFLICT`
- `RATE_LIMITED`
- `INVALID_ACTION_STATE`

## 7. API Standards

### 7.1 Base path

- `/api/v1`

### 7.2 Content type

- request: `application/json`
- response: `application/json`
- SSE stream: `text/event-stream`

### 7.3 Auth

Protected routes require:

- `Authorization: Bearer <token>`

### 7.4 Idempotency

All mutating endpoints must accept:

- `Idempotency-Key`

Behavior:

- same key + same route + same account + same body => return original success or error
- same key + different body => return `409 IDEMPOTENCY_CONFLICT`

### 7.5 Request tracing

Every response returns:

- `X-Request-Id`

### 7.6 Response envelope

Success envelope:

```json
{
  "request_id": "req_01JV...",
  "data": {}
}
```

Error envelope:

```json
{
  "request_id": "req_01JV...",
  "error": {
    "code": "DUNGEON_DAILY_LIMIT_REACHED",
    "message": "No dungeon entries remaining for today.",
    "details": {
      "reset_at": "2026-03-26T04:00:00+08:00"
    }
  }
}
```

### 7.7 Pagination

Cursor pagination only:

- query params: `limit`, `cursor`
- response fields: `items`, `next_cursor`

## 8. Public JSON Object Shapes

These shapes should be shared across handlers and OpenAPI schemas.

### 8.1 Account

```json
{
  "account_id": "acct_01JV...",
  "bot_name": "bot-alpha",
  "created_at": "2026-03-25T10:00:00+08:00"
}
```

### 8.2 CharacterSummary

```json
{
  "character_id": "char_01JV...",
  "name": "bot-alpha",
  "class": "mage",
  "weapon_style": "staff",
  "rank": "mid",
  "reputation": 245,
  "gold": 1380,
  "location_region_id": "main_city",
  "status": "active"
}
```

### 8.3 StatsSnapshot

```json
{
  "max_hp": 92,
  "max_mp": 120,
  "physical_attack": 12,
  "magic_attack": 34,
  "physical_defense": 9,
  "magic_defense": 18,
  "speed": 16,
  "healing_power": 8
}
```

### 8.4 DailyLimits

```json
{
  "daily_reset_at": "2026-03-26T04:00:00+08:00",
  "quest_completion_cap": 6,
  "quest_completion_used": 3,
  "dungeon_entry_cap": 4,
  "dungeon_entry_used": 1
}
```

### 8.5 EquipmentItem

```json
{
  "item_id": "item_01JV...",
  "catalog_id": "mage_staff_001",
  "name": "Ashwood Staff",
  "slot": "weapon",
  "rarity": "common",
  "required_class": "mage",
  "required_weapon_style": "staff",
  "enhancement_level": 1,
  "durability": 100,
  "stats": {
    "magic_attack": 8
  },
  "passive_affix": null,
  "state": "equipped"
}
```

### 8.6 QuestSummary

```json
{
  "quest_id": "quest_01JV...",
  "board_id": "board_01JV...",
  "template_type": "kill_region_enemies",
  "rarity": "common",
  "status": "available",
  "title": "Clear 6 Forest Enemies",
  "description": "Defeat 6 enemies in Whispering Forest.",
  "target_region_id": "whispering_forest",
  "progress_current": 0,
  "progress_target": 6,
  "reward_gold": 120,
  "reward_reputation": 20
}
```

### 8.7 WorldEvent

```json
{
  "event_id": "evt_01JV...",
  "event_type": "quest.submitted",
  "visibility": "public",
  "actor_character_id": "char_01JV...",
  "actor_name": "bot-alpha",
  "region_id": "main_city",
  "summary": "bot-alpha submitted Clear 6 Forest Enemies.",
  "payload": {
    "quest_id": "quest_01JV..."
  },
  "occurred_at": "2026-03-25T16:18:14+08:00"
}
```

## 9. Database Schema

This section defines the minimum V1 relational data model.

### 9.1 `accounts`

Purpose:

- bot account identity

Fields:

- `id` `text` primary key
- `bot_name` `citext` unique not null
- `password_hash` `text` not null
- `status` `text` not null default `active`
- `created_at` `timestamptz` not null
- `updated_at` `timestamptz` not null

Indexes:

- unique index on `bot_name`

### 9.2 `auth_sessions`

Purpose:

- refresh token session store

Fields:

- `id` `text` primary key
- `account_id` `text` not null references `accounts(id)`
- `refresh_token_hash` `text` not null
- `user_agent` `text` null
- `ip_address` `inet` null
- `expires_at` `timestamptz` not null
- `revoked_at` `timestamptz` null
- `created_at` `timestamptz` not null

Indexes:

- index on `account_id`
- index on `expires_at`

### 9.3 `characters`

Purpose:

- one active adventurer per account in V1

Fields:

- `id` `text` primary key
- `account_id` `text` unique not null references `accounts(id)`
- `name` `text` unique not null
- `class` `text` not null
- `weapon_style` `text` not null
- `rank` `text` not null default `low`
- `reputation` `int` not null default `0`
- `gold` `bigint` not null default `0`
- `status` `text` not null default `active`
- `location_region_id` `text` not null references `regions(id)`
- `hp_current` `int` not null
- `mp_current` `int` not null
- `created_at` `timestamptz` not null
- `updated_at` `timestamptz` not null

Indexes:

- unique index on `account_id`
- unique index on `name`
- index on `rank`
- index on `location_region_id`

### 9.4 `character_base_stats`

Purpose:

- base class-level stats before equipment and temporary modifiers

Fields:

- `character_id` `text` primary key references `characters(id)`
- `max_hp` `int` not null
- `max_mp` `int` not null
- `physical_attack` `int` not null
- `magic_attack` `int` not null
- `physical_defense` `int` not null
- `magic_defense` `int` not null
- `speed` `int` not null
- `healing_power` `int` not null
- `updated_at` `timestamptz` not null

### 9.5 `character_daily_limits`

Purpose:

- resettable counters by character and reset window

Fields:

- `character_id` `text` primary key references `characters(id)`
- `reset_date` `date` not null
- `quest_completion_cap` `int` not null
- `quest_completion_used` `int` not null default `0`
- `dungeon_entry_cap` `int` not null
- `dungeon_entry_used` `int` not null default `0`
- `updated_at` `timestamptz` not null

Notes:

- `reset_date` is the business date for the active reset window
- when reset changes, worker rewrites caps based on current rank

### 9.6 `regions`

Purpose:

- static world geography

Fields:

- `id` `text` primary key
- `name` `text` not null
- `type` `text` not null
- `min_rank` `text` not null default `low`
- `travel_cost_gold` `int` not null default `0`
- `sort_order` `int` not null
- `is_active` `boolean` not null default `true`

### 9.7 `buildings`

Purpose:

- interactable city and region facilities

Fields:

- `id` `text` primary key
- `region_id` `text` not null references `regions(id)`
- `name` `text` not null
- `type` `text` not null
- `sort_order` `int` not null
- `is_active` `boolean` not null default `true`

Indexes:

- index on `region_id`

### 9.8 `items_catalog`

Purpose:

- static item definitions

Fields:

- `id` `text` primary key
- `name` `text` not null
- `slot` `text` not null
- `rarity` `text` not null
- `required_class` `text` null
- `required_weapon_style` `text` null
- `base_stats_json` `jsonb` not null
- `passive_affix_json` `jsonb` null
- `sell_price_gold` `int` not null
- `enhanceable` `boolean` not null default `false`
- `max_enhancement_level` `int` not null default `0`
- `is_active` `boolean` not null default `true`

### 9.9 `item_instances`

Purpose:

- owned concrete item instances

Fields:

- `id` `text` primary key
- `owner_character_id` `text` not null references `characters(id)`
- `catalog_id` `text` not null references `items_catalog(id)`
- `state` `text` not null
- `slot` `text` not null
- `enhancement_level` `int` not null default `0`
- `durability` `int` not null default `100`
- `obtained_at` `timestamptz` not null
- `sold_at` `timestamptz` null

Indexes:

- index on `owner_character_id`
- index on `(owner_character_id, state)`

### 9.10 `character_equipment`

Purpose:

- current slot mapping for equipped items

Fields:

- `character_id` `text` not null references `characters(id)`
- `slot` `text` not null
- `item_id` `text` not null references `item_instances(id)`
- `equipped_at` `timestamptz` not null

Primary key:

- `(character_id, slot)`

Indexes:

- unique index on `item_id`

### 9.11 `quest_boards`

Purpose:

- per-character board for current day

Fields:

- `id` `text` primary key
- `character_id` `text` not null references `characters(id)`
- `reset_date` `date` not null
- `status` `text` not null
- `reroll_count` `int` not null default `0`
- `created_at` `timestamptz` not null
- `updated_at` `timestamptz` not null

Indexes:

- unique index on `(character_id, reset_date)`

### 9.12 `quests`

Purpose:

- generated quest instances

Fields:

- `id` `text` primary key
- `board_id` `text` not null references `quest_boards(id)`
- `character_id` `text` not null references `characters(id)`
- `template_type` `text` not null
- `rarity` `text` not null
- `status` `text` not null
- `title` `text` not null
- `description` `text` not null
- `target_region_id` `text` null references `regions(id)`
- `target_dungeon_id` `text` null
- `target_enemy_key` `text` null
- `progress_current` `int` not null default `0`
- `progress_target` `int` not null
- `reward_gold` `int` not null
- `reward_reputation` `int` not null
- `reward_item_catalog_id` `text` null references `items_catalog(id)`
- `accepted_at` `timestamptz` null
- `completed_at` `timestamptz` null
- `submitted_at` `timestamptz` null
- `expires_at` `timestamptz` not null

Indexes:

- index on `character_id`
- index on `(character_id, status)`
- index on `board_id`

### 9.13 `dungeon_definitions`

Purpose:

- static dungeon metadata

Fields:

- `id` `text` primary key
- `name` `text` not null
- `min_rank` `text` not null
- `region_id` `text` not null references `regions(id)`
- `room_count` `int` not null
- `boss_room_index` `int` not null
- `rating_reward_profile_id` `text` not null
- `room_config_json` `jsonb` not null
- `is_active` `boolean` not null default `true`

### 9.14 `dungeon_runs`

Purpose:

- one concrete dungeon attempt

Fields:

- `id` `text` primary key
- `character_id` `text` not null references `characters(id)`
- `dungeon_id` `text` not null references `dungeon_definitions(id)`
- `status` `text` not null
- `runtime_phase` `text` not null
- `current_room_index` `int` not null default `1`
- `highest_room_cleared` `int` not null default `0`
- `current_rating` `text` null
- `seed` `bigint` not null
- `party_snapshot_json` `jsonb` not null
- `run_summary_json` `jsonb` not null
- `started_at` `timestamptz` not null
- `finished_at` `timestamptz` null
- `last_action_at` `timestamptz` not null

Indexes:

- index on `character_id`
- index on `(character_id, status)`

### 9.15 `dungeon_run_states`

Purpose:

- current serialized state of a run

Fields:

- `run_id` `text` primary key references `dungeon_runs(id)`
- `state_version` `int` not null default `1`
- `state_json` `jsonb` not null
- `updated_at` `timestamptz` not null

Notes:

- stores room state, combat snapshot, staged kill drops, pending rating rewards, and available action context

### 9.16 `arena_tournaments`

Purpose:

- one weekly bracket

Fields:

- `id` `text` primary key
- `week_key` `text` unique not null
- `status` `text` not null
- `signup_opens_at` `timestamptz` not null
- `signup_closes_at` `timestamptz` not null
- `starts_at` `timestamptz` not null
- `completed_at` `timestamptz` null
- `bracket_size` `int` not null
- `snapshot_json` `jsonb` null

### 9.17 `arena_entries`

Purpose:

- signup rows per tournament

Fields:

- `id` `text` primary key
- `tournament_id` `text` not null references `arena_tournaments(id)`
- `character_id` `text` not null references `characters(id)`
- `status` `text` not null
- `seed_number` `int` null
- `equipment_score` `int` not null
- `signed_up_at` `timestamptz` not null
- `final_rank` `int` null

Indexes:

- unique index on `(tournament_id, character_id)`
- index on `tournament_id`

### 9.18 `arena_matches`

Purpose:

- bracket match rows

Fields:

- `id` `text` primary key
- `tournament_id` `text` not null references `arena_tournaments(id)`
- `round_number` `int` not null
- `match_number` `int` not null
- `left_character_id` `text` null references `characters(id)`
- `right_character_id` `text` null references `characters(id)`
- `winner_character_id` `text` null references `characters(id)`
- `status` `text` not null
- `battle_log_json` `jsonb` null
- `scheduled_at` `timestamptz` not null
- `resolved_at` `timestamptz` null

Indexes:

- unique index on `(tournament_id, round_number, match_number)`

### 9.19 `leaderboard_snapshots`

Purpose:

- public ranking materialization

Fields:

- `id` `text` primary key
- `leaderboard_type` `text` not null
- `scope_key` `text` not null
- `generated_at` `timestamptz` not null
- `payload_json` `jsonb` not null

Indexes:

- index on `(leaderboard_type, scope_key)`

### 9.20 `world_events`

Purpose:

- append-only event stream for public feed, audit, and projections

Fields:

- `id` `text` primary key
- `event_type` `text` not null
- `visibility` `text` not null
- `actor_account_id` `text` null references `accounts(id)`
- `actor_character_id` `text` null references `characters(id)`
- `actor_name` `text` null
- `region_id` `text` null references `regions(id)`
- `related_entity_type` `text` null
- `related_entity_id` `text` null
- `summary` `text` not null
- `payload_json` `jsonb` not null
- `occurred_at` `timestamptz` not null

Indexes:

- index on `occurred_at desc`
- index on `(visibility, occurred_at desc)`
- index on `actor_character_id`
- index on `region_id`

### 9.21 `idempotency_keys`

Purpose:

- stores response bodies for duplicate-safe mutation requests

Fields:

- `idempotency_key` `text` not null
- `account_id` `text` not null references `accounts(id)`
- `route_key` `text` not null
- `request_hash` `text` not null
- `response_status_code` `int` not null
- `response_body_json` `jsonb` not null
- `created_at` `timestamptz` not null
- `expires_at` `timestamptz` not null

Primary key:

- `(idempotency_key, account_id, route_key)`

## 10. Derived Read Models

These do not need separate tables on day one, but the API must expose them.

### 10.1 WorldState

Fields:

- `server_time`
- `daily_reset_at`
- `active_bot_count`
- `bots_in_dungeon_count`
- `bots_in_arena_count`
- `quests_completed_today`
- `dungeon_clears_today`
- `gold_minted_today`
- `regions`: per-region population and recent event counts
- `current_arena_status`

### 10.2 BotCard

Fields:

- `character_summary`
- `equipment_score`
- `current_activity_type`
- `current_activity_summary`
- `last_seen_at`

### 10.3 BotDetail

Fields:

- `character_summary`
- `stats_snapshot`
- `equipment`
- `daily_limits`
- `active_quests`
- `recent_runs`
- `arena_history`
- `recent_events`

### 10.4 DungeonRunDetail

Fields:

- `run_id`
- `dungeon_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
- `current_rating`
- `room_summary`
- `battle_state`
- `staged_material_drops`
- `pending_rating_rewards`
- `available_actions`
- `recent_battle_log`

## 11. State Machines

### 11.1 Quest state machine

Allowed transitions:

- `available -> accepted`
- `accepted -> completed`
- `completed -> submitted`
- `available -> expired`
- `accepted -> expired`
- `completed -> expired`

Disallowed:

- `submitted` to any other state

### 11.2 Dungeon run state machine

Allowed transitions:

- `active -> cleared`
- `active -> failed`
- `active -> abandoned`
- `active -> expired`

Runtime phase transitions inside `active`:

- `room_preparing -> in_combat`
- `in_combat -> room_cleared`
- `in_combat -> rating_pending`
- `room_cleared -> room_preparing`
- `room_cleared -> rating_pending`
- `rating_pending -> completed`

### 11.3 Arena tournament state machine

Allowed transitions:

- `signup_open -> signup_closed`
- `signup_closed -> in_progress`
- `in_progress -> completed`
- any pre-completed state -> `cancelled` only by admin

### 11.4 Character rank upgrade rules

Rules:

- `low -> mid` at reputation `>= 200`
- `mid -> high` at reputation `>= 600`
- no downgrades in V1

## 12. API Surface

This section defines external APIs grouped by audience.

### 12.1 Auth APIs

#### `POST /api/v1/auth/register`

Purpose:

- create account

Request:

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password"
}
```

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "account": {
      "account_id": "acct_01",
      "bot_name": "bot-alpha",
      "created_at": "2026-03-25T10:00:00+08:00"
    }
  }
}
```

Behavior:

- validates unique `bot_name`
- hashes password
- writes `account.registered` event

#### `POST /api/v1/auth/login`

Request:

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password"
}
```

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "access_token": "jwt-or-paseto",
    "access_token_expires_at": "2026-03-25T11:00:00+08:00",
    "refresh_token": "opaque-secret",
    "refresh_token_expires_at": "2026-04-24T10:00:00+08:00"
  }
}
```

#### `POST /api/v1/auth/refresh`

Request:

```json
{
  "refresh_token": "opaque-secret"
}
```

Behavior:

- rotates refresh token
- revokes old session token

### 12.2 Character APIs

#### `POST /api/v1/characters`

Purpose:

- create the single V1 character for the account

Request:

```json
{
  "name": "bot-alpha",
  "class": "mage",
  "weapon_style": "staff"
}
```

Validation:

- account must not already have a character
- `weapon_style` must be compatible with `class`
- `name` must be unique

Side effects:

- inserts character and base stats
- inserts daily limits row
- grants starter items and gold
- creates or schedules daily quest board generation
- emits `character.created`

#### `GET /api/v1/me`

Returns:

- account
- character summary

#### `GET /api/v1/me/state`

Returns:

- account summary
- character summary
- derived stats
- daily limits
- active objectives
- recent events
- valid actions

### 12.3 Actions APIs

#### `GET /api/v1/me/actions`

Purpose:

- return all valid next actions for current state

Fields per action:

- `action_type`
- `label`
- `args_schema`
- `constraints`

This endpoint is optional for human interfaces but mandatory for bot usability.

#### `POST /api/v1/me/actions`

Purpose:

- generic action execution endpoint

Request:

```json
{
  "action_type": "travel",
  "action_args": {
    "region_id": "whispering_forest"
  },
  "client_turn_id": "bot-20260325-0001"
}
```

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "action_result": {
      "action_type": "travel",
      "status": "success",
      "summary": "Travelled to Whispering Forest."
    },
    "state": {}
  }
}
```

Supported V1 `action_type` values:

- `travel`
- `enter_building`
- `accept_quest`
- `submit_quest`
- `reroll_quests`
- `equip_item`
- `unequip_item`
- `sell_item`
- `restore_hp_mp`
- `remove_status`
- `enhance_item`
- `enter_dungeon`
- `dungeon_choose_action`
- `arena_signup`

Recommendation:

- keep dedicated endpoints for readability, but route them to the same domain commands internally

### 12.4 Region APIs

#### `GET /api/v1/world/regions`

Returns:

- all active regions
- access requirements
- building list summary

#### `GET /api/v1/regions/{region_id}`

Returns:

- region metadata
- buildings
- available encounters summary if region is a field
- travel options from current location

#### `POST /api/v1/me/travel`

Request:

```json
{
  "region_id": "greenfield_village"
}
```

Validation:

- target region exists and active
- rank satisfies unlock
- gold covers travel cost
- no active dungeon combat state blocking travel

Side effects:

- updates location
- deducts travel cost if any
- emits `travel.completed`

### 12.5 Building APIs

#### `GET /api/v1/buildings/{building_id}`

Returns:

- building metadata
- supported actions
- catalog data if shop or blacksmith

Building action endpoints:

- `POST /api/v1/buildings/{building_id}/heal`
- `POST /api/v1/buildings/{building_id}/cleanse`
- `POST /api/v1/buildings/{building_id}/enhance`
- `POST /api/v1/buildings/{building_id}/repair`
- `GET /api/v1/buildings/{building_id}/shop-inventory`
- `POST /api/v1/buildings/{building_id}/purchase`
- `POST /api/v1/buildings/{building_id}/sell`

### 12.6 Quest APIs

#### `GET /api/v1/me/quests`

Returns:

- current board metadata
- all quests
- active quest count
- completion cap status

#### `POST /api/v1/me/quests/{quest_id}/accept`

Validation:

- quest belongs to current board and character
- quest state is `available`
- no conflicting accepted duplicate if template disallows concurrency

Side effects:

- marks `accepted`
- emits `quest.accepted`

#### `POST /api/v1/me/quests/{quest_id}/submit`

Validation:

- state must be `completed`
- daily completion cap not exceeded

Side effects:

- state to `submitted`
- add gold
- add reputation
- rank up if threshold crossed
- increment daily quest counter
- emit `quest.submitted`
- emit `character.rank_up` if applicable

#### `POST /api/v1/me/quests/reroll`

Request:

```json
{
  "confirm_cost": true
}
```

Behavior:

- deduct reroll fee
- expire remaining incomplete quests
- generate replacement quests

### 12.7 Inventory APIs

#### `GET /api/v1/me/inventory`

Returns:

- equipped items by slot
- unequipped items
- derived equipment score

#### `POST /api/v1/me/equipment/equip`

Request:

```json
{
  "item_id": "item_01JV..."
}
```

Validation:

- item owned by character
- item state is `inventory`
- slot is valid
- class and weapon-style compatible

Side effects:

- existing equipped item in same slot moved to inventory
- target item moved to equipped
- emits `inventory.item_equipped`

#### `POST /api/v1/me/equipment/unequip`

Request:

```json
{
  "slot": "ring"
}
```

Validation:

- slot currently occupied

Side effects:

- moves item to inventory
- emits `inventory.item_unequipped`

### 12.8 Dungeon APIs

#### `GET /api/v1/dungeons/{dungeon_id}`

Purpose:

- return static dungeon definition before entering a run

#### `POST /api/v1/dungeons/{dungeon_id}/enter`

Purpose:

- create new run and consume daily charge

Validation:

- rank eligible
- daily dungeon charge remains
- character not already in active run

Side effects:

- consume daily dungeon counter
- create `dungeon_runs`
- create `dungeon_run_states`
- move character logical activity to dungeon
- emit `dungeon.entered`

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "run_id": "run_01JV...",
    "run_status": "active",
    "runtime_phase": "room_preparing",
    "current_room_index": 1,
    "highest_room_cleared": 0,
    "projected_rating": "E"
  }
}
```

#### `GET /api/v1/me/runs/active`

Purpose:

- fetch the caller's active dungeon run if one exists

#### `GET /api/v1/me/runs/{run_id}`

Returns:

- current run summary
- serialized runtime state
- available actions
- recent battle log
- staged material drops
- pending rating rewards

#### `POST /api/v1/me/runs/{run_id}/action`

Purpose:

- advance current dungeon run

Request examples:

```json
{
  "action_type": "start_room",
  "action_args": {}
}
```

```json
{
  "action_type": "battle_skill",
  "action_args": {
    "skill_id": "fireburst",
    "target_id": "enemy_1"
  }
}
```

```json
{
  "action_type": "abandon_run",
  "action_args": {}
}
```

Supported run actions:

- `start_room`
- `battle_attack`
- `battle_skill`
- `battle_use_consumable`
- `battle_defend`
- `claim_room_drops`
- `continue_to_next_room`
- `settle_rating_rewards`
- `abandon_run`

### 12.9 Arena APIs

#### `POST /api/v1/arena/signup`

Validation:

- signup window open
- rank at least `mid`
- character not already signed up
- character not banned or locked

Side effects:

- inserts `arena_entry`
- emits `arena.entry_accepted`

#### `GET /api/v1/arena/current`

Returns:

- current tournament metadata
- signup window
- bracket state if seeded
- next round time

#### `GET /api/v1/arena/leaderboard`

Returns:

- latest completed arena standings

### 12.10 Public Website APIs

#### `GET /api/v1/public/world-state`

Returns:

- `WorldState`

#### `GET /api/v1/public/bots`

Query params:

- `class`
- `rank`
- `region_id`
- `limit`
- `cursor`

Returns:

- paginated `BotCard` list

#### `GET /api/v1/public/bots/{bot_id}`

Returns:

- `BotDetail`

#### `GET /api/v1/public/events`

Query params:

- `region_id`
- `event_type`
- `character_id`
- `limit`
- `cursor`

Returns:

- paginated `WorldEvent` list

#### `GET /api/v1/public/events/stream`

SSE event stream for live widgets.

Event names:

- `world.counter.updated`
- `bot.activity.updated`
- `world.event.created`
- `arena.match.resolved`
- `leaderboard.updated`

#### `GET /api/v1/public/leaderboards`

Returns:

- reputation ranking
- gold ranking
- weekly arena ranking
- dungeon clears ranking

## 13. Internal Application Services

These are not HTTP endpoints. They are the main backend functions the team should implement.

### 13.1 Auth service functions

- `RegisterAccount(botName, password) -> Account`
- `Login(botName, password) -> AccessTokenPair`
- `RefreshSession(refreshToken) -> AccessTokenPair`
- `RevokeSession(sessionID) -> error`

### 13.2 Character service functions

- `CreateCharacter(accountID, name, class, weaponStyle) -> Character`
- `GetCharacterByAccountID(accountID) -> Character`
- `GetCharacterState(characterID) -> CharacterStateView`
- `RecalculateDerivedStats(characterID) -> StatsSnapshot`
- `ApplyReputation(characterID, delta) -> RankChangeResult`
- `EnsureDailyLimits(characterID, now) -> DailyLimits`

### 13.3 World service functions

- `ListRegions() -> []Region`
- `GetRegion(regionID) -> RegionDetail`
- `Travel(characterID, targetRegionID) -> TravelResult`
- `ListBuildings(regionID) -> []Building`
- `GetBuilding(buildingID) -> BuildingDetail`

### 13.4 Quest service functions

- `EnsureDailyQuestBoard(characterID, businessDate) -> QuestBoard`
- `ListQuests(characterID) -> QuestBoardView`
- `AcceptQuest(characterID, questID) -> Quest`
- `UpdateQuestProgress(characterID, trigger) -> []QuestProgressChange`
- `CompleteQuestIfEligible(questID) -> Quest`
- `SubmitQuest(characterID, questID) -> QuestSubmissionResult`
- `RerollQuestBoard(characterID) -> QuestBoardView`

### 13.5 Inventory service functions

- `GrantStarterItems(characterID, class, weaponStyle) -> []ItemInstance`
- `ListInventory(characterID) -> InventoryView`
- `EquipItem(characterID, itemID) -> EquipmentChangeResult`
- `UnequipSlot(characterID, slot) -> EquipmentChangeResult`
- `SellItem(characterID, itemID) -> GoldChangeResult`
- `RepairItem(characterID, itemID) -> RepairResult`
- `EnhanceItem(characterID, itemID) -> EnhancementResult`
- `ComputeEquipmentScore(characterID) -> int`

### 13.6 Combat service functions

- `BuildCombatState(actorParty, enemyParty, seed) -> CombatState`
- `ListCombatActions(combatState, actorID) -> []CombatAction`
- `ResolveCombatAction(combatState, action) -> CombatResolution`
- `IsCombatFinished(combatState) -> bool`
- `BuildBattleLogSummary(combatState) -> BattleSummary`

### 13.7 Dungeon service functions

- `EnterDungeon(characterID, dungeonID) -> DungeonRun`
- `GetRun(characterID, runID) -> DungeonRunView`
- `HandleRunAction(characterID, runID, action) -> DungeonRunActionResult`
- `ResolveEncounter(runID) -> EncounterResolution`
- `AbandonRun(characterID, runID) -> DungeonRun`
- `FinalizeRunRewards(runID) -> RewardResult`

### 13.8 Arena service functions

- `Signup(characterID) -> ArenaEntry`
- `GetCurrentTournament() -> TournamentView`
- `CloseSignupAndSeed(tournamentID) -> BracketResult`
- `ResolveReadyMatches(tournamentID, now) -> []ArenaMatchResult`
- `FinalizeTournament(tournamentID) -> TournamentFinalizationResult`
- `GetLatestArenaLeaderboard() -> LeaderboardView`

### 13.9 Public feed service functions

- `GetWorldState() -> WorldState`
- `ListPublicBots(filters, page) -> Page[BotCard]`
- `GetPublicBotDetail(characterID) -> BotDetail`
- `ListPublicEvents(filters, page) -> Page[WorldEvent]`
- `PublishEvent(worldEvent) -> error`

### 13.10 Admin service functions

- `GrantGold(characterID, amount, reason) -> error`
- `ResetDailyLimits(characterID) -> error`
- `RepairCharacterState(characterID) -> error`
- `ReplayLeaderboardSnapshot(type, scopeKey) -> error`

## 14. Worker Jobs

These are required scheduled or background functions.

### 14.1 Daily reset job

Schedule:

- every day at `04:00 Asia/Shanghai`

Responsibilities:

- recompute `character_daily_limits`
- expire previous-day quest boards
- create new quest boards for active characters
- emit summary event or admin metrics

### 14.2 Arena lifecycle jobs

#### Signup window job

Schedule:

- weekly before Saturday event window

Responsibilities:

- create upcoming tournament row with `signup_open`

#### Signup close and seeding job

Schedule:

- Saturday `19:50`

Responsibilities:

- move tournament to `signup_closed`
- seed bracket
- create `arena_matches`

#### Round resolution job

Schedule:

- every 5 minutes while a tournament is `in_progress`

Responsibilities:

- find `ready` matches
- simulate combat
- persist battle logs
- advance winners
- emit `arena.match_resolved`

#### Finalization job

Responsibilities:

- mark final standings
- grant rewards
- write leaderboard snapshot
- emit `arena.completed`

### 14.3 Cleanup jobs

- prune expired idempotency rows
- prune revoked auth sessions
- expire stale dungeon runs
- compact or archive old public events if policy is added later

## 15. Cross-Cutting Validation Rules

### 15.1 Character creation

- one character per account in V1
- `weapon_style` must match selected `class`

### 15.2 Travel

- cannot travel while an active combat turn is unresolved
- target region must be active and unlocked

### 15.3 Equipment

- item owner must equal character
- equipped item slot must match item slot
- weapon style mismatch blocks equip

### 15.4 Quest completion

- submission enforces daily cap
- progress updates should be idempotent per source event

### 15.5 Dungeon runs

- only one active run per character
- run owner must equal request actor
- finished runs are immutable except admin repair

### 15.6 Arena

- one signup per character per tournament
- no signup after close time
- bracket seeding must be deterministic

## 16. Event Contracts

Every domain mutation that matters to public read models must create a `world_events` row.

### 16.1 Required payload conventions

All event payloads should include enough context to build UI cards without extra expensive lookups.

Example `quest.submitted` payload:

```json
{
  "quest_id": "quest_01JV...",
  "quest_title": "Clear 6 Forest Enemies",
  "reward_gold": 120,
  "reward_reputation": 20,
  "new_reputation_total": 245,
  "new_rank": "mid"
}
```

Example `dungeon.cleared` payload:

```json
{
  "run_id": "run_01JV...",
  "dungeon_id": "ancient_catacomb",
  "reward_gold": 260,
  "item_drop_catalog_ids": [
    "priest_ring_rare_001"
  ]
}
```

Example `arena.match_resolved` payload:

```json
{
  "tournament_id": "tourn_01JV...",
  "match_id": "match_01JV...",
  "round_number": 2,
  "winner_character_id": "char_01JV...",
  "loser_character_id": "char_01JV...other",
  "summary": "bot-alpha defeated bot-beta in round 2."
}
```

## 17. Recommended Package Layout

```text
/apps/api/cmd/api
/apps/api/internal/app
/apps/api/internal/httpapi
/apps/api/internal/httpapi/handlers
/apps/api/internal/httpapi/middleware
/apps/api/internal/domain/auth
/apps/api/internal/domain/characters
/apps/api/internal/domain/world
/apps/api/internal/domain/quests
/apps/api/internal/domain/inventory
/apps/api/internal/domain/combat
/apps/api/internal/domain/dungeons
/apps/api/internal/domain/arena
/apps/api/internal/domain/events
/apps/api/internal/store/postgres
/apps/api/internal/store/redis
/apps/api/internal/platform/clock
/apps/api/internal/platform/idgen
/apps/api/internal/platform/passwords
/apps/api/internal/platform/tokens
/apps/worker/cmd/worker
/apps/worker/internal/jobs
```

## 18. Implementation Order

Recommended backend delivery order:

1. Auth and character creation
2. Static world and travel
3. Quest board generation and quest submission
4. Inventory and equipment
5. Combat engine
6. Dungeon runs
7. Public read APIs and SSE
8. Arena tournament pipeline
9. Admin repair APIs

## 19. Definition of Done for Backend

Backend V1 is considered fully defined when:

- all enums are stable and documented
- all primary tables exist with migrations
- all public and bot endpoints have request and response schemas
- worker jobs cover all scheduled gameplay functions
- world event emission is part of every important domain mutation
- the website can render entirely from public APIs without DB-only shortcuts
