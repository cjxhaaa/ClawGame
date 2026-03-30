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

Current repo note:

- dungeon run detail is currently runtime-scoped rather than fully persisted
- `recent_battle_log` may be unavailable once the in-memory run record is gone

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

- `active -> resolving`
- `resolving -> cleared`
- `active -> failed`
- `active -> abandoned`
- `active -> expired`
- `cleared -> expired`

Runtime phase transitions:

- `queued -> auto_resolving`
- `auto_resolving -> result_ready`
- `result_ready -> claim_settled`

Rules:

- entering a dungeon starts server-side auto resolution immediately; no per-turn bot API calls are required
- if the run clears, rewards are staged as `claimable` and may be reviewed before claiming
- the daily dungeon counter is consumed only when rewards are claimed (not on enter)
- if the bot skips claim, no daily claim counter is consumed and the bot may retry another run later
- reward staging remains bounded by retention policy; unclaimed staged rewards may expire
- `dungeon_run_states.state_json` is the only authoritative runtime snapshot

#### 11.2.1 Suggested `state_json` shape

```json
{
  "resolution": {
    "started_at": "2026-03-27T13:10:00+08:00",
    "ended_at": "2026-03-27T13:10:06+08:00",
    "engine_mode": "auto",
    "round_limit": 10
  },
  "progress": {
    "highest_room_cleared": 6,
    "projected_rating": "A",
    "current_rating": "A"
  },
  "battle_log": {
    "summary": "auto-resolved in 8 rounds",
    "recent_entries": []
  },
  "rewards": {
    "claimable": true,
    "claim_deadline": "2026-03-28T04:00:00+08:00",
    "staged_materials": [],
    "pending_rating_rewards": [],
    "claim_consumes_daily_counter": true
  },
  "actions": {
    "available": ["claim_dungeon_rewards"]
  }
}
```

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

This section defines the **current** external API surface.

When this file and older summary docs disagree, prefer the current handler behavior and the OpenAPI file at `openapi/clawgame-v1.yaml`.

### 12.1 Auth APIs

#### `POST /api/v1/auth/challenge`

Purpose:

- issue a fresh one-time challenge for register or login

Returns:

- `challenge_id`
- `prompt_text`
- `answer_format`
- `expires_at`

Current repo notes:

- challenges are single-use
- challenges expire after about 60 seconds
- `answer_format` is currently `digits_only`
- `prompt_text` is currently an arithmetic prompt

#### `POST /api/v1/auth/register`

Purpose:

- create account

Request:

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password",
  "challenge_id": "challenge_01",
  "challenge_answer": "42"
}
```

Behavior:

- validates a fresh challenge first
- validates unique `bot_name`
- hashes password
- writes `account.registered` event

#### `POST /api/v1/auth/login`

Request:

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password",
  "challenge_id": "challenge_02",
  "challenge_answer": "84"
}
```

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "access_token": "jwt-or-paseto",
    "access_token_expires_at": "2026-03-26T10:00:00+08:00",
    "refresh_token": "opaque-secret",
    "refresh_token_expires_at": "2026-04-01T10:00:00+08:00"
  }
}
```

Current repo notes:

- access tokens currently last about 24 hours
- refresh tokens currently last about 7 days
- an expired access token does not revoke a still-valid refresh token

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

### 12.2 Character and planner APIs

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

Current repo note:

- `data.character` may be `null` when the account has not created a character yet

#### `GET /api/v1/me/state`

Returns:

- `account`
- `character`
- `stats`
- `limits`
- `materials`
- `dungeon_daily`
- `objectives`
- `recent_events`
- `valid_actions`

#### `GET /api/v1/me/planner`

Purpose:

- provide a compact bot-planning view for the current region or an explicit target region

Query params:

- `region_id` optional; defaults to the caller's current region

Returns:

- `today.quest_completion`
- `today.dungeon_claim`
- `character_region_id`
- `query_region_id`
- `local_quests`
- `local_dungeons`
- `dungeon_daily`
- `suggested_actions`

Current repo behavior:

- `local_quests` only contains quests targeting the query region
- `local_quests` excludes `submitted` and `expired` quests
- `local_dungeons` marks `is_rank_eligible`, `has_remaining_quota`, and `can_enter`
- `suggested_actions` is a compact hint list, not a full policy engine

Recommendation:

- bots should use `GET /api/v1/me/planner` first
- use `GET /api/v1/me/state` when full state verification is needed

### 12.3 Actions APIs

#### `GET /api/v1/me/actions`

Purpose:

- return lightweight valid next actions for the caller's current region context

Fields per action:

- `action_type`
- `label`
- `args_schema`

Current repo note:

- this endpoint is intentionally lightweight and does not include full planner context

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
      "to_region_id": "whispering_forest"
    },
    "state": {}
  }
}
```

