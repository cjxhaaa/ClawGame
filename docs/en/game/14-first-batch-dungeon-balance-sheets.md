# First Batch Dungeon Balance Sheets

## 1. Goal

This document provides ready-to-implement balance tables for the first batch of dungeons in the four-dungeon seasonal model, addressing:

- rooms upgraded from single-enemy to multi-enemy compositions
- clear normal / elite / boss role assignments
- complete easy / hard / nightmare 6-room configurations per dungeon
- boss only in room 6
- nightmare requires better gear and more complete builds

Season-one target coverage:

- `ancient_catacomb_v1`
- `thorned_hollow_v1`
- `sunscar_warvault_v1`
- `obsidian_spire_v1`

Current detailed balance-sheet coverage:

- `ancient_catacomb_v1`
- `sandworm_den_v1`

Current note:

- `sandworm_den_v1` remains a useful archived high-pressure sample for encounter design, but it is no longer part of the new four-dungeon seasonal loot plan
- `thorned_hollow_v1`, `sunscar_warvault_v1`, and `obsidian_spire_v1` still need full room-by-room sheets in a follow-up pass

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

### Seasonal pacing targets

The current seasonal difficulty pacing should roughly align with:

- easy: first reliable clear around day `5`
- hard: first reliable clear around day `10`
- nightmare: first reliable clear around day `15`

Based on the season progression framework, this approximately maps to:

- day `5`: level `20-25`
- day `10`: level `42-50`
- day `15`: level `68-78`

Tuning assumptions:

- easy should be beatable with mixed baseline gear, basic potion usage, and incomplete set progress
- hard should expect a coherent build, stronger panel power, and intentional potion preparation
- nightmare should be the first long-term farming tier where bots begin sustained set-piece and affix chasing
- `ancient_catacomb_v1`, `thorned_hollow_v1`, `sunscar_warvault_v1`, and `obsidian_spire_v1` should follow this target curve
- archived `sandworm_den_v1` may sit slightly above the active seasonal curve

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

## 4. Sandworm Den (Archived Reference)

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

## 5. Thorned Hollow

Identity: precision and crit dungeon. Theme: thorn packs, hunter pressure, marked-target execution.

Set direction:

- `Briarbound Sight`
- accuracy
- crit
- focus-fire execution

## 5.1 Monster Pool

### Normal

1. `Rootlash Hunter` (fast skirmisher)
2. `Briar Archer` (backline precision DPS)
3. `Thorn Prowler` (marking flanker)

### Elite

1. `Vine Snarekeeper` (bind and accuracy pressure)
2. `Hollow Fang Alpha` (crit-finisher bruiser)

### Boss

1. `Eldra, Thornseer Matron`

- Phase 1: marks priority targets + ranged pressure
- Phase 2: mass bind field + crit execution windows

## 5.2 Monster Stats

### Normal: Rootlash Hunter

Role: `assassin`, fast pressure and target setup

| Stat | Value |
| --- | --- |
| HP | 112 |
| P.Atk | 20 |
| M.Atk | 0 |
| P.Def | 6 |
| M.Def | 6 |
| Speed | 15 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Rootstep Slash` | `skill_attack` | normal | 1 | fast single-target physical hit |
| `Needle Feint` | `debuff` | advanced | 2 | applies short `evasion_down` or `marked` |

### Normal: Briar Archer

Role: `caster`, backline precision burst

| Stat | Value |
| --- | --- |
| HP | 104 |
| P.Atk | 22 |
| M.Atk | 6 |
| P.Def | 5 |
| M.Def | 7 |
| Speed | 13 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Thornshot` | `skill_attack` | normal | 1 | single-target physical shot |
| `Hunter's Mark` | `debuff` | advanced | 2 | applies `marked` and light `vulnerable` |

### Normal: Thorn Prowler

Role: `skirmisher`, flank and focus-fire support

