# First Batch Dungeon Balance Sheets

## 1. Goal

This document turns the dungeon specs into first-pass implementation sheets.

Current scope:

- `Ancient Catacomb`
- `Thorned Hollow`

Goals:

- provide baseline monster stats
- define monster skills and cooldowns
- define AI behavior
- define wave layouts by difficulty
- define equipment and material drop weights

## 2. Shared Balance Scale

### 2.1 T1 Scale

For `Ancient Catacomb`:

- normal monster HP: `90-150`
- elite monster HP: `220-320`
- boss HP: `520-760`
- normal damage: `16-30`
- elite damage: `28-48`
- boss core skill damage: `40-72`

### 2.2 T2 Scale

For `Thorned Hollow`:

- normal monster HP: `180-260`
- elite monster HP: `360-520`
- boss HP: `860-1180`
- normal damage: `28-46`
- elite damage: `42-68`
- boss core skill damage: `62-108`

## 3. Ancient Catacomb Monster Sheets

### 3.1 Normal: Catacomb Boneguard

Role:

- `bruiser`
- front-line stable damage

Base stats:

| Stat | Value |
| --- | --- |
| HP | 132 |
| P.Atk | 18 |
| M.Atk | 0 |
| P.Def | 10 |
| M.Def | 6 |
| Speed | 8 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Rust Slash` | `skill_attack` | `normal` | 1 | single-target physical hit |
| `Bone Brace` | `buff` | `advanced` | 2 | self `defense_up` for 2 rounds |

AI:

1. if no `defense_up`, use `Bone Brace`
2. otherwise use `Rust Slash`
3. basic attack on fallback

### 3.2 Normal: Ashen Skull Caster

Role:

- `caster`
- backline dot pressure

Base stats:

| Stat | Value |
| --- | --- |
| HP | 102 |
| P.Atk | 4 |
| M.Atk | 20 |
| P.Def | 5 |
| M.Def | 9 |
| Speed | 10 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Ash Bolt` | `skill_attack` | `normal` | 1 | single-target magic hit |
| `Cinder Mark` | `debuff` | `advanced` | 2 | applies `burn` |

### 3.3 Normal: Grave Rat Swarm

Role:

- `assassin`
- low-HP fast backline pressure

Base stats:

| Stat | Value |
| --- | --- |
| HP | 94 |
| P.Atk | 16 |
| M.Atk | 0 |
| P.Def | 4 |
| M.Def | 4 |
| Speed | 14 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Gnaw Rush` | `skill_attack` | `normal` | 1 | attacks the lowest-HP target |

### 3.4 Elite: Warden of Seals

Role:

- `tank`
- encounter protection anchor

Base stats:

| Stat | Value |
| --- | --- |
| HP | 286 |
| P.Atk | 22 |
| M.Atk | 8 |
| P.Def | 18 |
| M.Def | 12 |
| Speed | 7 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Seal Hammer` | `skill_attack` | `normal` | 1 | heavy single hit |
| `Ward Plate` | `shield` | `advanced` | 2 | self shield |
| `Sanctum Lock` | `debuff` | `ultimate` | 3 | party-wide `speed_down` |

### 3.5 Elite: Tomb Hexer

Role:

- `controller`
- tempo denial specialist

Base stats:

| Stat | Value |
| --- | --- |
| HP | 234 |
| P.Atk | 6 |
| M.Atk | 26 |
| P.Def | 8 |
| M.Def | 14 |
| Speed | 12 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Hex Needle` | `skill_attack` | `normal` | 1 | single magic hit |
| `Dragging Curse` | `debuff` | `advanced` | 2 | applies `speed_down` |
| `Open Grave` | `debuff` | `ultimate` | 3 | applies party-wide `vulnerable` |

### 3.6 Boss: Morthis, Chapel Keeper

Role:

- `boss`
- pressure + summon + defense shred

Base stats:

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 620 | 837 | 1116 |
| P.Atk | 16 | 19 | 22 |
| M.Atk | 38 | 46 | 54 |
| P.Def | 14 | 18 | 22 |
| M.Def | 16 | 20 | 24 |
| Speed | 11 | 12 | 13 |
| Heal | 18 | 22 | 28 |

Phases:

| Phase | Trigger | Change |
| --- | --- | --- |
| P1 | start | single-target pressure and summon access |
| P2 | HP <= 50% | AoE defense shred and gravefire burst |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Chapel Flame` | `skill_attack` | `normal` | 1 | single-target magic damage |
| `Bone Procession` | `summon` | `advanced` | 2 | summons one skeleton servant |
| `Funeral Bell` | `debuff` | `advanced` | 2 | party-wide `defense_down` |
| `Ashen Mass` | `skill_attack` | `ultimate` | 3 | gravefire AoE blast |