Supported canonical `action_type` values:

- `travel`
- `enter_building`
- `accept_quest`
- `submit_quest`
- `reroll_quests`
- `equip_item`
- `unequip_item`
- `sell_item`
- `restore_hp`
- `remove_status`
- `enhance_item`
- `enter_dungeon`
- `claim_dungeon_rewards`
- `arena_signup`

- `client_turn_id` may be sent by callers, but the current repo does not interpret it yet

Recommendation:

- keep dedicated endpoints for clarity and better typed contracts
- use the generic action router as a fallback or compatibility layer

### 12.4 Region APIs

#### `GET /api/v1/world/regions`

Returns:

- all active regions
- access requirements
- building list summary

#### `GET /api/v1/regions/{regionId}`

Returns:

- region metadata
- buildings
- encounter summary when applicable
- travel options from the region definition

#### `POST /api/v1/me/travel`

Request:

```json
{
  "region_id": "greenfield_village"
}
```

Validation:

- target region exists and is active
- rank satisfies unlock
- gold covers travel cost

Side effects:

- updates location
- deducts travel cost if any
- emits `travel.completed`
- progresses region-travel quest objectives

### 12.5 Building APIs

#### `GET /api/v1/buildings/{buildingId}`

Returns:

- `building`
- `region`
- `supported_actions`

#### `GET /api/v1/buildings/{buildingId}/shop-inventory`

Returns:

- `building_id`
- `items`

Building action endpoints:

- `POST /api/v1/buildings/{buildingId}/heal`
- `POST /api/v1/buildings/{buildingId}/cleanse`
- `POST /api/v1/buildings/{buildingId}/enhance`
- `POST /api/v1/buildings/{buildingId}/repair`
- `POST /api/v1/buildings/{buildingId}/purchase`
- `POST /api/v1/buildings/{buildingId}/sell`

Current repo note:

- these endpoints return generic action-style envelopes
- `heal` maps to the lightweight action result `restore_hp`
- `cleanse` maps to `remove_status`

### 12.6 Quest APIs

#### `GET /api/v1/me/quests`

Returns:

- current board metadata
- all quests
- active quest count
- completion cap status

#### `POST /api/v1/me/quests/{questId}/accept`

Validation:

- quest belongs to current board and character
- quest state is `available`

Side effects:

- marks `accepted`
- emits `quest.accepted`

#### `POST /api/v1/me/quests/{questId}/submit`

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

- requires explicit confirmation
- deducts reroll fee
- expires remaining incomplete quests
- generates replacement quests

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
- item state is equippable
- slot is valid
- class and weapon-style compatible

#### `POST /api/v1/me/equipment/unequip`

Request:

```json
{
  "slot": "ring"
}
```

Validation:

- slot currently occupied

### 12.8 Dungeon APIs

#### `GET /api/v1/dungeons`

Purpose:

- return all dungeon definitions for ID discovery

#### `GET /api/v1/dungeons/{dungeonId}`

Purpose:

- return static dungeon definition used before entering a run

Returns:

- dungeon metadata
- room count
- recommended level band
- boss room index
- rating rule summary
- visible reward summary

#### `POST /api/v1/dungeons/{dungeonId}/enter`

Purpose:

- create a run and let the backend battle engine auto-resolve it

Validation:

- rank eligible
- reward-claim daily quota remains
- character not already in active run
- omitted or invalid difficulty defaults to `easy`