| Stat | Value |
| --- | --- |
| HP | 118 |
| P.Atk | 18 |
| M.Atk | 8 |
| P.Def | 7 |
| M.Def | 8 |
| Speed | 14 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Prowl Pounce` | `skill_attack` | normal | 1 | dives lowest-HP backline target |
| `Exposed Trail` | `debuff` | advanced | 2 | extends marked-target pressure |

### Elite: Vine Snarekeeper

Role: `controller`, bind and tempo denial

| Stat | Value |
| --- | --- |
| HP | 248 |
| P.Atk | 10 |
| M.Atk | 26 |
| P.Def | 10 |
| M.Def | 14 |
| Speed | 11 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Bramble Lash` | `skill_attack` | normal | 1 | single-target magic hit |
| `Snare Bloom` | `debuff` | advanced | 2 | applies `bind` or `speed_down` |
| `Hunter's Lattice` | `debuff` | ultimate | 3 | party-wide bind pressure window |

### Elite: Hollow Fang Alpha

Role: `bruiser`, crit-finisher and execution threat

| Stat | Value |
| --- | --- |
| HP | 274 |
| P.Atk | 32 |
| M.Atk | 0 |
| P.Def | 14 |
| M.Def | 9 |
| Speed | 14 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Rending Bite` | `skill_attack` | normal | 1 | heavy physical single-target hit |
| `Pack Hunger` | `buff` | advanced | 2 | self crit-rate and speed boost |
| `Cull the Weak` | `skill_attack` | ultimate | 3 | bonus damage to marked or low-HP targets |

### Boss: Eldra, Thornseer Matron

Role: `boss`, mark setup + bind field + crit execution

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 640 | 865 | 1152 |
| P.Atk | 18 | 22 | 26 |
| M.Atk | 34 | 41 | 49 |
| P.Def | 13 | 16 | 20 |
| M.Def | 18 | 22 | 26 |
| Speed | 12 | 13 | 14 |
| Heal | 0 | 0 | 0 |

Phases:

| Phase | Trigger | Mechanic |
| --- | --- | --- |
| P1 | start | mark priority targets + ranged pressure |
| P2 | HP <= 50% | bind field + critical execution turns |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Seeker Thorn` | `skill_attack` | normal | 1 | single-target precision hit |
| `Matron's Brand` | `debuff` | advanced | 2 | applies `marked` and light `vulnerable` |
| `Briar Dominion` | `debuff` | advanced | 2 | party-wide bind pressure |
| `Widow Bloom` | `skill_attack` | ultimate | 3 | heavy execution hit, stronger against marked targets |

## 5.3 Room Composition Tables

### easy

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 1 Rootlash Hunter + 1 Briar Archer | warm-up precision pressure |
| 2 | 1 Rootlash Hunter + 1 Thorn Prowler + 1 Briar Archer | first mark-and-focus test |
| 3 | 1 Vine Snarekeeper + 1 Briar Archer | first elite control check |
| 4 | 1 Thorn Prowler + 1 Briar Archer + 1 Vine Snarekeeper | bind + backline focus |
| 5 | 1 Hollow Fang Alpha + 1 Rootlash Hunter + 1 Briar Archer | crit-finisher pressure |
| 6 | Eldra (Boss) + 1 Thorn Prowler | finale with marked-target threat |

### hard

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Rootlash Hunter + 1 Briar Archer | fast opener |
| 2 | 1 Rootlash Hunter + 1 Thorn Prowler + 2 Briar Archer | stacked backline pressure |
| 3 | 1 Vine Snarekeeper + 1 Rootlash Hunter + 1 Briar Archer | elite control anchor |
| 4 | 1 Hollow Fang Alpha + 1 Thorn Prowler + 1 Briar Archer + 1 Rootlash Hunter | crit burst follow-up |
| 5 | 1 Vine Snarekeeper + 1 Hollow Fang Alpha + 1 Briar Archer + 1 Thorn Prowler | control into execution |
| 6 | Eldra (Boss) + 1 Vine Snarekeeper + 1 Briar Archer | boss + bind support |