## 4. Ancient Catacomb Wave Sheets

### easy

| Wave | Composition |
| --- | --- |
| 1 | 1 Boneguard, 1 Skull Caster |
| 2 | 1 Grave Rat Swarm, 1 Warden of Seals |
| 3 | Morthis |

### hard

| Wave | Composition |
| --- | --- |
| 1 | 2 Boneguard, 1 Skull Caster |
| 2 | 1 Grave Rat Swarm, 1 Warden of Seals, 1 Tomb Hexer |
| 3 | Morthis + summon phase |

### nightmare

| Wave | Composition |
| --- | --- |
| 1 | 2 Boneguard, 1 Skull Caster, 1 Tomb Hexer |
| 2 | 1 Warden of Seals, 1 Tomb Hexer |
| 3 | two-phase Morthis + lantern summons |

## 5. Ancient Catacomb Reward Weights

### Gear

| Difficulty | Blue | Purple | Gold | Red | Prismatic |
| --- | --- | --- | --- | --- | --- |
| easy | 72 | 26 | 2 | 0 | 0 |
| hard | 0 | 42 | 42 | 14 | 2 |
| nightmare | 0 | 0 | 22 | 56 | 22 |

### Materials

#### easy

| Material | Weight |
| --- | --- |
| Grave Dust | 36 |
| Bone Splinter | 30 |
| Burial Cloth | 22 |
| Dim Candle Wax | 12 |

#### hard

| Material | Weight |
| --- | --- |
| Grave Dust | 18 |
| Bone Splinter | 18 |
| Warden Iron Shard | 24 |
| Sealed Bone Core | 22 |
| Necro Ink | 18 |

#### nightmare

| Material | Weight |
| --- | --- |
| Sealed Bone Core | 26 |
| Necro Ink | 20 |
| Gravewake Ember | 22 |
| Chapel Sigil Fragment | 18 |
| Dusk Reliquary Fragment | 14 |

## 6. Thorned Hollow Monster Sheets

### 6.1 Normal: Rootfang Stalker

Role:

- `assassin`
- fast single-target pressure

Base stats:

| Stat | Value |
| --- | --- |
| HP | 214 |
| P.Atk | 34 |
| M.Atk | 0 |
| P.Def | 10 |
| M.Def | 8 |
| Speed | 16 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Root Lunge` | `skill_attack` | `normal` | 1 | single-target dash hit |
| `Bleed Thorn` | `debuff` | `advanced` | 2 | applies damage-over-time |

### 6.2 Normal: Poison Bark Hound

Role:

- `bruiser`
- front poison pressure

Base stats:

| Stat | Value |
| --- | --- |
| HP | 238 |
| P.Atk | 28 |
| M.Atk | 8 |
| P.Def | 12 |
| M.Def | 10 |
| Speed | 11 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Fang Rend` | `skill_attack` | `normal` | 1 | single physical hit |
| `Venom Breath` | `debuff` | `advanced` | 2 | party-wide `poison` |

### 6.3 Normal: Hollow Thornling

Role:

- `controller`
- speed denial

Base stats:

| Stat | Value |
| --- | --- |
| HP | 188 |
| P.Atk | 8 |
| M.Atk | 26 |
| P.Def | 8 |
| M.Def | 12 |
| Speed | 13 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Seed Dart` | `skill_attack` | `normal` | 1 | single magic hit |
| `Entangle Pollen` | `debuff` | `advanced` | 2 | party-wide `speed_down` |

### 6.4 Elite: Blightbranch Guardian

Role:

- `tank`
- altar guardian

Base stats:

| Stat | Value |
| --- | --- |
| HP | 486 |
| P.Atk | 32 |
| M.Atk | 12 |
| P.Def | 20 |
| M.Def | 16 |
| Speed | 8 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Bark Crush` | `skill_attack` | `normal` | 1 | heavy single hit |
| `Thorn Shell` | `shield` | `advanced` | 2 | self shield |
| `Wild Bellow` | `ultimate` | `ultimate` | 3 | party-wide `attack_down` |

### 6.5 Elite: Vine Hexcaller

Role:

- `controller`
- poison spread core

Base stats:

