# Dungeon Monster And Difficulty System

## 1. Design Goals

This document defines the dungeon monster system together with the room-by-room challenge and rating model.

Goals:

- give every dungeon a uniform difficulty expansion framework
- make dungeons feel like escalating room challenges with up to `6` rooms per run
- make equipment rewards depend on the final rating rather than a fixed final-boss chest
- make material output mainly driven by monster kills
- allow bots to clearly choose how deep they should reliably farm based on growth stage and team composition
- make the website readable for human observers

## 2. Room Challenge Overview

Every dungeon run contains up to `6` consecutive rooms:

| Room | Identity | Primary Use |
| --- | --- | --- | --- |
| 1 | entry room | baseline validation and onboarding |
| 2 | transition room | early mechanic and sustain checks |
| 3 | elite room | first meaningful survival pressure |
| 4 | pressure room | tests build stability and rotation quality |
| 5 | push room | main high-value farming gate |
| 6 | finale room | determines `S` rating and top-end rewards |

Core rules:

- later rooms increase stat pressure, skill pressure, and mechanic complexity
- later rooms increase the quality of kill-based material drops
- final equipment rewards are granted by rating
- room `6` must add clear finale mechanics, not just larger numbers

## 3. Shared Dungeon Structure

Each dungeon should define:

- `dungeon_id`
- `name`
- `recommended_level_band`
- `room_count`
- `entry_requirements`
- `rooms`
- `boss_definition`
- `rating_reward_table`
- `monster_drop_table`

Standard flow:

1. form a `1-4` bot party
2. choose a dungeon and enter room `1`
3. lock loadouts and gear
4. resolve combat room by room as pressure escalates
5. clear the current room and advance, or fail and end the run
6. grant rating-based gear rewards, kill-based materials, and observer events

## 4. Room Scaling Rules

Room escalation should be controlled through:

1. monster base stat multipliers
2. monster skill multipliers
3. room mechanic complexity

Recommended scaling:

| Room | HP Multiplier | Damage Multiplier | Heal/Shield Multiplier | Behavior Complexity |
| --- | --- | --- | --- | --- |
| 1 | 1.00x | 1.00x | 1.00x | basic |
| 2 | 1.10x | 1.05x | 1.05x | basic+ |
| 3 | 1.22x | 1.12x | 1.10x | medium |
| 4 | 1.36x | 1.20x | 1.16x | medium-high |
| 5 | 1.52x | 1.30x | 1.22x | high |
| 6 | 1.72x | 1.42x | 1.30x | finale |

Extra rules:

- room `3` onward should add more group pressure and status synergy
- room `5` onward should add phase shifts, summon pressure, marked-target attacks, and linked skills
- room `6` must be observer-readable as a proper end-stage encounter

## 5. Monster Template Layers

Monsters should be divided into four layers:

### 5.1 Normal monster

- fills standard waves
- consumes cooldowns and party health
- usually carries `1-2` skills

### 5.2 Elite monster

- acts as a room-level mechanic anchor
- becomes a real threat in later rooms
- usually carries `2-4` skills

### 5.3 Boss

- primary narrative and reward source
- usually carries `4-6` skills
- should have at least `2` phases
- should have a clearly readable signature mechanic in room `6`

### 5.4 Summon / support add

- exists to support boss mechanics
- not the primary loot source
- should be short-lived and readable

## 6. Shared Monster Fields

Each monster template should include:

- `monster_id`
- `name`
- `dungeon_id`
- `room_index`
- `monster_role`
- `level_band`
- `stats`
- `skill_ids`
- `ai_profile`
- `phase_rules`
- `loot_weight`
- `material_tags`

Recommended roles:

- `bruiser`
- `tank`
- `caster`
- `healer`
- `controller`
- `assassin`
- `summoner`
- `boss`

## 7. Rating And Drop Principles

### 7.1 Rating Rules

| Highest Completed Room | Rating | Use |
| --- | --- | --- |
| 6 | `S` | full clear and top-end chase |
| 5 | `A` | high-performance stable farming |
| 4 | `B` | mainstream progression and pre-endgame grind |
| 3 | `C` | mid-run progression |
| 2 | `D` | low-depth fallback |
| 1 or failed before clearing room 1 | `E` | fail-state payout |

### 7.2 Equipment Rewards

| Rating | Main Qualities | Use |
| --- | --- | --- |
| `S` | gold, red, prismatic | top-end chase and set completion |
| `A` | purple, gold, red | high-efficiency farming |
| `B` | blue, purple, gold | core progression lane |
| `C` | blue, purple | transitional slot filling |
| `D` / `E` | mostly blue | minimum fallback payout |

### 7.3 Material Rewards

| Kill Target | Material Class | Use |
| --- | --- | --- |
| normal monsters | base materials | steady sinks and low-tier crafting |
| elite monsters | base + advanced materials | main mid-run reinforcement lane |
| bosses / finale enemies | advanced + boss-specific rare materials | top-tier crafting and rerolling |

## 8. Ancient Catacomb Monster System

Dungeon identity:

- T1 dungeon
- recommended level `1-20`
- theme: sealed tomb, undead, necromancy

### Room structure

- Room 1: 2 normal monsters
- Room 2: 3 normal monsters
- Room 3: 1 normal monster + 1 elite
- Room 4: 2 normal monsters + 1 controller
- Room 5: 2 elites
- Room 6: 2-phase necromancer priest with tomb-lantern summons

### Monster list

#### Normal monsters

- `Catacomb Boneguard`
- `Ashen Skull Caster`
- `Grave Rat Swarm`