### nightmare

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Rootlash Hunter + 1 Briar Archer + 1 Thorn Prowler | early speed check |
| 2 | 1 Vine Snarekeeper + 1 Rootlash Hunter + 1 Thorn Prowler + 1 Briar Archer | elite from room 2 |
| 3 | 1 Vine Snarekeeper + 1 Hollow Fang Alpha + 1 Briar Archer + 1 Thorn Prowler | double-elite precision core |
| 4 | 2 Briar Archer + 1 Thorn Prowler + 1 Rootlash Hunter + 1 Vine Snarekeeper | ranged execution pressure |
| 5 | 1 Hollow Fang Alpha + 1 Vine Snarekeeper + 1 Briar Archer + 1 Thorn Prowler + 1 Rootlash Hunter | full focus-fire drain |
| 6 | Eldra (Boss) + 1 Hollow Fang Alpha + 1 Vine Snarekeeper (P2 adds 1 Briar Archer) | high-risk bind and crit finale |

Notes:

- this dungeon should feel fastest among the four seasonal farms
- nightmare should punish slow or low-accuracy teams heavily

---

## 6. Sunscar Warvault

Identity: physical burst dungeon. Theme: military constructs, armor-break lanes, execution windows.

Set direction:

- `Sunscar Assault`
- physical attack
- burst
- execution

## 6.1 Monster Pool

### Normal

1. `Vault Legionnaire` (frontline bruiser)
2. `Ember Rifleman` (backline burst shot)
3. `Siege Hound` (fast armor-break harasser)

### Elite

1. `Breach Captain` (frontline breaker)
2. `Warvault Crusher` (heavy burst elite)

### Boss

1. `General Rhazek, Ash Commander`

- Phase 1: formation pressure + armor break
- Phase 2: kill order check + burst volleys

## 6.2 Monster Stats

### Normal: Vault Legionnaire

Role: `bruiser`, frontline physical pressure

| Stat | Value |
| --- | --- |
| HP | 146 |
| P.Atk | 24 |
| M.Atk | 4 |
| P.Def | 12 |
| M.Def | 7 |
| Speed | 9 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Vault Cleave` | `skill_attack` | normal | 1 | single heavy physical hit |
| `Shield Bash` | `debuff` | advanced | 2 | applies short `defense_down` |

### Normal: Ember Rifleman

Role: `caster`, backline burst shot

| Stat | Value |
| --- | --- |
| HP | 108 |
| P.Atk | 26 |
| M.Atk | 6 |
| P.Def | 5 |
| M.Def | 7 |
| Speed | 12 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Ember Round` | `skill_attack` | normal | 1 | backline burst shot |
| `Tracer Volley` | `skill_attack` | advanced | 2 | stronger hit against `defense_down` targets |

### Normal: Siege Hound

Role: `assassin`, armor-break harasser

| Stat | Value |
| --- | --- |
| HP | 118 |
| P.Atk | 20 |
| M.Atk | 0 |
| P.Def | 7 |
| M.Def | 5 |
| Speed | 15 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Rivet Fang` | `skill_attack` | normal | 1 | targets lowest-defense enemy |
| `Armor Gnaw` | `debuff` | advanced | 2 | applies `defense_down` |

### Elite: Breach Captain

Role: `tank`, formation breaker

| Stat | Value |
| --- | --- |
| HP | 296 |
| P.Atk | 34 |
| M.Atk | 8 |
| P.Def | 18 |
| M.Def | 11 |
| Speed | 10 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Breach Swing` | `skill_attack` | normal | 1 | high single-target damage |
| `Linebreaker Order` | `debuff` | advanced | 2 | applies party-side `defense_down` pressure |
| `Open the Wall` | `skill_attack` | ultimate | 3 | bonus damage to already-broken targets |

### Elite: Warvault Crusher

Role: `bruiser`, high-burst execution elite

