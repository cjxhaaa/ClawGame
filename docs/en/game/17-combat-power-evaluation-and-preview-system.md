# Combat Power Evaluation And Preview System

## 1. Goal

This specification defines a unified, explainable, and comparable combat power system for three scenarios:

- Character panel display
- Per-item equipment value display
- Pre-fight estimation for dungeons and arena

It is an evaluation system and does not replace real combat simulation outcomes.

## 2. Design Principles

- Explainable: every score has traceable sources
- Decomposable: total score can be split into stable components
- Comparable: same-stage characters are comparable
- Extensible: future affixes, sets, and talents can be integrated incrementally
- Stable: avoid excessive score jumps on minor gear changes

## 3. Total Score Model

Total combat power is defined as:

Total Combat Power = Base Growth Score + Equipment Score + Build Modifier + Scenario Modifier

Panel display score uses:

Panel Combat Power = Base Growth Score + Equipment Score + Build Modifier

Scenario modifier is excluded from the default panel to keep UI stable.

Usage rule:

- panel combat power is the single primary strength number shown on character panels, dungeon preparation surfaces, and arena entrant displays
- equipment score is a breakdown component and should be shown as supporting information rather than the main headline strength

## 4. Base Growth Score

### 4.1 Progression Coefficient

Open-era characters use a single baseline progression coefficient in normal play.

### 4.2 Level Growth

Level score:

Level Score = 20 × Level

Smooth growth bonus:

Smooth Score = 8 × sqrt(Level)

### 4.3 Stat Conversion

Stat score:

Stat Score =
1.2 × max_hp
+ 6.0 × primary_attack
+ 4.5 × secondary_attack
+ 5.0 × physical_defense
+ 5.0 × magic_defense
+ 4.0 × speed
+ 4.0 × healing_power

Primary attack by class:

- warrior: physical_attack
- mage: magic_attack
- priest: magic_attack

Final base score:

Base Growth Score = Progression Coefficient × (Level Score + Smooth Score + Stat Score)

## 5. Per-Item Equipment Score

Per-item score:

Item Score = Rarity Base + Main Stat Score + Sub Stat Score + Enhancement Score + Set Bonus Allocation

### 5.1 Quality Base

The seasonal equipment system uses the same five quality names as the dungeon and loot framework:

- Blue: 36
- Purple: 62
- Gold: 96
- Red: 144
- Prismatic: 188

Calibration rule:

- Blue is the baseline quality and should keep total equipment share near `30%-35%` of panel combat power once the season reaches stable mid-cycle progression
- Purple and Gold are transition upgrades and should not outscale level growth on their own
- Red is the main endgame set-chase quality
- Prismatic is the premium chase layer and should improve ceiling without invalidating the rest of the progression model
- a representative full Red set should usually land around `31%-33%` of endgame panel combat power
- a representative full Prismatic set should usually land around `33%-35%`, not meaningfully above it

### 5.2 Main/Sub Stat Coefficients

Main stat coefficients:

- max_hp: 1.2
- physical_attack: 6.0
- magic_attack: 6.0
- physical_defense: 5.0
- magic_defense: 5.0
- speed: 4.0
- healing_power: 4.0

Sub stat efficiency weights:

- effective: 1.0x
- semi-effective: 0.55x
- low-effective: 0.25x

### 5.3 Enhancement Score

Enhancement score curve:

Enhancement Score = 12 × EnhanceLevel + 1.8 × EnhanceLevel^2

### 5.4 Set Bonus Allocation

Allocated equally across set pieces:

- 2-piece active: total +60
- 4-piece active: additional +140
- 6-piece active: additional +220

## 6. Equipment Total

Equipment Total = sum of all equipped item scores

Recommended display:

- current equipment total
- candidate replacement delta

## 7. Build Modifier

Recommended range: -8% to +12% of total score.

### 7.1 Skill Loadout Modifier

- complete slot coverage with coherent cooldown structure: +0 to +120
- severe imbalance between offense and survivability: -0 to -80

### 7.2 Potion Readiness Modifier

Based on the potions available before dungeon or arena entry:

- recommended potion bundle prepared: +40 to +140
- missing key potions: 0 to -90

### 7.3 Resistance And Weakness Modifier

Critical weaknesses should apply penalties:

- speed far below scenario threshold: up to -120
- clearly underbuilt physical and magic defenses: up to -160

## 8. Scenario Preview Model

### 8.1 Dungeon Preview

Define recommended power bands per dungeon:

- lower bound: P50 clear line
- median: P70 stable clear line
- upper bound: P90 fast clear line

Dungeon Power = Panel Combat Power + Dungeon Scenario Modifier

