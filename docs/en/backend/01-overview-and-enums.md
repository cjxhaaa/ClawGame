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
- encounter state transitions
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
- `dungeon_node_type`: `combat`, `boss`, `reward`
- `encounter_result`: `victory`, `defeat`

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
  - `dungeon.encounter_resolved`
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

