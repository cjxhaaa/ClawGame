## 9. Seasonal Leveling and Stat Framework

### 9.1 Goals

This module defines the season-based level system, which is the main long-form character growth axis.

Design goals:

- each season lasts `30` days and resets on a fixed server schedule
- bots should be able to reach max level in about `20` active days
- level-derived power must fully stabilize at max level
- the final `10` days of a season should emphasize materials, dungeon routing, and equipment optimization
- the stat model must remain bot-friendly, explicit, and deterministic enough for later combat automation

### 9.2 Relationship to Existing Progression

The project uses two progression currencies with different jobs:

- `Reputation`:
  - earned mostly from contract submission
  - spent on extra dungeon reward claims
- `Season level`:
  - driven by seasonal XP
  - controls the character's fixed level-based combat stats for the current season
- also gates the profession choice at level `10`

Rules:

- season level resets every `30` days
- seasonal XP resets every `30` days
- after a bot reaches level `100`, no more level-based stat growth is granted until the next season

### 9.3 Season Cadence

Each season is `30` days long.

Recommended time split:

- day `1-7`: onboarding, route discovery, and early contract acceleration
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
- total XP from level `1` to `100` is `103,647`
- this keeps level `10` reachable on day `1`, then slows the curve sharply enough that optimized bots approach cap around day `20`

XP required per next level should increase every level instead of staying flat.

Recommended formula:

- if current level `L` is `1-9`, `xp_to_next(L) = 420 + 20 x (L - 1)`
- if current level `L` is `10-99`, `xp_to_next(L) = round(880 + 2 x (L - 10) + 0.05 x (L - 10)^2)`

Reference points on this curve:

| Transition | XP to Next Level |
| --- | --- |
| 1 -> 2 | 420 |
| 5 -> 6 | 500 |
| 9 -> 10 | 580 |
| 10 -> 11 | 880 |
| 20 -> 21 | 905 |
| 40 -> 41 | 985 |
| 60 -> 61 | 1,105 |
| 80 -> 81 | 1,265 |
| 99 -> 100 | 1,454 |

Milestone totals:

- total XP from level `1` to `10`: `4,500`
- total XP from level `1` to `20`: `13,403`
- total XP from level `1` to `100`: `103,647`

Milestone target for active bots:

| Day | Expected Level Range |
| --- | --- |
| 1 | 10-11 |
| 3 | 21-23 |
| 7 | 42-47 |
| 10 | 57-62 |
| 15 | 78-85 |
| 20 | 97-100 |

This curve intentionally front-loads the civilian onboarding window so new bots can experience the level-10 profession decision immediately, then slows enough after that breakpoint to keep routing and build choices meaningful.

Career-shift expectation:

- most stable bots should reach level `10` on day `1`
- the civilian phase is therefore a short onboarding window rather than a multi-day progression lane
- profession choice should happen before the bot meaningfully enters the mid-season dungeon and gear-farming stage, but remaining civilian is still a valid option

### 9.5 XP Source Budget

The current implementation only grants seasonal XP from three system-level settlement points: quest submission, successful field-encounter resolution, and dungeon reward claims.

| Activity | Typical XP | Notes |
| --- | --- | --- |
| submit normal quest | 220 | covers common board contracts and normal quest templates |
| submit hard quest | 320 | covers uncommon board contracts and hard quest templates |
| submit nightmare quest | 520 | covers challenge board contracts and nightmare quest templates |
| resolve successful field encounter | 100 | fixed payout on field victory today |
| claim dungeon rewards with rating `D` | 280 | dungeon XP is paid at claim time, not at entry time |
| claim dungeon rewards with rating `C` | 340 | current dungeon XP depends on final rating |
| claim dungeon rewards with rating `B` | 400 | current dungeon XP depends on final rating |
| claim dungeon rewards with rating `A` | 460 | current dungeon XP depends on final rating |
| claim dungeon rewards with rating `S` | 520 | current dungeon XP depends on final rating |

Clarifications for the current live design:

- delivery routes, material turn-ins, curio follow-up deliveries, dungeon contracts, and dungeon-elite quests are currently quest templates, not separate XP systems
- there is currently no independent "elite field loop" XP payout
- there is currently no per-region first-completion XP bonus
- if new XP hooks are added later, they should be documented as future-state additions rather than mixed into the live budget table

Recommended daily target composition for a bot on pace for day-20 cap:

- `2-4` quest submissions across normal, hard, or nightmare contracts
- `1-2` dungeon reward claims with stable `B-A` ratings once the build can support them
- successful field encounters between travel segments to provide the fixed `100` XP fallback

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
   - defined by career stage before level `10`
   - defined by the current profession class after level `10`, or by the civilian profile if no profession is chosen
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

Canonical combat-stat definitions are documented in:

- `02-progression-and-combat.md`
- `10-combat-system-framework.md`

This module only defines how seasonal leveling contributes to the stat framework.

Level-growth baseline stats in this module:

