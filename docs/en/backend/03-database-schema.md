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
- `template_id` `text` null references `quest_templates(id)`
- `template_type` `text` not null
- `difficulty` `text` not null
- `flow_kind` `text` not null
- `rarity` `text` not null
- `status` `text` not null
- `title` `text` not null
- `description` `text` not null
- `target_region_id` `text` null references `regions(id)`
- `target_dungeon_id` `text` null
- `target_enemy_key` `text` null
- `current_step_key` `text` null
- `progress_current` `int` not null default `0`
- `progress_target` `int` not null
- `state_json` `jsonb` not null
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
- index on `template_id`
- index on `(character_id, current_step_key)`

Notes:

- `state_json` stores discovered clues, completed steps, branch choices, and runtime variables
- `progress_current/progress_target` remain useful for simple counter quests
- multi-step quests should not rely on a single numeric progress pair to represent the whole runtime

### 9.13 `quest_templates`

Purpose:

- reusable quest blueprints used to generate daily or special quest instances
- persistent mirror of the quest-template catalog when templates are promoted into relational storage

Current source-of-truth note:

- YAML files under `apps/api/internal/quests/configs/` should be treated as the primary authoring surface for quest templates
- `quest_templates` rows are a future persistence/admin surface, not the required day-one editing path

Fields:

- `id` `text` primary key
- `template_type` `text` not null
- `difficulty` `text` not null
- `flow_kind` `text` not null
- `rarity` `text` not null
- `title` `text` not null
- `description` `text` not null
- `targeting_rules_json` `jsonb` not null
- `step_graph_json` `jsonb` not null
- `generation_rules_json` `jsonb` not null
- `reward_profile_id` `text` null
- `narrative_json` `jsonb` not null
- `is_daily_pool` `boolean` not null default `true`
- `is_active` `boolean` not null default `true`

Indexes:

- index on `(difficulty, is_daily_pool, is_active)`
- index on `(flow_kind, is_active)`

Notes:

- template authoring should prefer YAML-first changes and service-side validation at startup
- one template file should carry quest identity, targeting, runtime steps, interaction specs, choice specs, progression triggers, and optional rank overrides together
- `step_graph_json` and `generation_rules_json` should remain structurally compatible with the YAML representation so the catalog can later move between file-backed and database-backed sources without redefining the model

### 9.14 `quest_progress_events`

Purpose:

- idempotent trigger log for quest progression

Fields:

- `id` `text` primary key
- `character_id` `text` not null references `characters(id)`
- `quest_id` `text` not null references `quests(id)`
- `trigger_type` `text` not null
- `source_event_id` `text` not null
- `payload_json` `jsonb` not null
- `applied_at` `timestamptz` not null

Indexes:

- unique index on `(quest_id, source_event_id)`
- index on `(character_id, trigger_type)`

### 9.15 `dungeon_definitions`

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

### 9.16 `dungeon_runs`

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

### 9.17 `dungeon_run_states`

Purpose:

- current serialized state of a run

Fields:

- `run_id` `text` primary key references `dungeon_runs(id)`
- `state_version` `int` not null default `1`
- `state_json` `jsonb` not null
- `updated_at` `timestamptz` not null

Notes:

- stores room state, combat snapshot, staged kill drops, pending rating rewards, and available action context

### 9.18 `arena_tournaments`

Purpose:

- one daily qualifier cycle

Fields:

- `id` `text` primary key
- `day_key` `text` unique not null
- `status` `text` not null
- `signup_opens_at` `timestamptz` not null
- `signup_closes_at` `timestamptz` not null
- `starts_at` `timestamptz` not null
- `completed_at` `timestamptz` null
- `bracket_size` `int` not null
- `snapshot_json` `jsonb` null

### 9.19 `arena_entries`

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

### 9.20 `arena_matches`

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

### 9.21 `leaderboard_snapshots`

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

### 9.22 `world_events`

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

### 9.23 `idempotency_keys`

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