Clear-confidence bands:

- below lower: low (<35%)
- lower to median: medium (35%-65%)
- median to upper: high (65%-85%)
- above upper: very high (>85%)

### 8.1.1 First Live Season Dungeon Bands

The first live season uses four parallel dungeons with shared pacing targets:

- `easy`: first reliable clear around day `5`
- `hard`: first reliable clear around day `10`
- `nightmare`: first reliable clear around day `15`

The table below gives the concrete panel combat power bands that should be written into dungeon definitions for V1.

| Dungeon | Difficulty | Target day | Lower bound | Median | Upper bound | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Ancient Catacomb | easy | day 5 | 2350 | 2800 | 3250 | most forgiving first-clear profile |
| Thorned Hollow | easy | day 5 | 2420 | 2880 | 3340 | slightly stricter on speed and hit consistency |
| Sunscar Warvault | easy | day 5 | 2460 | 2920 | 3380 | front-loaded damage check is a bit sharper |
| Obsidian Spire | easy | day 5 | 2440 | 2900 | 3360 | punishes weak sustain and poor skill rotation |
| Ancient Catacomb | hard | day 10 | 5350 | 6100 | 6850 | defensive set line makes this the gentlest hard entry |
| Thorned Hollow | hard | day 10 | 5480 | 6230 | 6980 | accuracy and turn-order pressure raise the floor |
| Sunscar Warvault | hard | day 10 | 5560 | 6310 | 7060 | elite burst windows increase wipe risk |
| Obsidian Spire | hard | day 10 | 5520 | 6270 | 7020 | repeated caster pressure stresses cooldown planning |
| Ancient Catacomb | nightmare | day 15 | 8900 | 10050 | 11200 | preferred first nightmare farm for stable bots |
| Thorned Hollow | nightmare | day 15 | 9080 | 10230 | 11380 | needs cleaner precision and crit scaling to stabilize |
| Sunscar Warvault | nightmare | day 15 | 9200 | 10350 | 11500 | highest physical burst check among the four |
| Obsidian Spire | nightmare | day 15 | 9140 | 10290 | 11440 | hardest on mana-tempo and healing throughput |

Interpretation:

- lower bound means a roughly `P50` clear line and is acceptable for first attempts
- median means a stable farming line and should be the default recommendation for bots
- upper bound means high-confidence speed clears and is the point where bots should start optimizing for affixes rather than raw survivability

Implementation note:

- store these as `recommended_power_floor`, `recommended_power_mid`, and `recommended_power_ceiling`
- dungeon UI and bots should prefer `median` as the "recommended combat power" headline
- if real clear data drifts by more than `5-7%`, recalibrate the table instead of changing the formula weights first

### 8.2 Arena Preview

For two sides A and B:

Delta = A - B

Win probability mapping:

WinRate(A) = 1 / (1 + exp(-Delta / 420))

UI should show confidence tiers instead of exact percentages:

- hard disadvantage: <30%
- disadvantage: 30%-45%
- close: 45%-55%
- advantage: 55%-70%
- strong advantage: >70%

Arena usage rule:

- arena signup, entrant cards, bracket cards, and arena leaderboards should use panel combat power as the visible strength field
- equipment score may be exposed as a secondary breakdown value but should not be the main arena score label

## 9. UI Display Requirements

### 9.1 Priority In Status Bar

Show first in character status bar:

- total combat power
- delta from last update
- scene preview tag (dungeon confidence or arena advantage)

### 9.2 Character Panel Breakdown

Panel sections:

- base growth score
- equipment total score
- build modifier

### 9.3 Equipment Card Display

Per item:

- item combat score
- same-slot delta against currently equipped item
- expected total score impact

## 10. Calibration

Track these curves:

- average combat power by class and level
- dungeon recommended power vs real clear rate
- arena power delta vs real win rate

Run seasonal recalibration to avoid drift.

## 11. Example

Example character:

- level 30, open-era baseline progression
- base growth score = 4280
- equipment score = 2360
- build modifier = 180

Panel combat power:

- 6820

Dungeon scenario modifier = -120:

- dungeon power = 6700
- target dungeon band = 6400 to 7200
- preview = medium-high confidence

## 12. Integration Points

- character base stats from character state snapshot
- per-item score computed and cached by inventory service
- dungeon thresholds from dungeon definition tables
- arena preview computed at match preview time

## 13. Non-Goals

- not replacing real combat simulation and logs
- not directly deciding rewards
- not serving as sole anti-cheat signal

## 14. Versioning

Maintain explicit formula versioning, e.g. power_score_v1_0.
Return formula version in backend responses for explainability parity.
