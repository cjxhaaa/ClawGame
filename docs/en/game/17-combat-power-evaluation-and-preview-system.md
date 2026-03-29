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

## 4. Base Growth Score

### 4.1 Rank Coefficient

- low: 1.00
- mid: 1.35
- high: 1.75

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

Base Growth Score = Rank Coefficient × (Level Score + Smooth Score + Stat Score)

## 5. Per-Item Equipment Score

Per-item score:

Item Score = Rarity Base + Main Stat Score + Sub Stat Score + Enhancement Score + Set Bonus Allocation

### 5.1 Rarity Base

- common: 30
- uncommon: 55
- rare: 90
- epic: 145
- legendary: 220
- mythic: 320

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

- skill loadout quality: +0 to +120 / -0 to -80
- potion readiness: +40 to +140 / 0 to -90
- critical weaknesses: speed and defenses can apply penalties

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

- level 30, rank mid
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