| Field | Meaning in this module |
| --- | --- |
| `max_hp` | Main survival stat that scales steadily with season level. |
| `physical_attack` | Main physical-output growth stat. |
| `magic_attack` | Main magic-output growth stat. |
| `physical_defense` | Main physical-mitigation growth stat. |
| `magic_defense` | Main magic-mitigation growth stat. |
| `speed` | Main tempo stat that grows slowly across the season. |
| `healing_power` | Main healing-scaling growth stat. |

Extended combat stats such as `crit_rate`, `crit_damage`, `block_rate`, `precision`, `evasion_rate`, `physical_mastery`, and `magic_mastery` belong to the unified combat-stat model, but are not redefined here. Their growth should be introduced through equipment, set effects, passives, or later numeric extensions when needed.

### 9.9 Civilian Template and Early Growth

All characters begin the season as civilians and use one shared early template before promotion.

Suggested civilian base template at level `1`:

| Stage | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | 96 | 14 | 14 | 10 | 10 | 12 | 6 |

Suggested civilian balanced growth per level:

| Stage | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | +6.00 | +1.00 | +1.00 | +0.75 | +0.75 | +0.11 | +0.44 |

Design intent:

- civilians should feel viable in the early world without already revealing the final class identity
- the civilian template should be stable enough for quests, travel, and novice combat
- it should not outscale the dedicated profession profiles once level `10+` progression begins

### 9.10 Promotion at Level 10

When a character reaches level `10`, it may change class at the Adventurers Guild among `civilian`, `warrior`, `mage`, and `priest`.

Profession-change rules:

- each change costs `800` gold
- the current class sets the character's class identity for equipment and class-skill access
- moving from `civilian` into a promoted class grants one class-aligned starter weapon
- the class change rebases the character to the chosen class's current-level growth profile
- later stat growth always follows the current class rather than the previously selected class
- learned skill levels are preserved across class changes, but the active loadout removes skills that are unusable in the new class
- if the newly chosen class cannot use the currently equipped weapon, that weapon is automatically moved to inventory

Suggested class baselines at level `10`:

| Class | Recommended starter weapon | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | none | 150 | 23 | 23 | 17 | 17 | 13 | 10 |
| Warrior | `sword_shield` | 174 | 35 | 8 | 22 | 15 | 11 | 4 |
| Mage | `spellbook` | 132 | 12 | 45 | 12 | 22 | 17 | 10 |
| Priest | `holy_tome` | 150 | 11 | 33 | 14 | 23 | 14 | 24 |

### 9.11 Level Growth by Class

Every level from `2` to `100` grants fixed class-based growth.

Growth per level:

| Class | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | +6.00 | +1.00 | +1.00 | +0.75 | +0.75 | +0.11 | +0.44 |
| Warrior | +6.22 | +1.67 | +0.33 | +0.89 | +0.56 | +0.11 | +0.11 |
| Mage | +4.67 | +0.33 | +1.78 | +0.44 | +0.67 | +0.22 | +0.22 |
| Priest | +5.11 | +0.22 | +1.44 | +0.44 | +0.78 | +0.11 | +0.89 |

Rounding rule:

- growth is accumulated in decimal form internally
- final displayed battle stats are rounded to the nearest integer at each derived breakpoint

This keeps progression smooth while avoiding noisy level jumps.

Design intent:

- `civilian` is the balanced profile and remains a legal long-term choice
- `warrior` tilts heavily toward HP, physical attack, and physical defense
- `mage` tilts heavily toward magic attack and speed, while remaining fragile
- `priest` tilts toward sustain, healing, and magic defense
- the old nine-route model is not a separate profession layer in V1; those labels are better treated as skill tracks inside each class

### 9.12 Reference Max-Level Stats Without Gear

Approximate level `100` stats before equipment and temporary buffs:

| Class | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | 690 | 113 | 113 | 84 | 84 | 23 | 50 |
| Warrior | 734 | 185 | 38 | 102 | 65 | 21 | 14 |
| Mage | 552 | 42 | 205 | 52 | 82 | 37 | 30 |
| Priest | 610 | 31 | 163 | 54 | 93 | 24 | 104 |

These numbers are intentionally conservative for attack values and more generous for HP and defensive scaling, because equipment progression should remain meaningful after level cap.

### 9.13 Equipment System Foundations

The equipment system is built directly on top of the stat set above.

Equipment slots:

- `head`
- `chest`
- `necklace`
- `ring`
- `boots`
- `weapon`

Stat families by gear type:

- weapons:
  - `physical_attack`
  - `magic_attack`
  - `healing_power`
  - limited `speed`
- armor:
  - `max_hp`
  - `physical_defense`
  - `magic_defense`
- accessories:
  - `max_hp`
  - `speed`
  - `status_mastery`
  - `status_resistance`
  - build-specific bonuses

Equipment should not grant raw level replacement. It should amplify and specialize a max-level shell rather than override it.

### 9.14 Implementation Requirements

Backend storage keeps the following fields:

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

API payloads expose:

- season summary in `/me/state`
- XP gain records in action results
- season reset timestamp in public and private state payloads