| Stat | Value |
| --- | --- |
| HP | 322 |
| P.Atk | 38 |
| M.Atk | 6 |
| P.Def | 16 |
| M.Def | 10 |
| Speed | 8 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Hammerfall` | `skill_attack` | normal | 1 | heavy single-target hit |
| `Crush Momentum` | `buff` | advanced | 2 | self attack-up after a hit lands |
| `Siege Burst` | `skill_attack` | ultimate | 3 | AoE physical burst |

### Boss: General Rhazek, Ash Commander

Role: `boss`, armor break + formation pressure + burst window checks

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 700 | 946 | 1260 |
| P.Atk | 34 | 41 | 51 |
| M.Atk | 10 | 12 | 15 |
| P.Def | 18 | 22 | 27 |
| M.Def | 12 | 15 | 18 |
| Speed | 11 | 12 | 13 |
| Heal | 0 | 0 | 0 |

Phases:

| Phase | Trigger | Mechanic |
| --- | --- | --- |
| P1 | start | formation pressure + armor break |
| P2 | HP <= 50% | execution volley and add pressure |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Command Slash` | `skill_attack` | normal | 1 | heavy physical strike |
| `Ashen Breach` | `debuff` | advanced | 2 | applies strong `defense_down` |
| `Volley Order` | `skill_attack` | advanced | 2 | party-side burst shot pattern |
| `Last Barrage` | `skill_attack` | ultimate | 3 | high-damage multi-target finisher |

## 6.3 Room Composition Tables

### easy

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 1 Vault Legionnaire + 1 Ember Rifleman | basic frontline/backline burst |
| 2 | 1 Siege Hound + 1 Vault Legionnaire + 1 Ember Rifleman | first armor-break combo |
| 3 | 1 Breach Captain + 1 Vault Legionnaire | first elite breakpoint |
| 4 | 1 Siege Hound + 1 Ember Rifleman + 1 Breach Captain | break then burst |
| 5 | 1 Warvault Crusher + 1 Vault Legionnaire + 1 Ember Rifleman | physical pressure spike |
| 6 | Rhazek (Boss) + 1 Siege Hound | finale with break support |

### hard

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Vault Legionnaire + 1 Ember Rifleman | heavy front pressure |
| 2 | 1 Siege Hound + 2 Vault Legionnaire + 1 Ember Rifleman | stacked melee line |
| 3 | 1 Breach Captain + 1 Siege Hound + 1 Ember Rifleman | elite-led burst check |
| 4 | 1 Warvault Crusher + 1 Breach Captain + 1 Ember Rifleman + 1 Siege Hound | double-elite kill window |
| 5 | 1 Warvault Crusher + 1 Breach Captain + 1 Vault Legionnaire + 1 Ember Rifleman | peak physical pressure |
| 6 | Rhazek (Boss) + 1 Warvault Crusher + 1 Breach Captain | boss + double breaker |

### nightmare

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Vault Legionnaire + 1 Ember Rifleman + 1 Siege Hound | immediate breakpoint pressure |
| 2 | 1 Breach Captain + 1 Siege Hound + 1 Vault Legionnaire + 1 Ember Rifleman | elite appears early |
| 3 | 1 Warvault Crusher + 1 Breach Captain + 1 Ember Rifleman + 1 Siege Hound | double-elite burst core |
| 4 | 1 Warvault Crusher + 1 Breach Captain + 2 Ember Rifleman + 1 Siege Hound | ranged burst stack |
| 5 | 1 Warvault Crusher + 1 Breach Captain + 1 Vault Legionnaire + 1 Ember Rifleman + 1 Siege Hound | full execution drain |
| 6 | Rhazek (Boss) + 1 Warvault Crusher + 1 Breach Captain (P2 adds 2 Siege Hounds) | lethal burst finale |

Notes:

- this dungeon should be the cleanest physical damage benchmark
- nightmare should punish weak frontline protection and slow kill speed

---

## 7. Obsidian Spire

Identity: spell-chain dungeon. Theme: cursed mirrors, void casters, tempo denial.

Set direction:

- `Nightglass Arcanum`
- magic attack
- holy power
- cast chaining

## 7.1 Monster Pool

### Normal

1. `Spire Channeler` (backline caster)
2. `Mirror Wisp` (tempo disruptor)
3. `Obsidian Thrall` (frontline spellguard)

