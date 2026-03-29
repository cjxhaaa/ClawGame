# First Batch Dungeon Balance Sheets

## 1. Goal

This document provides ready-to-implement balance tables for the current two dungeons, addressing:

- rooms upgraded from single-enemy to multi-enemy compositions
- clear normal / elite / boss role assignments
- complete easy / hard / nightmare 6-room configurations per dungeon
- boss only in room 6
- nightmare requires better gear and more complete builds

Current coverage:

- `ancient_catacomb_v1`
- `sandworm_den_v1`

## 2. Global Difficulty Multipliers

| Difficulty | HP | Damage | Defense | Speed | Elite Density |
| --- | --- | --- | --- | --- | --- |
| easy | 1.00x | 1.00x | 1.00x | 1.00x | low |
| hard | 1.28x | 1.22x | 1.15x | 1.08x | medium |
| nightmare | 1.62x | 1.50x | 1.35x | 1.16x | high |

Nightmare widens the gap through three mechanisms:

- higher elite density
- stronger control and debuff synergies
- boss phases trigger earlier and more frequently with high-pressure skills

## 3. Ancient Catacomb

Identity: novice to early-mid dungeon. Theme: undead formation + control debuffs + ritual boss.

## 3.1 Monster Pool

### Normal

1. `Catacomb Boneguard` (frontline fighter)
2. `Ashen Skull Caster` (backline caster)
3. `Grave Rat Swarm` (high-speed burst)

### Elite

1. `Warden of Seals` (shield tank)
2. `Tomb Hexer` (slow and vulnerable controller)

### Boss

1. `Morthis, Chapel Keeper`

- Phase 1: summons skeleton servants + single-target pressure
- Phase 2: party-wide defense reduction + gravefire burst

## 3.2 Monster Stats

### Normal: Catacomb Boneguard

Role: `bruiser`, frontline sustained damage

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
| `Rust Slash` | `skill_attack` | normal | 1 | single-target physical hit |
| `Bone Brace` | `buff` | advanced | 2 | self `defense_up` for 2 rounds |

### Normal: Ashen Skull Caster

Role: `caster`, backline DoT

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
| `Ash Bolt` | `skill_attack` | normal | 1 | single magic hit |
| `Cinder Mark` | `debuff` | advanced | 2 | applies `burn` |

### Normal: Grave Rat Swarm

Role: `assassin`, high-speed backline pressure

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
| `Gnaw Rush` | `skill_attack` | normal | 1 | targets lowest-HP enemy |

### Elite: Warden of Seals

Role: `tank`, encounter protection anchor

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
| `Seal Hammer` | `skill_attack` | normal | 1 | heavy single hit |
| `Ward Plate` | `shield` | advanced | 2 | self shield |
| `Sanctum Lock` | `debuff` | ultimate | 3 | party-wide `speed_down` |

### Elite: Tomb Hexer

Role: `controller`, tempo denial

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
| `Hex Needle` | `skill_attack` | normal | 1 | single magic hit |
| `Dragging Curse` | `debuff` | advanced | 2 | applies `speed_down` |
| `Open Grave` | `debuff` | ultimate | 3 | party-wide `vulnerable` |

### Boss: Morthis, Chapel Keeper

Role: `boss`, pressure + summon + defense shred

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

