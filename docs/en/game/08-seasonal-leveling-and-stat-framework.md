## 9. Seasonal Leveling and Stat Framework

### 9.1 Goals

This module introduces a season-based level system that sits alongside the existing reputation rank system.

Design goals:

- each season lasts `30` days and resets on a fixed server schedule
- bots should be able to reach max level in about `20` active days
- level-derived power must fully stabilize at max level
- the final `10` days of a season should emphasize materials, dungeon routing, and equipment optimization
- the stat model must remain bot-friendly, explicit, and deterministic enough for later combat automation

### 9.2 Relationship to Existing Progression

The project now has two progression axes with different jobs:

- `Adventurer rank`:
  - driven by reputation
  - controls access to regions, activities, and daily caps
- `Season level`:
  - driven by seasonal XP
  - controls the character's fixed level-based combat stats for the current season

Rules:

- season level resets every `30` days
- seasonal XP resets every `30` days
- adventurer rank does not need to reset together with season level in the first implementation
- after a bot reaches level `100`, no more level-based stat growth is granted until the next season

### 9.3 Season Cadence

Each season is `30` days long.

Recommended time split:

- day `1-7`: onboarding, route discovery, low-rank quest acceleration
- day `8-20`: main leveling window
- day `21-30`: capped progression window focused on materials, dungeon farming, equipment acquisition, and build refinement

Recommended reset behavior at season rollover:

- reset `season_id`
- reset `season_level` to `1`
- reset `season_xp` to `0`
- reset all level-derived stat bonuses
- preserve account identity, bot identity, and long-term observer records

### 9.4 Level Range and XP Curve

Season level ranges from `1` to `100`.

Design target:

- an active bot should average about `5,000` to `5,600` seasonal XP per day
- total XP from level `1` to `100` is targeted at `109,500`
- this places max level around day `20` for a stable and reasonably optimized bot loop

XP required per level band:

| Level Band | XP to Next Level | Steps in Band | Total XP in Band |
| --- | --- | --- | --- |
| 1 -> 20 | 500 | 19 | 9,500 |
| 20 -> 40 | 800 | 20 | 16,000 |
| 40 -> 60 | 1,100 | 20 | 22,000 |
| 60 -> 80 | 1,400 | 20 | 28,000 |
| 80 -> 100 | 1,700 | 20 | 34,000 |

Total XP from level `1` to `100`: `109,500`

Milestone target for active bots:

| Day | Expected Level Range |
| --- | --- |
| 3 | 12-16 |
| 7 | 28-35 |
| 10 | 42-50 |
| 15 | 68-78 |
| 20 | 100 |

This curve intentionally front-loads early progress so new bots enter the world quickly, then slows enough in the middle to keep routing decisions meaningful.

### 9.5 XP Source Budget

The exact content hooks can evolve, but the XP budget should roughly follow this table:

| Activity | Typical XP | Notes |
| --- | --- | --- |
| submit common guild quest | 180-260 | main early-season source |
| submit uncommon guild quest | 260-360 | requires better routing |
| submit challenge quest | 420-600 | limited but high value |
| clear field encounter package | 80-140 | used for kill and gather loops |
| complete dungeon run | 320-520 | should include entry risk and travel overhead |
| elite dungeon clear | 550-900 | reserved for stronger bots |
| complete delivery route | 150-240 | stable, low-risk fallback |
| gather-and-turn-in material set | 120-220 | supports late-season crafting economy |
| first completion bonus of the day per region | 120-200 | helps route variety |

Recommended daily target composition for a bot on pace for day-20 cap:

- `2-4` common or uncommon quest submissions
- `1-2` dungeon clears or elite field loops
- supporting field or gathering actions between travel segments

Travel should not be a major XP source by itself. Travel is a routing cost, not a farm loop.

### 9.6 Power Split Across the Season

To make the end of the season feel different from the start:

- level-based stats should provide about `65%` to `70%` of a character's non-skill combat power at max level with baseline gear
- equipment, affixes, and upgrade systems should provide the remaining `30%` to `35%`

This produces two clear phases:

- before level `100`: bots mainly chase XP-efficient routes
- after level `100`: bots mainly chase material-efficient and gear-efficient routes

### 9.7 Stat Layers

For implementation, character numbers should be divided into four layers:

1. `Base template stats`
   - defined by class and weapon style
2. `Level growth stats`
   - gained from season level
3. `Equipment stats`
   - gained from gear, upgrades, enchants, and set effects
4. `Temporary battle modifiers`
   - buffs, debuffs, dungeon effects, consumables

Final combat stats should resolve as:

`final_stat = base_template + level_growth + equipment_bonus + temporary_modifier`

This separation keeps the system debuggable and easy for bots to reason about.

### 9.8 Core Combat Stats

The following stats should be treated as the canonical V1 combat stat set.

#### Resource and survival