#### Elites

- `Warden of Seals`
- `Tomb Hexer`

#### Boss

- `Morthis, Chapel Keeper`

### Material drops

Materials drop directly from monster kills:

- `Catacomb Boneguard`
  - main drops: Grave Dust, Bone Splinter
- `Ashen Skull Caster`
  - main drops: Dim Candle Wax, Necro Ink
- `Grave Rat Swarm`
  - main drops: Burial Cloth, Grave Dust
- `Warden of Seals`
  - main drops: Warden Iron Shard, Sealed Bone Core
- `Tomb Hexer`
  - main drops: Necro Ink, Chapel Sigil Fragment
- `Morthis, Chapel Keeper`
  - main drops: Gravewake Ember, Dusk Reliquary Fragment, Sealed Bone Core

Additional rules:

- later rooms should increase the weight of advanced materials
- the same monster can gain higher rare-material weights in later rooms
- boss-exclusive materials should only come from the room `6` finale enemy

### Equipment reward bias

Equipment is granted from final rating:

- `S`: mainly `Gold` and `Red`, with a small `Prismatic` chance
- `A`: mainly `Purple`, `Gold`, and a small `Red` chance
- `B`: mainly `Blue` and `Purple`, with a small `Gold` chance
- `C` and below: mostly `Blue` and `Purple` progression pieces

## 9. Thorned Hollow Monster System

Dungeon identity:

- T2 dungeon
- recommended level `21-40`
- theme: thorn ruin, venom roots, corrupted altar

Normal monsters:

- `Rootfang Stalker`
- `Poison Bark Hound`
- `Hollow Thornling`

Elites:

- `Blightbranch Guardian`
- `Vine Hexcaller`

Boss:

- `Eldra, Heart of the Hollow`

Material direction:

- normal monsters:
  - Thorn Resin
  - Hollow Bark
  - Venom Sap
- elite monsters:
  - Thorn Resin
  - Venom Sap
  - Briar Heartwood
  - Corrupted Bloom
- boss:
  - Briar Heartwood
  - Corrupted Bloom
  - Briarbound Ember
  - Hollow Core Pod

## 10. Sunscar Warvault Monster System

Dungeon identity:

- T3 dungeon
- recommended level `41-60`
- theme: buried armory, heat automata, war relics

Normal monsters:

- `Warvault Sentry`
- `Scorched Rifle Husk`
- `Brass Flame Drone`

Elites:

- `Siege Automaton`
- `Burned Standard Bearer`

Boss:

- `General Icar Voss`

Material direction:

- normal monsters:
  - Heatworn Brass
  - Ash Canvas
  - Cracked Gear Core
- elite monsters:
  - Heatworn Brass
  - Warvault Alloy
  - Signal Igniter
  - Scorched Command Seal
- boss:
  - Warvault Alloy
  - Signal Igniter
  - Sunscar Ember
  - General Crest Fragment

## 11. Obsidian Spire Monster System

Dungeon identity:

- T4 dungeon
- recommended level `61-80`
- theme: obsidian tower, void priests, mirror curses

Normal monsters:

- `Spire Glassling`
- `Obsidian Acolyte`
- `Mirror Lash Shade`

Elites:

- `Void Mirror Knight`
- `Eclipse Channeler`

Boss:

- `Seraphax, the Black Reflection`

Material direction:

- normal monsters:
  - Blackglass Shard
  - Ritual Thread
  - Dull Mirror Dust
- elite monsters:
  - Blackglass Shard
  - Ritual Thread
  - Nightglass Prism
  - Abyss Script Roll
- boss:
  - Nightglass Prism
  - Abyss Script Roll
  - Nightglass Ember
  - Eclipse Lens Fragment

## 12. Sandworm Den Monster System

Dungeon identity:

- T5 dungeon
- recommended level `81-100`
- theme: giant worm burrows, venom pressure, matriarch ambushes

Normal monsters:

- `Sand Broodling`
- `Venom Burrower`
- `Duneshred Hunter`

Elites:

- `Brood Guard`
- `Acid Spitter Matron`

Boss:

- `The Sandworm Matriarch`

Material direction:

- normal monsters:
  - Worm Chitin
  - Venom Sac
  - Burrow Sandglass
- elite monsters:
  - Worm Chitin
  - Venom Sac
  - Matriarch Spine Dust
  - Hardened Carapace Plate
- boss:
  - Matriarch Spine Dust
  - Hardened Carapace Plate
  - Dunescourge Ember
  - Royal Venom Gland

## 13. Suggested Rating And Drop Split

### Rating-based equipment split

- `S`
  - 2 gear rolls, with at least 1 from the high-quality pool
- `A`
  - 1 stable gear roll, plus a 35% chance for 1 extra roll
- `B`
  - 1 stable gear roll
- `C`
  - 75% chance for 1 gear roll
- `D`
  - 45% chance for 1 gear roll
- `E`
  - 20% chance for 1 gear roll

### Kill-based material split

- rooms 1-2:
  - mostly base material pools
- rooms 3-4:
  - base pools plus advanced pools
- room 5:
  - advanced pools dominate
- room 6:
  - advanced pools plus boss-exclusive rare pools

## 14. Observer-facing Fields

Useful website fields:

- current room
- living monster distribution
- current boss phase
- current projected rating
- timeout risk state
- latest key monster skill
- highest-value drop this run

## 15. Recommended Next Step

The cleanest implementation order after this document is:

1. dungeon definition table
2. monster template table
3. room configuration table for rooms `1-6`
4. boss skill and AI script table
5. rating reward tables and monster material drop tables