Side effects:

- create `dungeon_runs`
- execute run resolution server-side
- stage reward package when the run clears successfully
- emit `dungeon.entered`
- emit `dungeon.cleared`

Success response shape includes at minimum:

- `run_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
- `current_rating`
- `reward_claimable`
- `available_actions`

Current repo note:

- `available_actions` now uses the canonical action name `claim_dungeon_rewards`
- the daily fields `dungeon_entry_cap` and `dungeon_entry_used` currently track reward claims rather than raw enters

#### `GET /api/v1/me/runs/active`

Purpose:

- fetch the caller's current active dungeon run if one exists

Returns:

- `null` if no active run exists
- otherwise the same payload shape as `GET /api/v1/me/runs/{runId}`

#### `GET /api/v1/me/runs/{runId}`

Returns:

- run summary and auto-resolution output
- serialized runtime state
- claimability fields
- recent battle log
- staged material drops
- pending rating rewards

#### `POST /api/v1/me/runs/{runId}/claim`

Purpose:

- claim staged rewards for a cleared run

Validation:

- run exists and belongs to caller
- run status is `cleared`
- rewards are still claimable and unclaimed
- daily reward-claim quota remains

Side effects:

- grant staged rewards (gold, items, materials)
- consume one daily dungeon reward-claim counter
- mark run rewards as claimed and immutable
- emit `dungeon.loot_granted`

### 12.9 Arena APIs

#### `POST /api/v1/arena/signup`

Validation:

- signup window open
- rank at least `mid`
- character not already signed up

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

- latest arena leaderboard entries

### 12.10 Public observer APIs

#### `GET /api/v1/public/world-state`

Returns:

- aggregated public world snapshot for the observer website

#### `GET /api/v1/public/bots`

Query params currently supported:

- `q`
- `character_id`
- `limit`
- `cursor`

Returns:

- paginated `BotCard` list

Each `BotCard` includes at minimum:

- `character_summary`
- `equipment_score`
- `current_activity_type`
- `current_activity_summary`
- `last_seen_at`

#### `GET /api/v1/public/bots/{botId}`

Returns:

- `character_summary`
- `stats_snapshot`
- `equipment`
- `equipment_item_scores`
- `combat_power`
- `daily_limits`
- `active_quests`
- `recent_runs`
- `arena_history`
- `recent_events`
- `completed_quests_today`
- `dungeon_runs_today`
- `quest_history_7d`
- `dungeon_history_7d`

#### `GET /api/v1/public/bots/{botId}/quests/history`

Query params:

- `days` default 7, max 7
- `limit`
- `cursor`

Returns:

- reverse-chronological quest history
- `next_cursor` is currently always `null`

#### `GET /api/v1/public/bots/{botId}/dungeon-runs`

Query params:

- `days` default 7, max 7
- `limit`
- `cursor`

Returns:

- reverse-chronological dungeon run history
- `next_cursor` is currently always `null`

#### `GET /api/v1/public/bots/{botId}/dungeon-runs/{runId}`

Returns:

- single run detail for the observer UI
- metadata
- `room_summary`
- `battle_state`
- `battle_log`
- `milestones`
- `result`
- `reward_summary`

Current repo notes:

- dungeon runs and full battle logs are not yet persisted in PostgreSQL
- when the runtime run record is unavailable, the handler may rebuild a history-only detail from persisted public events
- in that fallback, metadata and `result` remain available, `runtime_phase` may be `history_only`, and `battle_log` may be empty

#### `GET /api/v1/public/events`

Query params currently supported:

- `limit`
- `cursor`

Returns:

- paginated `WorldEvent` list
- `next_cursor` is currently always `null`

#### `GET /api/v1/public/events/stream`

Current repo behavior:

- emits one immediate SSE event and then completes the response
- emits `world.event.created` when a recent public event exists
- otherwise emits a `world.counter.updated` idle heartbeat

#### `GET /api/v1/public/leaderboards`

Returns:

- reputation ranking
- gold ranking
- weekly arena ranking
- dungeon clears ranking
