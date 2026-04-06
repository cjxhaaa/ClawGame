## 9. Seasonal Leveling and Stat Framework

### 9.1 Goals

This module defines the season-based level system, which is now the main long-form character growth axis.

Design goals:

- each season lasts `30` days and resets on a fixed server schedule
- bots should be able to reach max level in about `20` active days
- level-derived power must fully stabilize at max level
- the final `10` days of a season should emphasize materials, dungeon routing, and equipment optimization
- the stat model must remain bot-friendly, explicit, and deterministic enough for later combat automation

### 9.2 Relationship to Existing Progression

The project now has two progression currencies with different jobs:

- `Reputation`:
  - earned mostly from contract submission
  - spent on extra dungeon reward claims
- `Season level`:
  - driven by seasonal XP
  - controls the character's fixed level-based combat stats for the current season
  - also gates the profession-route choice at level `10`

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

Career-shift expectation:

- most stable bots should reach level `10` around day `2-3`
- the civilian phase is therefore short, but long enough to create a readable onboarding and unlock window
- profession choice should happen before the bot meaningfully enters the hard dungeon and gear-farming stage

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
   - defined by career stage before level `10`
   - defined by chosen profession route after level `10`
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

### 9.9 Civilian Template and Early Growth

All characters begin the season as civilians and use one shared early template before promotion.

Suggested civilian base template at level `1`:

| Stage | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | 96 | 14 | 14 | 10 | 10 | 12 | 6 |

Suggested civilian growth from level `2` to `9`:

| Stage | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Civilian | +5 | +0.55 | +0.55 | +0.40 | +0.40 | +0.07 | +0.12 |

Design intent:

- civilians should feel viable in the early world without already revealing the final class identity
- the civilian template should be stable enough for quests, travel, and novice combat
- it should not outscale the promoted routes once level `10+` progression begins

### 9.10 Promotion at Level 10

When a character reaches level `10`, it chooses exactly one profession route at the Adventurers Guild.

Promotion rules:

- the chosen route sets the character's class identity for equipment and class-skill access
- the game grants one route-aligned starter weapon at promotion time
- the promotion also rebases the character to the route's intended level-10 baseline
- after promotion, later stat growth follows the chosen route rather than the civilian template

Suggested route baselines at level `10`:

| Class | Route | Recommended starter weapon | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Warrior | `tank` | `sword_shield` | 180 | 34 | 8 | 26 | 16 | 11 | 4 |
| Warrior | `physical_burst` | `great_axe` | 168 | 42 | 6 | 20 | 13 | 12 | 3 |
| Warrior | `magic_burst` | `sword_shield` | 164 | 28 | 26 | 19 | 18 | 11 | 5 |
| Mage | `single_burst` | `spellbook` | 128 | 12 | 48 | 11 | 21 | 17 | 11 |
| Mage | `aoe_burst` | `staff` | 132 | 13 | 46 | 12 | 22 | 16 | 9 |
| Mage | `control` | `spellbook` | 136 | 11 | 41 | 13 | 23 | 18 | 10 |
| Priest | `healing_support` | `holy_tome` | 150 | 10 | 31 | 14 | 24 | 14 | 30 |
| Priest | `curse` | `scepter` | 148 | 12 | 36 | 14 | 22 | 15 | 19 |
| Priest | `summon` | `holy_tome` | 152 | 11 | 33 | 15 | 23 | 14 | 24 |

### 9.11 Level Growth by Profession Route

Every level from `11` to `100` grants fixed route-based growth.

Growth per level:

| Class | Route | HP | P.Atk | M.Atk | P.Def | M.Def | Speed | Heal |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Warrior | `tank` | +8.2 | +1.00 | +0.20 | +0.95 | +0.55 | +0.07 | +0.05 |
| Warrior | `physical_burst` | +8.0 | +1.32 | +0.12 | +0.72 | +0.42 | +0.08 | +0.03 |
| Warrior | `magic_burst` | +7.4 | +0.75 | +0.78 | +0.68 | +0.62 | +0.08 | +0.08 |
| Mage | `single_burst` | +5.1 | +0.18 | +1.42 | +0.32 | +0.72 | +0.10 | +0.16 |
| Mage | `aoe_burst` | +5.3 | +0.16 | +1.30 | +0.35 | +0.80 | +0.09 | +0.18 |
| Mage | `control` | +5.4 | +0.10 | +1.16 | +0.40 | +0.86 | +0.11 | +0.16 |
| Priest | `healing_support` | +6.2 | +0.18 | +0.86 | +0.42 | +0.96 | +0.08 | +0.80 |
| Priest | `curse` | +6.0 | +0.25 | +1.02 | +0.45 | +0.86 | +0.09 | +0.46 |
| Priest | `summon` | +6.1 | +0.22 | +0.92 | +0.50 | +0.91 | +0.08 | +0.58 |

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
