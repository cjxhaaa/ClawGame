# Battle Consumable And Potion System

## 1. Purpose

This document defines a potion-driven battle sustain model for ClawGame.

It addresses the current issue where battles become too easy if HP is fully restored between dungeon rooms.

Goals:

- remove automatic full HP restoration between dungeon rooms
- make dungeon progression depend on resource planning
- introduce rank-gated potion progression sold through shops
- provide four first-class potion families:
  - HP sustain
  - attack boost
  - defense boost
  - speed boost
- keep rules deterministic enough for bot planning

## 2. Core Rule Change

### 2.1 No Full HP Refill Between Rooms

For dungeon room transitions in V1.1:

- party members keep their current HP from the previous room
- no automatic full HP refill on entering the next room
- no hidden free refill in battle initialization for dungeon room 2+

Optional softening rule for onboarding dungeons only:

- novice dungeons may grant a small transition recovery of `5% max_hp` per room
- this must be explicit in dungeon metadata, never implicit

### 2.2 Scope

This rule applies to:

- `dungeon_wave`
- `boss_phase` inside dungeons

This rule does not change field encounters unless explicitly configured later.

## 3. Potion Families

V1.1 potion families:

1. HP Potion
   - instant healing
2. Attack Potion
   - temporary attack increase
3. Defense Potion
   - temporary defense increase
4. Speed Potion
   - temporary speed increase

All non-HP potions are temporary combat buffs, not permanent character growth.

## 4. Rank-Gated Shop Access

Potion purchase access is controlled by character rank.

| Character Rank | Shop Tier | Allowed Potion Tier |
| --- | --- | --- |
| low | novice | T1 |
| mid | advanced | T1-T2 |
| high | elite | T1-T3 |

Principles:

- higher rank unlocks stronger potion tiers
- lower-tier potions remain purchasable after rank-up
- potion power should scale by tier, not by random rolls

## 5. Potion Catalog (Initial)

### 5.1 HP Potions

| Potion ID | Tier | Effect | Notes |
| --- | --- | --- | --- |
| potion_hp_t1 | T1 | restore `25% max_hp`, cap `220` | cheap sustain for low rank |
| potion_hp_t2 | T2 | restore `35% max_hp`, cap `520` | mid-rank dungeon sustain |
| potion_hp_t3 | T3 | restore `45% max_hp`, cap `980` | high-rank sustain |

### 5.2 Attack Potions

| Potion ID | Tier | Effect | Duration |
| --- | --- | --- | --- |
| potion_atk_t1 | T1 | `+10%` primary attack | `3` rounds |
| potion_atk_t2 | T2 | `+16%` primary attack | `3` rounds |
| potion_atk_t3 | T3 | `+24%` primary attack | `4` rounds |

Primary attack means:

- warrior: `physical_attack`
- mage/priest: `magic_attack`

### 5.3 Defense Potions

| Potion ID | Tier | Effect | Duration |
| --- | --- | --- | --- |
| potion_def_t1 | T1 | `+10%` physical and magic defense | `3` rounds |
| potion_def_t2 | T2 | `+16%` physical and magic defense | `3` rounds |
| potion_def_t3 | T3 | `+24%` physical and magic defense | `4` rounds |

### 5.4 Speed Potions

| Potion ID | Tier | Effect | Duration |
| --- | --- | --- | --- |
| potion_spd_t1 | T1 | `+8%` speed | `3` rounds |
| potion_spd_t2 | T2 | `+12%` speed | `3` rounds |
| potion_spd_t3 | T3 | `+18%` speed | `4` rounds |

## 6. Usage Rules

### 6.0 Dungeon Entry Loadout Rule

Before entering a dungeon, OpenClaw may choose up to two potion IDs as the run loadout.

Rules:

- the loadout may contain zero, one, or two potion IDs
- selected potion IDs must be different
- only the selected potion families are available to the dungeon auto-battle engine during that run
- the selected potion IDs must already exist in inventory; dungeon entry does not mint free potions
- the selected potion IDs are snapshotted onto the run so later inventory changes do not rewrite historical run configuration

Design goal:

- make potion planning an explicit pre-run decision
- keep the choice surface small enough for OpenClaw to reason about
- prevent “bring everything” behavior that removes dungeon preparation tradeoffs

### 6.1 Per-Battle And Per-Run Limits

Recommended limits:

- each character can consume at most `1` potion per round
- each character can consume at most `3` potions per dungeon run
- HP potion can be used in and out of combat within the dungeon run
- buff potions can only be consumed in combat

### 6.2 Stacking Rules

To avoid runaway scaling:

- same-family potion effects do not stack
- consuming a same-family potion refreshes duration and uses the stronger potency
- different-family buffs can coexist
- potion buffs stack multiplicatively with skills only if explicitly allowed by status definition

### 6.3 Trigger Timing

- HP potion resolves immediately when consumed
- buff potion applies a status at action resolution time
- speed potion affects next-round turn order because order is recalculated each round

## 7. Auto-Battle Policy (Bots)

Potion use should be deterministic and explainable.

Recommended default policy:

1. Use HP potion if current HP <= `35% max_hp` and potion is available
2. Use Defense potion before high incoming burst windows
3. Use Attack potion on boss or elite pressure windows
4. Use Speed potion when first-turn control race matters
5. Do not consume non-HP potions in trivial low-risk room states

All potion decisions must be written into structured battle logs.

## 8. Data Model Additions

Character inventory item metadata should include:

- `item_id`
- `item_type = potion`
- `potion_family`
- `tier`
- `effect_payload`
- `allowed_rank_min`
- `is_combat_only`

Run state should include:

- `potions_used_by_character`
- `remaining_run_potion_uses`
- `active_potion_statuses`

Battle log should include:

- `potion.consumed`
- `potion.effect_applied`
- `potion.effect_expired`

## 9. Economy And Shop Rules

- potions are purchased from building shops
- shop inventory is rank-gated by buyer character rank
- potion prices should create real trade-offs with gear repair and enhancement
- potion stock can refresh daily to prevent infinite cheap sustain loops

Recommended starting price bands:

- HP: low
- Defense: low-mid
- Attack: mid
- Speed: mid-high

## 10. Balance Guardrails

- potion system should increase strategic depth, not replace skill usage
- no potion should outscale equivalent skill effects by itself
- total expected potion value per run should be lower than the value of one major gear tier jump
- bosses should be balanced with the assumption that parties bring some, but not full, potion optimization

## 11. Integration Notes

This document updates and constrains:

- `10-combat-system-framework.md`
- `07-location-catalog-and-resource-definition.md`
- dungeon tuning sheets under `12` and `14`

Required engineering follow-up after design confirmation:

- remove automatic full HP restore between dungeon rooms
- implement potion item definitions and rank-gated shop inventory
- add potion consumption actions and log events
- add bot-policy defaults for potion use
