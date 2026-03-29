# Dungeon Data Tables And Template Spec

## 1. Goal

This document defines the data structures for the multi-enemy dungeon system, ensuring the following capabilities are configurable:

- multi-enemy compositions per room
- normal / elite / boss monster tiering
- easy / hard / nightmare three-tier difficulty
- boss only in the final room

## 2. Data Layers

Split into seven layers:

1. dungeon definition layer
2. difficulty profile layer
3. room config layer
4. monster template layer
5. room wave layer
6. boss phase script layer
7. reward and drop layer

## 3. Core Table List

- `dungeon_definitions`
- `dungeon_difficulty_profiles`
- `dungeon_room_configs`
- `dungeon_monster_templates`
- `dungeon_monster_skills`
- `dungeon_room_waves`
- `dungeon_wave_slots`
- `dungeon_boss_phases`
- `dungeon_boss_phase_actions`
- `dungeon_rating_reward_tables`
- `monster_material_drop_tables`

## 4. dungeon_definitions

Defines the dungeon itself.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | dungeon ID |
| `name` | text | display name |
| `region_id` | text | attached region |
| `room_count` | int | fixed `6` |
| `boss_room_index` | int | fixed `6` |
| `min_rank` | text | minimum character rank |
| `recommended_level_min` | int | recommended minimum level |
| `recommended_level_max` | int | recommended maximum level |
| `is_novice` | bool | novice dungeon flag |

## 5. dungeon_difficulty_profiles

Defines global multipliers for each of the three difficulty tiers.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | profile ID |
| `dungeon_id` | text | owning dungeon |
| `difficulty` | text | `easy` / `hard` / `nightmare` |
| `hp_multiplier` | numeric | global HP multiplier |
| `damage_multiplier` | numeric | global damage multiplier |
| `defense_multiplier` | numeric | global defense multiplier |
| `speed_multiplier` | numeric | global speed multiplier |
| `elite_spawn_bias` | numeric | elite appearance probability bias |
| `mechanic_intensity` | text | `low` / `mid` / `high` |

## 6. dungeon_room_configs

Defines the pressure target for each room at each difficulty.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | room config ID |
| `dungeon_id` | text | owning dungeon |
| `difficulty` | text | difficulty tier |
| `room_index` | int | room number `1-6` |
| `room_type` | text | `normal` / `elite` / `boss` |
| `target_monster_count_min` | int | minimum monster count |
| `target_monster_count_max` | int | maximum monster count |
| `must_have_elite` | bool | whether elite is required |
| `allow_boss` | bool | only `true` when `room_index = 6` |
| `rating_if_failed_here` | text | rating if run ends here |

Constraint:

- `allow_boss = true` is only valid when `room_index = 6`

## 7. dungeon_monster_templates

Defines monster templates.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | monster ID |
| `dungeon_id` | text | owning dungeon |
| `name` | text | name |
| `monster_tier` | text | `normal` / `elite` / `boss` |
| `monster_role` | text | `bruiser` / `caster` / `tank` / `controller` / `assassin` / `summoner` |
| `base_hp` | int | base HP |
| `base_physical_attack` | int | base physical attack |
| `base_magic_attack` | int | base magic attack |
| `base_physical_defense` | int | base physical defense |
| `base_magic_defense` | int | base magic defense |
| `base_speed` | int | base speed |
| `base_healing_power` | int | base healing power |
| `ai_profile` | text | AI profile ID |

## 8. dungeon_monster_skills

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | skill ID |
| `monster_id` | text | owning monster |
| `name` | text | skill name |
| `skill_tier` | text | `normal` / `advanced` / `ultimate` |
| `action_type` | text | attack / buff / debuff / heal / summon |
| `target_type` | text | target rule |
| `damage_type` | text | physical / magic / poison / etc |
| `cooldown_rounds` | int | cooldown |
| `skill_power` | int | base power |
| `atk_ratio` | numeric | attack ratio |
| `status_payload_json` | jsonb | applied statuses |
| `priority_weight` | int | AI priority weight |
| `condition_json` | jsonb | trigger conditions |

## 9. dungeon_room_waves

Defines the waves for each room. V1 starts with single-wave rooms; can expand later.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | wave ID |
| `dungeon_id` | text | owning dungeon |
| `difficulty` | text | difficulty tier |
| `room_index` | int | room number |
| `wave_index` | int | wave number |
| `is_boss_wave` | bool | whether this is a boss wave |

## 10. dungeon_wave_slots

Defines the specific monster slots within a wave.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | slot ID |
| `wave_id` | text | owning wave |
| `slot_index` | int | slot position |
| `monster_id` | text | monster template ID |
| `count` | int | quantity |
| `spawn_timing` | text | `start` / `round_2` / `phase_2` |

Constraints:

- when `room_index < 6`, `monster_tier = boss` is forbidden
- when `room_index = 6`, at least 1 slot must have `monster_tier = boss`

## 11. dungeon_boss_phases

Defines boss phases.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | phase ID |
| `dungeon_id` | text | owning dungeon |
| `difficulty` | text | difficulty tier |
| `boss_monster_id` | text | boss monster ID |
| `phase_index` | int | phase number |
| `trigger_type` | text | `hp_threshold` / `round_threshold` |
| `trigger_value` | numeric | threshold value |
| `phase_name` | text | phase label |

## 12. dungeon_boss_phase_actions

Defines phase action scripts.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | action ID |
| `phase_id` | text | owning phase |
| `round_offset` | int | rounds after phase start |
| `action_kind` | text | `cast_skill` / `summon` / `field_effect` |
| `skill_id` | text | linked skill ID |
| `target_rule` | text | target selection rule |

## 13. dungeon_rating_reward_tables

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | reward table ID |
| `dungeon_id` | text | owning dungeon |
| `difficulty` | text | difficulty tier |
| `rating` | text | `S` / `A` / `B` / `C` / `D` / `E` |
| `guaranteed_drop_count` | int | guaranteed gear roll count |
| `bonus_drop_chance` | numeric | extra roll chance |
| `rarity_profile` | text | rarity profile name |

## 14. monster_material_drop_tables

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | table ID |
| `monster_id` | text | owning monster |
| `difficulty` | text | difficulty tier |
| `drop_scope` | text | `on_kill` |
| `guaranteed_material_count` | int | guaranteed count per kill |

## 15. Rewards And Drops

- equipment rewards are resolved by final clear rating
- materials drop from monster kills
- `nightmare` may increase weights for high-tier materials but does not change the boss-only-final-room structure rule

## 16. Enforced Constraints (must be validated in code)

1. every dungeon must have three difficulty profiles: `easy`, `hard`, and `nightmare`
2. every difficulty must have a complete room config for `room_index = 1..6`
3. `room_index = 6` must contain a boss
4. `room_index < 6` must not contain a boss
5. `nightmare` room monster counts, elite density, or mechanic density must be higher than `hard`

## 17. Relation To Balance Sheet Document

Specific monster compositions and three-tier per-room tables are in:

- `14-first-batch-dungeon-balance-sheets.md`

## 18. ID Prefix Conventions

- dungeon: `dng_*`
- difficulty profile: `ddp_*`
- room config: `drc_*`
- monster template: `mon_*`
- boss: `boss_*`
- wave: `wave_*`
- wave slot: `slot_*`
- phase: `phase_*`
- reward table: `reward_*`
- material table: `mat_*`
