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