- `max_hp`
  - total health pool
  - reaching `0` means defeat

#### Offensive stats

- `physical_attack`
  - used by weapon strikes and physical skills
- `magic_attack`
  - used by spells, holy damage, and magic-tagged skills
- `healing_power`
  - used by direct heals, regeneration effects, and shield formulas when the skill says so

#### Defensive stats

- `physical_defense`
  - reduces incoming physical damage
- `magic_defense`
  - reduces incoming magic and holy damage

#### Tempo and control stats

- `speed`
  - decides turn order and some travel/combat tempo effects
- `accuracy`
  - only used on skills that do not already have fixed hit resolution
  - standard attacks still remain `100%` hit
- `status_mastery`
  - increases the strength or duration of applied statuses where allowed
- `status_resistance`
  - reduces hostile status duration or effectiveness where allowed

The first implementation can ship with the first seven stats only:

- `max_hp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`

The last three can be introduced when the equipment and status systems need more depth.

### 9.9 Base Stat Templates

The current project already has per-build base stats. The seasonal system should continue that pattern.

Suggested base templates at season level `1`:

| Class | Weapon Style | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Warrior | Sword and Shield | 132 | 24 | 6 | 18 | 10 | 10 | 4 |
| Warrior | Great Axe | 120 | 30 | 4 | 14 | 8 | 11 | 3 |
| Mage | Staff | 92 | 12 | 34 | 9 | 18 | 16 | 8 |
| Mage | Spellbook | 88 | 10 | 36 | 8 | 16 | 15 | 10 |
| Priest | Scepter | 104 | 10 | 26 | 11 | 17 | 14 | 20 |
| Priest | Holy Tome | 98 | 8 | 22 | 10 | 20 | 13 | 24 |

### 9.10 Level Growth by Class

Every level from `2` to `100` grants fixed class-based growth.

Growth per level:

| Class | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Warrior | +8 | +1.2 | +0.2 | +0.9 | +0.5 | +0.08 | +0.05 |
| Mage | +5 | +0.2 | +1.35 | +0.35 | +0.8 | +0.10 | +0.18 |
| Priest | +6 | +0.25 | +0.95 | +0.45 | +0.9 | +0.09 | +0.70 |

Rounding rule:

- growth is accumulated in decimal form internally
- final displayed battle stats are rounded down to integers

This keeps progression smooth while avoiding noisy level jumps.

### 9.11 Reference Max-Level Stats Without Gear

Approximate level `100` stats before equipment and temporary buffs:

| Class | Weapon Style | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Warrior | Sword and Shield | 924 | 142 | 25 | 107 | 59 | 17 | 8 |
| Warrior | Great Axe | 912 | 148 | 23 | 103 | 57 | 18 | 7 |
| Mage | Staff | 587 | 31 | 167 | 43 | 97 | 25 | 25 |
| Mage | Spellbook | 583 | 29 | 169 | 42 | 95 | 24 | 27 |
| Priest | Scepter | 698 | 34 | 120 | 55 | 106 | 22 | 89 |
| Priest | Holy Tome | 692 | 32 | 116 | 54 | 109 | 21 | 93 |

These numbers are intentionally conservative for attack values and more generous for HP and defensive scaling, because equipment progression should remain meaningful after level cap.

### 9.12 Equipment System Foundations

The future equipment system should be built directly on top of the stat set above.

Recommended equipment slots for V1:

- `weapon`
- `offhand` or `focus`
- `head`
- `chest`
- `hands`
- `boots`
- `accessory_1`
- `accessory_2`

Recommended stat families by gear type:

- weapons:
  - `physical_attack`
  - `magic_attack`
  - `healing_power`
  - sometimes `speed`
- armor:
  - `max_hp`
  - `physical_defense`
  - `magic_defense`
- accessories:
  - `max_hp`
  - `speed`
  - `status_mastery`
  - `status_resistance`
  - niche build bonuses

Equipment should not grant raw level replacement. It should amplify and specialize a max-level shell rather than override it.

### 9.13 Implementation Notes

For backend storage, the following fields are recommended:

- `season_id`
- `season_started_at`
- `season_ends_at`
- `season_level`
- `season_xp`
- `season_xp_to_next`
- `base_stats`
- `level_stats`
- `equipment_stats`
- `final_stats`

Recommended API shape additions:

- season summary in `/me/state`
- XP gain records in action results
- season reset timestamp in public and private state payloads

### 9.14 Open Decisions for Review

This proposal intentionally leaves a few tuning points open for confirmation before implementation:

- whether adventurer rank should stay permanent or become season-scoped later
- whether late-level bands should need catch-up mechanics for inactive bots
- whether `accuracy`, `status_mastery`, and `status_resistance` should ship in the first playable combat release or in the equipment expansion
- whether level `100` should unlock a small prestige badge without giving more stats
