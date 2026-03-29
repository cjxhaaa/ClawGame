# Dungeon Monster And Difficulty System

## 1. Design Goals

This document defines the dungeon monster composition, room progression, and three-tier difficulty rules. It upgrades dungeon combat from single-enemy fights to multi-enemy group encounters.

Core goals:

- rooms default to multi-enemy compositions, not single enemies
- monsters are tiered: normal, elite, boss
- each room has a clear tactical theme with escalating pressure
- boss only appears in the final room (room 6)
- three difficulty tiers (easy / hard / nightmare) have clear stat and mechanic differences
- nightmare must noticeably raise the bar, requiring better gear and build completeness

## 2. Room Structure

Every dungeon run has exactly `6` rooms:

| Room | Role | Design Intent |
| --- | --- | --- |
| 1 | entry room | set the pace, allow warm-up |
| 2 | pressure room | introduce two-core enemy combinations |
| 3 | elite room | first elite-anchored pressure check |
| 4 | mechanic room | elite + control or summon synergy |
| 5 | high-pressure room | resource check, forces potions and cooldown management |
| 6 | boss room | boss finale (may include adds), determines clear rating |

Hard constraints:

- rooms `1` to `5` must not contain boss-tier monsters
- boss is only allowed in room `6`
- room `6` may include boss + adds, but boss must be the primary target

## 3. Monster Tiers And Roles

## 3.1 Normal Monster

Role: fill squad slots, apply sustained pressure, consume player skill cooldowns.

Design requirements:

- `1-2` skills
- single-target or light AoE
- clear position identity (frontline, backline DPS, harasser, status applier)

## 3.2 Elite Monster

Role: serves as the tactical anchor of a room, forcing players to adjust skill priorities.

Design requirements:

- `2-4` skills
- at least 1 mechanic skill (control, shield, debuff, summon, burst telegraphing)
- must provide combo value with normal monsters

## 3.3 Boss

Role: dungeon finale target.

Design requirements:

- only in room `6`
- at least 2 phases
- each phase has a recognizable mechanic theme
- nightmare adds new mechanics or increases key skill frequency

## 4. Multi-Enemy Composition Rules

## 4.1 Monster Count Per Room

| Difficulty | Rooms 1-2 | Rooms 3-4 | Room 5 | Room 6 |
| --- | --- | --- | --- | --- |
| easy | 2-3 | 2-3 | 2-3 | Boss + 0-1 adds |
| hard | 3-4 | 3-4 | 3-4 | Boss + 1-2 adds |
| nightmare | 4-5 | 4-5 | 4-5 | Boss + 2 adds (or phase summons) |

## 4.2 Composition Guidelines

- easy: normal monsters dominate; elites only appear in rooms 3/4/5
- hard: room 3 onward must include at least 1 elite per room
- nightmare: elites can appear from room 2 onward; rooms 4-5 commonly feature double-elite setups

## 5. Three-Difficulty Framework

## 5.1 Stat Multipliers

Multipliers are relative to the easy baseline for the same dungeon:

| Difficulty | HP | Damage | Defense | Speed |
| --- | --- | --- | --- | --- |
| easy | 1.00x | 1.00x | 1.00x | 1.00x |
| hard | 1.28x | 1.22x | 1.15x | 1.08x |
| nightmare | 1.62x | 1.50x | 1.35x | 1.16x |

## 5.2 Mechanic Density

| Difficulty | Control Density | AoE Density | Summon/Synergy Density |
| --- | --- | --- | --- |
| easy | low | low | low |
| hard | medium | medium | medium |
| nightmare | high | high | high |

Nightmare constraints:

- difficulty must not come only from stat inflation
- must include composition synergies (for example: frontline damage reduction + backline burst, elite applies vulnerable + AoE follow-up)
- expected to require good gear + reasonable potions + complete build to clear reliably

## 6. Room Progression Design Rules

## 6.1 General Escalation Logic

1. room 1: single-mechanic warm-up (for example pure frontline + one backline)
2. room 2: two-mechanic combination (for example slow + burst)
3. room 3: first elite mechanic check
4. room 4: control or summon mixed pressure
5. room 5: resource drain room (force potions, force cooldown rotation)
6. room 6: boss phase encounter

## 6.2 Boss Room Rules

- boss must have a phase-switch trigger (HP threshold or round threshold)
- adds in the boss room exist only to reinforce boss mechanics
- boss defeat ends the dungeon run

## 7. Drops And Difficulty

- equipment rewards are resolved by final clear rating
- materials drop from monster kills
- higher difficulty increases weight of rare materials
- nightmare improves access to red and prismatic material resources, but should correspond to higher fail risk

## 8. Current Dungeon Coverage

This spec currently covers and requires implementation for:

- `ancient_catacomb_v1`
- `sandworm_den_v1`

Each dungeon must provide:

- a named normal / elite / boss monster list
- easy / hard / nightmare complete 6-room composition tables
- explicit annotation that boss only appears in room 6

## 9. Implementation Document Links

- per-room composition and balance tables: `14-first-batch-dungeon-balance-sheets.md`
- data fields and schema: `13-dungeon-data-tables-and-template-spec.md`
