# Dungeon Data Tables And Template Spec

## 1. Goal

This document converts the dungeon, monster, difficulty, combat, and loot specs into table-oriented structures that can be implemented directly.

Goals:

- define clean dungeon system tables
- keep monsters and wave layouts data-driven
- make an escalating run of up to `6` rooms fully data-driven
- control rating-based gear rewards and monster-kill material drops from tables rather than code branches

## 2. Recommended Layers

The dungeon system should be split into:

1. dungeon definition layer
2. room config layer
3. monster template layer
4. room spawn layer
5. boss phase and script layer
6. reward layer

## 3. Core Table List

Recommended tables:

- `dungeon_definitions`
- `dungeon_room_configs`
- `dungeon_monster_templates`
- `dungeon_monster_skills`
- `dungeon_rooms`
- `dungeon_room_monsters`
- `dungeon_boss_phases`
- `dungeon_boss_phase_actions`
- `dungeon_rating_reward_tables`
- `dungeon_rating_reward_entries`
- `monster_material_drop_tables`
- `monster_material_drop_entries`

## 4. `dungeon_definitions`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | dungeon ID |
| `region_id` | text | attached region |
| `name` | text | public name |
| `public_name_zh` | text | Chinese display name |
| `tier` | text | `T1-T5` |
| `recommended_level_min` | int | recommended minimum |
| `recommended_level_max` | int | recommended maximum |
| `entry_rank_requirement` | text | rank gate |
| `party_size_min` | int | minimum party size |
| `party_size_max` | int | maximum party size, currently `4` |
| `room_count` | int | current max is `6` |
| `round_limit` | int | per-room round cap, currently `10` |
| `boss_monster_id` | text | main boss template |
| `is_active` | bool | enabled flag |

## 5. `dungeon_room_configs`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | room config ID |
| `dungeon_id` | text | owning dungeon |
| `room_index` | int | room number, `1-6` |
| `hp_multiplier` | numeric | monster HP multiplier |
| `damage_multiplier` | numeric | damage multiplier |
| `heal_shield_multiplier` | numeric | healing/shield multiplier |
| `ai_complexity` | text | `basic` / `medium` / `high` |
| `extra_mechanic_flag` | bool | extra mechanic enabled |
| `rating_if_failed_here` | text | default rating if the run ends here |
| `timeout_consumes_entry` | bool | currently `false` |

## 6. `dungeon_monster_templates`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | monster ID |
| `dungeon_id` | text | owning dungeon |
| `name` | text | monster name |
| `public_name_zh` | text | Chinese display |
| `monster_role` | text | bruiser/tank/caster/etc |
| `monster_class` | text | normal / elite / boss / summon |
| `base_hp` | int | base HP |
| `base_physical_attack` | int | base physical attack |
| `base_magic_attack` | int | base magic attack |
| `base_physical_defense` | int | base physical defense |
| `base_magic_defense` | int | base magic defense |
| `base_speed` | int | base speed |
| `base_healing_power` | int | base healing power |
| `ai_profile` | text | AI profile |
| `flavor_text` | text | short flavor |
| `is_boss` | bool | boss flag |

## 7. `dungeon_monster_skills`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | skill ID |
| `monster_id` | text | owner monster |
| `name` | text | skill name |
| `skill_tier` | text | normal / advanced / ultimate |
| `action_type` | text | attack/buff/debuff/heal/etc |
| `target_type` | text | single/all/random/etc |
| `damage_type` | text | physical/magic/poison/etc |
| `cooldown_rounds` | int | cooldown |
| `skill_power` | int | base power |
| `atk_ratio` | numeric | attack ratio |
| `def_ratio` | numeric | defense ratio |
| `heal_ratio` | numeric | heal ratio |
| `shield_ratio` | numeric | shield ratio |
| `status_payload_json` | jsonb | statuses |
| `priority_weight` | int | AI weight |
| `condition_json` | jsonb | trigger conditions |

## 8. `dungeon_rooms`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | room ID |
| `dungeon_id` | text | owning dungeon |
| `room_index` | int | room number |
| `room_type` | text | normal / elite / boss / event |
| `spawn_rule` | text | spawn mode |
| `rating_on_clear` | text | base rating after clearing the room |
| `notes` | text | encounter note |

## 9. `dungeon_room_monsters`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | row ID |
| `room_id` | text | owning room |
| `monster_id` | text | monster template |
| `slot_index` | int | slot |
| `count` | int | quantity |
| `is_elite` | bool | elite flag |
| `spawn_phase` | text | opening / summoned |