| Phase | Trigger | Mechanic |
| --- | --- | --- |
| P1 | start | single-target pressure + summon access |
| P2 | HP <= 50% | party-wide `defense_down` + gravefire burst |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Chapel Flame` | `skill_attack` | normal | 1 | single-target magic damage |
| `Bone Procession` | `summon` | advanced | 2 | summons one skeleton servant |
| `Funeral Bell` | `debuff` | advanced | 2 | party-wide `defense_down` |
| `Ashen Mass` | `skill_attack` | ultimate | 3 | gravefire AoE blast |

## 3.3 Room Composition Tables

### easy

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 1 Boneguard + 1 Skull Caster | basic warm-up |
| 2 | 1 Boneguard + 1 Rat Swarm + 1 Skull Caster | first front/back pressure |
| 3 | 1 Boneguard + 1 Warden of Seals | first elite check |
| 4 | 1 Rat Swarm + 1 Skull Caster + 1 Tomb Hexer | control + burst combo |
| 5 | 1 Warden + 1 Tomb Hexer + 1 Boneguard | resource pressure pre-boss |
| 6 | Morthis (Boss) + 1 Boneguard | finale (boss-led) |

### hard

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Boneguard + 1 Skull Caster | thicker frontline |
| 2 | 1 Boneguard + 2 Rat Swarm + 1 Skull Caster | double burst pressure |
| 3 | 1 Warden + 1 Boneguard + 1 Skull Caster | elite guard formation |
| 4 | 1 Tomb Hexer + 1 Rat Swarm + 1 Skull Caster + 1 Boneguard | control into burst |
| 5 | 1 Warden + 1 Tomb Hexer + 1 Boneguard + 1 Rat Swarm | high-pressure mixed squad |
| 6 | Morthis (Boss) + 1 Warden + 1 Skull Caster | boss + protective adds |

### nightmare

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Boneguard + 1 Skull Caster + 1 Rat Swarm | immediate high pressure |
| 2 | 1 Warden + 1 Rat Swarm + 1 Skull Caster + 1 Boneguard | early elite in room 2 |
| 3 | 1 Warden + 1 Tomb Hexer + 1 Boneguard + 1 Rat Swarm | double-elite core begins |
| 4 | 1 Tomb Hexer + 2 Rat Swarm + 1 Skull Caster + 1 Boneguard | control + harvest chain |
| 5 | 1 Warden + 1 Tomb Hexer + 1 Skull Caster + 1 Boneguard + 1 Rat Swarm | five-monster full resource drain |
| 6 | Morthis (Boss) + 1 Warden + 1 Tomb Hexer (P2 summons 1 Rat) | finale, requires quality gear |

Notes:

- boss only appears in room 6
- nightmare is expected to require good gear + reasonable potions for a reliable clear

---

## 4. Sandworm Den

Identity: mid-to-high tier dungeon. Theme: armor-break + sustained damage + burst boss.

## 4.1 Monster Pool

### Normal

1. `Dune Skitterer` (high-speed harasser)
2. `Sand Burrower` (armor-break melee)
3. `Scorched Spitter` (backline burn)

### Elite

1. `Carapace Crusher` (heavy hit and shield-break)
2. `Venom Herald` (poison spread and vulnerable)

### Boss

1. `Kharzug, Dunescourge Matriarch`

- Phase 1: heavy armor pressure + targeted charge
- Phase 2: party-wide sandstorm + larva summons

## 4.2 Monster Stats

### Normal: Dune Skitterer

Role: `assassin`, high-speed harassment

| Stat | Value |
| --- | --- |
| HP | 108 |
| P.Atk | 22 |
| M.Atk | 0 |
| P.Def | 6 |
| M.Def | 6 |
| Speed | 16 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Sand Dash` | `skill_attack` | normal | 1 | targets lowest-HP enemy |

### Normal: Sand Burrower

Role: `bruiser`, armor-break melee

| Stat | Value |
| --- | --- |
| HP | 140 |
| P.Atk | 24 |
| M.Atk | 0 |
| P.Def | 12 |
| M.Def | 8 |
| Speed | 9 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Burrow Strike` | `skill_attack` | normal | 1 | single heavy hit |
| `Carapace Tear` | `debuff` | advanced | 2 | applies `defense_down` |

### Normal: Scorched Spitter

Role: `caster`, backline burn

| Stat | Value |
| --- | --- |
| HP | 112 |
| P.Atk | 4 |
| M.Atk | 22 |
| P.Def | 6 |
| M.Def | 10 |
| Speed | 12 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Acid Spit` | `skill_attack` | normal | 1 | single magic hit |
| `Scorch Mark` | `debuff` | advanced | 2 | applies `burn` |

### Elite: Carapace Crusher

Role: `bruiser`, heavy hit and shield-break

| Stat | Value |
| --- | --- |
| HP | 310 |
| P.Atk | 36 |
| M.Atk | 8 |
| P.Def | 20 |
| M.Def | 12 |
| Speed | 8 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Crush Slam` | `skill_attack` | normal | 1 | heavy single physical hit |
| `Shell Shatter` | `debuff` | advanced | 2 | applies `defense_down` + removes `shielded` |
| `Dune Overrun` | `skill_attack` | ultimate | 3 | AoE physical hit |

### Elite: Venom Herald

Role: `controller`, poison spread and vulnerable

| Stat | Value |
| --- | --- |
| HP | 262 |
| P.Atk | 8 |
| M.Atk | 28 |
| P.Def | 10 |
| M.Def | 16 |
| Speed | 13 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Venom Dart` | `skill_attack` | normal | 1 | single magic hit with `poison` |
| `Blight Mist` | `debuff` | advanced | 2 | party-wide `poison` |
| `Expose Wound` | `debuff` | ultimate | 3 | target `vulnerable` + `poison` on already-poisoned targets |