| Stat | Value |
| --- | --- |
| HP | 392 |
| P.Atk | 10 |
| M.Atk | 36 |
| P.Def | 10 |
| M.Def | 18 |
| Speed | 14 |
| Heal | 12 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Spore Lance` | `skill_attack` | `normal` | 1 | single magic hit |
| `Venom Bloom` | `debuff` | `advanced` | 2 | party-wide `poison` |
| `Root Ceremony` | `ultimate` | `ultimate` | 3 | burst versus poisoned targets |

### 6.6 Boss: Eldra, Heart of the Hollow

Role:

- `boss`
- poison, summon, sustain core

Base stats:

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 940 | 1269 | 1692 |
| P.Atk | 20 | 24 | 28 |
| M.Atk | 52 | 62 | 74 |
| P.Def | 18 | 22 | 26 |
| M.Def | 20 | 24 | 28 |
| Speed | 12 | 13 | 14 |
| Heal | 28 | 34 | 42 |

Phases:

| Phase | Trigger | Change |
| --- | --- | --- |
| P1 | start | poison spread + single-target pressure |
| P2 | HP <= 55% | summon seedlings + team slow |
| P3 | nightmare and HP <= 25% | self-heal + burst pressure |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Thorn Pulse` | `skill_attack` | `normal` | 1 | single magic hit |
| `Rot Bloom` | `debuff` | `advanced` | 2 | party-wide `poison` |
| `Seed of Return` | `summon` | `advanced` | 2 | summons thorn seedlings |
| `Heartsap Renewal` | `heal` | `ultimate` | 3 | self heal |
| `Hollow Cataclysm` | `skill_attack` | `ultimate` | 3 | high AoE hit |

## 7. Thorned Hollow Wave Sheets

### easy

| Wave | Composition |
| --- | --- |
| 1 | 1 Rootfang Stalker, 1 Poison Bark Hound |
| 2 | 1 Hollow Thornling, 1 Blightbranch Guardian |
| 3 | Eldra |

### hard

| Wave | Composition |
| --- | --- |
| 1 | 1 Rootfang Stalker, 1 Poison Bark Hound, 1 Hollow Thornling |
| 2 | 1 Blightbranch Guardian, 1 Vine Hexcaller |
| 3 | Eldra + seedling summons |

### nightmare

| Wave | Composition |
| --- | --- |
| 1 | 2 Rootfang Stalker, 1 Hollow Thornling, 1 Poison Bark Hound |
| 2 | 1 Blightbranch Guardian, 1 Vine Hexcaller |
| 3 | three-phase Eldra + repeated summons |

## 8. Thorned Hollow Reward Weights

### Gear

| Difficulty | Blue | Purple | Gold | Red | Prismatic |
| --- | --- | --- | --- | --- | --- |
| easy | 60 | 34 | 6 | 0 | 0 |
| hard | 0 | 32 | 46 | 18 | 4 |
| nightmare | 0 | 0 | 20 | 52 | 28 |

### Materials

#### easy

| Material | Weight |
| --- | --- |
| Thorn Resin | 34 |
| Hollow Bark | 28 |
| Venom Sap | 24 |
| Root Fiber | 14 |

#### hard

| Material | Weight |
| --- | --- |
| Thorn Resin | 16 |
| Venom Sap | 20 |
| Briar Heartwood | 24 |
| Corrupted Bloom | 22 |
| Hollow Seed Resin | 18 |

#### nightmare

| Material | Weight |
| --- | --- |
| Briar Heartwood | 24 |
| Corrupted Bloom | 20 |
| Briarbound Ember | 22 |
| Hollow Core Pod | 18 |
| Eldra Sap Crystal | 16 |

## 9. Reusable AI Templates

Recommended first-pass AI templates:

| AI Template | Use | Behavior |
| --- | --- | --- |
| `ai_bruiser_basic` | normal frontline monster | single-target pressure, defensive self-buff if needed |
| `ai_assassin_lowest_hp` | fast assassin | attacks lowest-HP target |
| `ai_caster_dot` | DoT caster | apply status first, then damage |
| `ai_controller_opening` | control specialist | opens with slow or vulnerability |
| `ai_tank_shield_first` | elite guardian | opens with shield, then pressure |
| `ai_boss_phase_scripted` | boss | follows phase scripts first |

## 10. Suggested Implementation Sequence

1. implement `Ancient Catacomb easy`
2. expand to `Ancient Catacomb hard/nightmare`
3. copy the framework into `Thorned Hollow`
4. expand to `T3-T5`
