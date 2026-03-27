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
    "available": ["claim_run_rewards"]
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
- `claim_dungeon_rewards`
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

- return static dungeon definition used before entering a run

Returns:

- dungeon metadata
- room count
- recommended level band
- boss room index
- rating rule summary
- visible reward summary

#### `POST /api/v1/dungeons/{dungeon_id}/enter`

Purpose:

- create a run and let the backend battle engine auto-resolve it

Validation:

- rank eligible
- reward-claim daily quota remains (evaluated again at claim time)
- character not already in active run

Side effects:

- create `dungeon_runs`
- create `dungeon_run_states`
- execute run resolution server-side to completion or failure
- stage reward package when run is clear-successful
- emit `dungeon.entered`
- emit `dungeon.cleared` or `dungeon.failed`

Success response:

```json
{
  "request_id": "req_01",
  "data": {
    "run_id": "run_01JV...",
    "run_status": "cleared",
    "runtime_phase": "result_ready",
    "current_room_index": 6,
    "highest_room_cleared": 6,
    "projected_rating": "A",
    "current_rating": "A",
    "reward_claimable": true,
    "available_actions": [
      {
        "action_type": "claim_run_rewards"
      }
    ]
  }
}
```

#### `GET /api/v1/me/runs/active`

Purpose:

- fetch the caller's current active dungeon run if one exists

Returns:

- `null` if no active run exists
- otherwise the same payload shape as `GET /api/v1/me/runs/{run_id}`

#### `GET /api/v1/me/runs/{run_id}`

Returns:

- run summary and auto-resolution output
- serialized runtime state
- claimability fields
- recent battle log (read-only)
- staged material drops
- pending rating rewards

#### `POST /api/v1/me/runs/{run_id}/claim`

Purpose:

- claim staged rewards for a cleared run

Validation:

- run exists and belongs to caller
- run status is `cleared`
- rewards are still claimable and unclaimed
- daily reward-claim quota remains

Side effects:

- grant staged rewards (gold/items/materials)
- consume one daily dungeon reward claim counter
- mark run rewards as claimed and immutable
- emit `dungeon.loot_granted`

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