### Boss: Kharzug, Dunescourge Matriarch

Role: `boss`, armor crush + summon + AoE

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 720 | 972 | 1296 |
| P.Atk | 32 | 39 | 46 |
| M.Atk | 18 | 22 | 26 |
| P.Def | 20 | 26 | 32 |
| M.Def | 14 | 18 | 22 |
| Speed | 10 | 11 | 12 |
| Heal | 0 | 0 | 0 |

Phases:

| Phase | Trigger | Mechanic |
| --- | --- | --- |
| P1 | start | heavy armor pressure + targeted charge |
| P2 | HP <= 50% | party-wide `defense_down` sandstorm + larva summons |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Dune Crash` | `skill_attack` | normal | 1 | heavy single physical hit |
| `Burrow Charge` | `skill_attack` | advanced | 2 | targets highest-threat enemy with `defense_down` |
| `Larva Storm` | `summon` | advanced | 2 | summons two larva adds |
| `Sandscour Tempest` | `skill_attack` | ultimate | 3 | party-wide physical + `defense_down` |

## 4.3 Room Composition Tables

### easy

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 1 Sand Burrower + 1 Scorched Spitter | basic armor-break + burn |
| 2 | 1 Dune Skitterer + 1 Sand Burrower + 1 Spitter | front/back combined pressure |
| 3 | 1 Carapace Crusher + 1 Sand Burrower | first elite heavy hit check |
| 4 | 1 Venom Herald + 1 Spitter + 1 Skitterer | poison synergy intro |
| 5 | 1 Crusher + 1 Herald + 1 Burrower | high-pressure transition |
| 6 | Kharzug (Boss) + 1 Sand Burrower | boss room |

### hard

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Sand Burrower + 1 Spitter | heavier frontline |
| 2 | 1 Skitterer + 2 Burrower + 1 Spitter | multiple melee pressure |
| 3 | 1 Crusher + 1 Burrower + 1 Spitter | elite-led push |
| 4 | 1 Herald + 1 Crusher + 1 Skitterer + 1 Spitter | double-elite synergy starts |
| 5 | 1 Herald + 1 Crusher + 1 Burrower + 1 Skitterer | high-pressure resource room |
| 6 | Kharzug (Boss) + 1 Crusher + 1 Herald | boss + double-elite support |

### nightmare

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Burrower + 1 Spitter + 1 Skitterer | opening four-enemy pressure |
| 2 | 1 Crusher + 1 Burrower + 1 Spitter + 1 Skitterer | elite appears in room 2 |
| 3 | 1 Crusher + 1 Herald + 1 Burrower + 1 Spitter | double-elite formation |
| 4 | 1 Herald + 1 Crusher + 2 Skitterer + 1 Spitter | five-enemy sustained pressure |
| 5 | 1 Herald + 1 Crusher + 1 Burrower + 1 Spitter + 1 Skitterer | extreme resource check |
| 6 | Kharzug (Boss) + 1 Crusher + 1 Herald (P2 summons 2 larva) | high gear-requirement finale |

Notes:

- boss only appears in room 6
- in nightmare, rooms 4-6 can cause consecutive failures without good gear and defensive build

---

## 5. Nightmare Clear Threshold Guidance

To reflect "nightmare is meaningfully harder", the following tuning targets are recommended:

- panel combat power should be at least `1.18x` the hard recommended value for the same dungeon
- primary defense stats must not fall below same-level median
- bring at least 2 potion types (for example HP + DEF or HP + ATK)
- party must include at least 1 reliable sustain role (high defense, healer, or shield)

## 6. Validation Checklist

1. every dungeon has three-tier complete 6-room wave configs
2. boss only appears in room 6 for every dungeon and difficulty
3. hard adds at least 1 more elite pressure point compared to easy
4. nightmare includes at least 1 double-elite or 5-monster full squad encounter compared to hard
5. observer logs can identify each room's "composition theme" and boss phase switches