### Elite

1. `Void Preceptor` (cooldown and silence pressure)
2. `Nightglass Engine` (AoE spell anchor)

### Boss

1. `Serathiel, Prism of Ash`

- Phase 1: mirrored spell volleys + debuff setup
- Phase 2: chained arcane blasts + holy-burn pressure

## 7.2 Monster Stats

### Normal: Spire Channeler

Role: `caster`, steady backline spell pressure

| Stat | Value |
| --- | --- |
| HP | 102 |
| P.Atk | 4 |
| M.Atk | 24 |
| P.Def | 5 |
| M.Def | 10 |
| Speed | 12 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Glass Ray` | `skill_attack` | normal | 1 | single-target magic hit |
| `Echo Pulse` | `skill_attack` | advanced | 2 | bonus damage if target is debuffed |

### Normal: Mirror Wisp

Role: `controller`, turn-order disruption

| Stat | Value |
| --- | --- |
| HP | 92 |
| P.Atk | 0 |
| M.Atk | 20 |
| P.Def | 4 |
| M.Def | 9 |
| Speed | 16 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Prism Flicker` | `skill_attack` | normal | 1 | light magic hit |
| `Mirror Drag` | `debuff` | advanced | 2 | applies `speed_down` or `silence`-like pressure |

### Normal: Obsidian Thrall

Role: `tank`, anti-caster frontline guard

| Stat | Value |
| --- | --- |
| HP | 148 |
| P.Atk | 18 |
| M.Atk | 8 |
| P.Def | 12 |
| M.Def | 12 |
| Speed | 8 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Blackglass Swing` | `skill_attack` | normal | 1 | frontline strike |
| `Mirror Guard` | `buff` | advanced | 2 | self spell damage reduction |

### Elite: Void Preceptor

Role: `controller`, silence and cooldown pressure

| Stat | Value |
| --- | --- |
| HP | 244 |
| P.Atk | 8 |
| M.Atk | 30 |
| P.Def | 9 |
| M.Def | 16 |
| Speed | 12 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Void Script` | `skill_attack` | normal | 1 | single-target magic hit |
| `Mute Sigil` | `debuff` | advanced | 2 | silence or cast-delay pressure |
| `Lesson of Ash` | `debuff` | ultimate | 3 | party-side cooldown disruption window |

### Elite: Nightglass Engine

Role: `summoner`, AoE spell anchor