## 10. `dungeon_boss_phases`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | phase ID |
| `dungeon_id` | text | owning dungeon |
| `boss_monster_id` | text | boss |
| `room_index` | int | owning room |
| `phase_index` | int | phase number |
| `trigger_type` | text | hp / round / summon trigger |
| `trigger_value` | numeric | trigger value |
| `phase_name` | text | phase label |
| `phase_summary` | text | phase summary |

## 11. `dungeon_boss_phase_actions`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | action row |
| `phase_id` | text | owning phase |
| `round_offset` | int | rounds since phase start |
| `action_kind` | text | cast/summon/buff/field effect |
| `skill_id` | text | linked skill |
| `target_rule` | text | target rule |
| `payload_json` | jsonb | extra parameters |

## 12. `dungeon_rating_reward_tables`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | rating reward table ID |
| `dungeon_id` | text | owning dungeon |
| `rating` | text | `S` / `A` / `B` / `C` / `D` / `E` |
| `guaranteed_drop_count` | int | guaranteed gear roll count |
| `bonus_drop_chance` | numeric | extra gear roll chance |
| `high_quality_bonus_count` | int | extra high-quality roll count |
| `rarity_profile` | text | rarity profile name |

## 13. `dungeon_rating_reward_entries`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | row ID |
| `reward_table_id` | text | owning rating reward table |
| `item_template_id` | text | item template |
| `slot` | text | gear slot |
| `rarity` | text | rarity |
| `weight` | int | roll weight |
| `min_roll` | int | min count |
| `max_roll` | int | max count |
| `class_bias` | text | class bias rule |

## 14. `monster_material_drop_tables`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | table ID |
| `monster_id` | text | owning monster template |
| `room_band_min` | int | minimum applicable room |
| `room_band_max` | int | maximum applicable room |
| `drop_scope` | text | `on_kill` |
| `guaranteed_material_count` | int | guaranteed count on kill |

## 15. `monster_material_drop_entries`

Suggested fields:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | text | row ID |
| `material_table_id` | text | owning material table |
| `material_id` | text | material template |
| `weight` | int | roll weight |
| `min_count` | int | min count |
| `max_count` | int | max count |
| `is_boss_exclusive` | bool | boss-exclusive flag |

## 16. Ancient Catacomb Example

### 16.1 `dungeon_definitions`

| Field | Example |
| --- | --- |
| `id` | `dng_ancient_catacomb` |
| `region_id` | `ancient_catacomb` |
| `name` | `Ancient Catacomb` |
| `public_name_zh` | `远古墓窟` |
| `tier` | `T1` |
| `recommended_level_min` | `1` |
| `recommended_level_max` | `20` |
| `entry_rank_requirement` | `low` |
| `party_size_min` | `1` |
| `party_size_max` | `4` |
| `round_limit` | `10` |
| `boss_monster_id` | `boss_morthis` |

### 16.2 `dungeon_room_configs`

| room_index | hp_multiplier | damage_multiplier | extra_mechanic_flag | rating_if_failed_here |
| --- | --- | --- | --- | --- |
| `1` | `1.00` | `1.00` | `false` | `E` |
| `2` | `1.10` | `1.05` | `false` | `D` |
| `3` | `1.22` | `1.12` | `false` | `C` |
| `4` | `1.36` | `1.20` | `true` | `B` |
| `5` | `1.52` | `1.30` | `true` | `A` |
| `6` | `1.72` | `1.42` | `true` | `S` |

### 16.3 `dungeon_rooms`

| room_index | room_type | rating_on_clear | notes |
| --- | --- | --- | --- |
| `1` | `normal` | `E` | 2 normal monsters |
| `2` | `normal` | `D` | 3 normal monsters |
| `3` | `elite` | `C` | 1 normal monster + 1 elite |
| `4` | `normal` | `B` | 2 normal monsters + 1 controller |
| `5` | `elite` | `A` | 2 elites |
| `6` | `boss` | `S` | two-phase boss |

## 17. Naming Rules

Recommended ID prefixes:

- dungeon: `dng_*`
- room config: `drc_*`
- monster: `mon_*`
- boss: `boss_*`
- room: `room_*`
- phase: `phase_*`
- rating reward table: `reward_*`
- material table: `mat_*`

## 18. Suggested Implementation Order

1. `dungeon_definitions` + `dungeon_room_configs`
2. `dungeon_monster_templates` + `dungeon_monster_skills`
3. `dungeon_rooms` + `dungeon_room_monsters`
4. `dungeon_boss_phases` + `dungeon_boss_phase_actions`
5. `dungeon_rating_reward_tables` + `monster_material_drop_tables`

## 19. Next Step

The cleanest follow-up document is:

- `14-first-batch-dungeon-balance-sheets.md`

That file can contain concrete monster stats, skills, room configs, rating rewards, and kill-drop weights for Ancient Catacomb and Thorned Hollow.