| Stat | Value |
| --- | --- |
| HP | 282 |
| P.Atk | 6 |
| M.Atk | 34 |
| P.Def | 14 |
| M.Def | 18 |
| Speed | 9 |
| Heal | 0 |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Arc Furnace` | `skill_attack` | normal | 1 | single-target magic attack |
| `Refraction Wave` | `skill_attack` | advanced | 2 | party-wide magic splash |
| `Overload Prism` | `skill_attack` | ultimate | 3 | large AoE spell burst |

### Boss: Serathiel, Prism of Ash

Role: `boss`, chained spell bursts + tempo denial + holy-burn pressure

| Stat | easy | hard | nightmare |
| --- | --- | --- | --- |
| HP | 660 | 891 | 1188 |
| P.Atk | 8 | 10 | 12 |
| M.Atk | 42 | 51 | 63 |
| P.Def | 12 | 15 | 18 |
| M.Def | 20 | 24 | 28 |
| Speed | 11 | 12 | 13 |
| Heal | 16 | 20 | 24 |

Phases:

| Phase | Trigger | Mechanic |
| --- | --- | --- |
| P1 | start | mirrored volleys + debuff setup |
| P2 | HP <= 50% | chained arcane blasts + holy-burn pressure |

Skills:

| Skill | Type | Tier | Cooldown | Notes |
| --- | --- | --- | --- | --- |
| `Prism Lance` | `skill_attack` | normal | 1 | heavy magic hit |
| `Mirror Sentence` | `debuff` | advanced | 2 | applies silence or `magic_vulnerable` |
| `Ash Cathedral` | `skill_attack` | advanced | 2 | party-wide holy-burn pulse |
| `Shattered Halo` | `skill_attack` | ultimate | 3 | chained burst hit across multiple targets |

## 7.3 Room Composition Tables

### easy

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 1 Obsidian Thrall + 1 Spire Channeler | baseline spell lane |
| 2 | 1 Mirror Wisp + 1 Obsidian Thrall + 1 Spire Channeler | first tempo disruption |
| 3 | 1 Void Preceptor + 1 Spire Channeler | elite cast-pressure check |
| 4 | 1 Mirror Wisp + 1 Spire Channeler + 1 Void Preceptor | silence and burst setup |
| 5 | 1 Nightglass Engine + 1 Obsidian Thrall + 1 Spire Channeler | AoE spell pressure |
| 6 | Serathiel (Boss) + 1 Mirror Wisp | finale with cast denial support |

### hard

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Obsidian Thrall + 1 Spire Channeler | heavier spellguard line |
| 2 | 1 Mirror Wisp + 1 Obsidian Thrall + 2 Spire Channeler | stacked casting pressure |
| 3 | 1 Void Preceptor + 1 Obsidian Thrall + 1 Spire Channeler | elite tempo anchor |
| 4 | 1 Nightglass Engine + 1 Mirror Wisp + 1 Spire Channeler + 1 Obsidian Thrall | AoE plus disruption |
| 5 | 1 Void Preceptor + 1 Nightglass Engine + 1 Spire Channeler + 1 Mirror Wisp | double-elite spell chain |
| 6 | Serathiel (Boss) + 1 Void Preceptor + 1 Spire Channeler | boss + control support |

### nightmare

| Room | Composition | Design Intent |
| --- | --- | --- |
| 1 | 2 Obsidian Thrall + 1 Spire Channeler + 1 Mirror Wisp | early cast disruption |
| 2 | 1 Void Preceptor + 1 Obsidian Thrall + 1 Spire Channeler + 1 Mirror Wisp | elite from room 2 |
| 3 | 1 Void Preceptor + 1 Nightglass Engine + 1 Spire Channeler + 1 Mirror Wisp | double-elite spell core |
| 4 | 2 Spire Channeler + 1 Nightglass Engine + 1 Mirror Wisp + 1 Obsidian Thrall | repeated magic burst |
| 5 | 1 Void Preceptor + 1 Nightglass Engine + 1 Spire Channeler + 1 Mirror Wisp + 1 Obsidian Thrall | full tempo-drain room |
| 6 | Serathiel (Boss) + 1 Void Preceptor + 1 Nightglass Engine (P2 adds 1 Mirror Wisp) | chained-cast finale |

Notes:

- this dungeon should be the clearest anti-caster and caster-mirror pressure dungeon
- nightmare should strongly reward potion timing and skill-rotation discipline

---

## 8. Nightmare Clear Threshold Guidance

To reflect "nightmare is meaningfully harder", the following tuning targets are recommended:

- panel combat power should be at least `1.18x` the hard recommended value for the same dungeon
- primary defense stats must not fall below same-level median
- bring at least 2 potion types (for example HP + DEF or HP + ATK)
- party must include at least 1 reliable sustain role (high defense, healer, or shield)

## 9. Validation Checklist

1. every dungeon has three-tier complete 6-room wave configs
2. boss only appears in room 6 for every dungeon and difficulty
3. hard adds at least 1 more elite pressure point compared to easy
4. nightmare includes at least 1 double-elite or 5-monster full squad encounter compared to hard
5. observer logs can identify each room's "composition theme" and boss phase switches

## 10. Follow-Up Work

The next update to this document should:

1. align every room sheet with the new recommended combat power bands from the combat-power specification
2. validate real clear-rate telemetry against the day `5` / day `10` / day `15` season pacing targets
3. decide whether `sandworm_den_v1` should remain as a permanent archive appendix or move out of the main seasonal balance sheet document
4. add a compact materials and reward-summary appendix for each of the four active seasonal dungeons
