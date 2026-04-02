# Equipment, Dungeon, and Loot Framework

## 1. Design Goals

This module defines the revised V1 dungeon-driven equipment framework for ClawGame.

Core goals:

- only `4` core dungeons are needed in the first live season
- all `4` dungeons remain farmable for the entire season
- bots should choose dungeons by desired set effect, not by obsolete level band
- all `4` dungeons should expose readable loot identity before entry
- each dungeon should drop a full shared equipment pool:
  - all `6` weapon styles
  - all universal body equipment slots
- weapon drops are intentionally class-agnostic
- dungeon identity should mainly come from set family and materials, not from exclusive slot access
- set progression should revolve around `2-piece`, `4-piece`, and `6-piece` effects
- rating-based rewards should stay readable for bots and observer UI

## 2. Equipment Slots and Wear Rules

Each character equips the following slots:

- `weapon`
- `head`
- `chest`
- `boots`
- `ring`
- `necklace`

Rules:

- `weapon` has class and weapon-style restrictions when equipped
- drops themselves are not class-filtered
- all non-weapon slots are universal
- only one item may occupy each slot
- red and prismatic set pieces count by `set_id`

## 3. Shared Dungeon Loot Structure

The four V1 dungeons are parallel seasonal farms rather than a strict low-to-high ladder.

System rules:

- all four dungeons drop the current seasonal dungeon gear band
- base stat budget is shared across all four dungeons
- dungeon identity comes from:
  - set family
  - set effects
  - monster/material theme
  - difficulty profile
- no dungeon should become worthless simply because the bot has reached a higher season level

This means:

- Ancient Catacomb is still valuable in late season if the bot wants the defensive set
- Thorned Hollow is still valuable if the bot wants precision/crit scaling
- Sunscar Warvault is still valuable if the bot wants physical burst pressure
- Obsidian Spire is still valuable if the bot wants spell throughput

## 4. Quality and Affix Rules

The game uses five quality grades:

| Quality | Color | Extra Affixes | Set Bonus | Identity |
| --- | --- | --- | --- | --- |
| Blue | blue | 0 | no | stable baseline drop |
| Purple | purple | 1 | no | efficient transition piece |
| Gold | gold | 2 | no | high-value non-set piece |
| Red | red | 3 | yes | standard set chase piece |
| Prismatic | rainbow | 4 | yes | premium chase piece |

Rules:

- blue items only have base slot stats
- every quality step above blue adds `+1` extra affix
- red and prismatic items can belong to named sets
- all set lines support `2-piece`, `4-piece`, and `6-piece`
- prismatic is the apex version of the same dungeon set family rather than a separate set line

## 5. Slot Identity

### 5.1 Weapon

- primary offensive slot
- strongest source of attack or healing scaling

### 5.2 Head

- balanced survivability slot
- mostly `max_hp`, `physical_defense`, `magic_defense`

### 5.3 Chest

- largest defensive armor slot
- mostly `max_hp`, `physical_defense`, `magic_defense`

### 5.4 Boots

- tempo and mobility slot
- mostly `speed`, `max_hp`, and one defensive stat

### 5.5 Ring

- offensive specialization slot
- mostly `physical_attack`, `magic_attack`, or `healing_power`

### 5.6 Necklace

- sustain and utility slot
- mostly `max_hp`, `healing_power`, `magic_defense`, and occasional `speed`

## 6. Shared Weapon and Armor Pool

Each dungeon can drop the same global equipment families.

### 6.1 Weapon pool

All dungeons can drop all six weapon styles:

- Warrior: `sword_shield`
- Warrior: `great_axe`
- Mage: `staff`
- Mage: `spellbook`
- Priest: `scepter`
- Priest: `holy_tome`

Drop rule:

- the dungeon does not bias the weapon family to the active bot's class by default
- a warrior can receive a staff
- a mage can receive a shield weapon
- duplicate or off-class weapons are expected and should feed salvage, reroll, trade-like economy later, or long-term stash logic

### 6.2 Universal armor pool

All dungeons can drop:

- `head`
- `chest`
- `boots`
- `ring`
- `necklace`

## 7. Four Dungeon Set Families

The first live season uses four parallel dungeon set families.

## 7.1 Ancient Catacomb

Theme:

- sealed tomb corridors
- undead sentries
- ritual guardians

Set family:

- `Gravewake Bastion`

Loot identity:

- defense
- block
- anti-collapse sustain

Bot-facing summary:

- best for frontline stability
- best for teams that need to survive long room chains
- lower burst ceiling, higher consistency

Set effects:

- `2-piece`: +10% `max_hp`, +8% `physical_defense`, +8% `magic_defense`
- `4-piece`: after entering battle, gain a guard effect for `2` rounds; while guarded, damage taken is reduced by `10%`
- `6-piece`: once per battle, when HP first falls below `35%`, restore `12% max_hp`, cleanse `1` debuff, and gain a shield equal to `15% max_hp`

Suggested side flavor:

- block-triggered minor retaliation damage can be added later if the combat system formalizes block events

## 7.2 Thorned Hollow

Theme:

- overgrown ruin
- venom roots
- cursed hunting grounds

Set family:

- `Briarbound Sight`

Loot identity:

- precision
- crit
- target focus

Bot-facing summary:

- best for accuracy-sensitive builds
- best for execution and priority-target removal
- strongest when the team can maintain tempo and focus fire

Set effects:

- `2-piece`: +10% `accuracy`, +8% `speed`
- `4-piece`: the first hit against a full-HP enemy each round gains +12% crit chance and +18% crit damage
- `6-piece`: when landing a critical hit, gain `1` stack of Hunter's Focus for `2` rounds; at `2` stacks, the next damaging skill gains +15% final damage and ignores `12%` of the target's defense

Suggested side flavor:

- future expansion can let this set slightly extend mark, poison, or exposed-target mechanics

## 7.3 Sunscar Warvault

Theme:

- buried war armory
- scorched legion relics
- siege pressure

Set family:

- `Sunscar Assault`

Loot identity:

- physical offense
- burst windows
- elite and boss breakpoints

Bot-facing summary:

- best for warrior-led physical teams
- strongest for short kill windows
- weaker on defensive recovery than the Catacomb set

Set effects:

- `2-piece`: +12% `physical_attack`
- `4-piece`: the first active damaging skill each battle gains +20% damage; if it defeats a target, recover `1` skill cooldown turn from a random non-ultimate skill
- `6-piece`: after defeating an elite, boss add, or boss phase threshold, gain `2` rounds of +15% `physical_attack` and +10% `speed`

Suggested side flavor:

- later tuning can add bonus effect against armor-broken targets

## 7.4 Obsidian Spire

Theme:

- volcanic black tower
- void priests
- mirror curses and spell engines

Set family:

- `Nightglass Arcanum`

Loot identity:

- spell damage
- cast chaining
- magic burst and sustain casting

Bot-facing summary:

- best for mage and priest spell throughput
- strongest in builds that rotate multiple active skills
- especially valuable in long boss fights

Set effects:

- `2-piece`: +12% `magic_attack`, +10% `healing_power`
- `4-piece`: after casting a non-basic skill, gain Arc Echo; the next magic or holy skill within `2` rounds gains +16% effect value
- `6-piece`: every `3` rounds, the next magic or holy skill costs no tempo penalty, gains +20% effect value, and reduces another random skill cooldown by `1`

Suggested side flavor:

- later tuning can let this set convert over-heal into a minor shield

## 8. Dungeon Info Exposure for Bots

Bots should be able to read dungeon output identity directly before entering.

Each dungeon definition should expose at least:

- `dungeon_id`
- `name`
- `set_id`
- `set_identity_tags`
- `set_bonus_preview`
- `weapon_pool_summary`
- `armor_pool_summary`
- `material_identity`
- `difficulty_summary`

Example identity tags:

- Ancient Catacomb: `defense`, `block`, `survival`
- Thorned Hollow: `accuracy`, `crit`, `focus_fire`
- Sunscar Warvault: `physical_attack`, `burst`, `execution`
- Obsidian Spire: `magic_attack`, `holy_power`, `cast_chain`

## 9. Rating and Reward Rules

Dungeon progression still uses a `6`-room structure with end-of-run rating.

Suggested mapping:

| Highest Completed Room | Rating |
| --- | --- |
| 6 | `S` |
| 5 | `A` |
| 4 | `B` |
| 3 | `C` |
| 2 | `D` |
| 1 or failed before room 1 clear | `E` |

Rules:

- rating determines end-of-run equipment reward quality
- monster kills determine material rewards
- all four dungeons use the same reward-shape logic
- only `set_id` and material identity change by dungeon

## 10. Equipment Reward Pool

Shared reward rules:

- all four dungeons use the same slot distribution and quality distribution tables
- dungeon selection changes set family, not the overall item power budget
- red and prismatic pieces from a dungeon always belong to that dungeon's set family
- blue, purple, and gold pieces may be non-set items from the shared seasonal pool

Suggested slot drop weight:

| Slot | Weight |
| --- | --- |
| Weapon | 18% |
| Head | 16% |
| Chest | 18% |
| Boots | 16% |
| Ring | 16% |
| Necklace | 16% |

## 11. Difficulty Reward Efficiency

The four seasonal dungeons share the same reward structure, but difficulty changes how efficiently bots progress toward real set completion.

Design intent:

- `easy` is the onboarding tier and should be worth clearing early, but not the best long-term farm once `hard` is stable
- `hard` is the transition tier and should reliably move bots into real set ownership
- `nightmare` is the long-term farming tier for red pieces, prismatic chase pieces, and affix refinement

Recommended equipment reward rolls by difficulty and rating:

| Difficulty | Rating | Equipment rolls | Quality distribution per roll | Material multiplier | Intended use |
| --- | --- | --- | --- | --- | --- |
| easy | `S` | 2 | Blue `50%`, Purple `32%`, Gold `14%`, Red `4%`, Prismatic `0%` | `1.00x` | first 2-piece starts, early replacement upgrades |
| easy | `A` | 1 | Blue `56%`, Purple `28%`, Gold `13%`, Red `3%`, Prismatic `0%` | `0.85x` | acceptable recovery reward |
| easy | `B` | 1 | Blue `68%`, Purple `22%`, Gold `9%`, Red `1%`, Prismatic `0%` | `0.70x` | weak fallback, not a stable farm target |
| hard | `S` | 2 | Blue `28%`, Purple `34%`, Gold `24%`, Red `12%`, Prismatic `2%` | `1.30x` | reliable progression into real set ownership |
| hard | `A` | 2 | Blue `35%`, Purple `34%`, Gold `21%`, Red `9%`, Prismatic `1%` | `1.10x` | solid mid-season farming baseline |
| hard | `B` | 1 | Blue `44%`, Purple `31%`, Gold `18%`, Red `6%`, Prismatic `1%` | `0.90x` | okay while stabilizing the difficulty |
| nightmare | `S` | 3 | Blue `12%`, Purple `24%`, Gold `28%`, Red `28%`, Prismatic `8%` | `1.70x` | primary endgame farm line |
| nightmare | `A` | 2 | Blue `18%`, Purple `28%`, Gold `28%`, Red `22%`, Prismatic `4%` | `1.45x` | default efficient farm for most mature bots |
| nightmare | `B` | 2 | Blue `26%`, Purple `30%`, Gold `24%`, Red `17%`, Prismatic `3%` | `1.20x` | acceptable if the bot can finish but not dominate |

Rules:

- `C`, `D`, and `E` should award materials only and should not grant rating-based equipment rolls
- red and prismatic results always use the active dungeon's `set_id`
- blue, purple, and gold results may use the shared seasonal off-set pool
- difficulty changes efficiency, not the base stat budget of an item at the same quality

### 11.1 Bot Farming Guidance

Bots should use the reward table with the recommended power bands instead of always forcing the highest unlocked difficulty.

Suggested logic:

- clear `easy` aggressively until `hard` median is within reach, especially if a desired 2-piece bonus can be activated quickly
- farm `hard` as the main bridge if `nightmare` clear confidence is below the target stable line
- switch to `nightmare` as the main destination once the bot can sustain at least `A` clears
- only stay on `easy` after day `10` for emergency slot replacement or if a build specifically needs a fast 2-piece stopgap

### 11.2 Season Completion Pace Targets

Assuming the bot is farming one chosen dungeon for one main build and is not requiring perfect affixes:

| Farming profile | Expected milestone cadence |
| --- | --- |
| stable `easy S` from early season | first usable `2-piece` in about `5-7` reward claims; not a reliable path to `6-piece` graduation |
| stable `hard A-S` around day `10` | `2-piece` in `4-6` claims, `4-piece` in `12-18` claims, `6-piece` in `30-42` claims |
| stable `nightmare A` around day `15` | `2-piece` in `3-5` claims, `4-piece` in `10-15` claims, `6-piece` in `24-34` claims |
| stable `nightmare S` after build stabilization | first meaningful affix refinement in `45-60` claims; near-finished main build in `80-110` claims |

Interpretation:

- a bot that reaches stable `nightmare` by about day `15` should spend the rest of the season improving set coverage and affix quality, not hunting raw item tier unlocks
- full prismatic best-in-slot completion should remain unlikely for every slot within one season
- off-class weapon drops are part of the time budget and are intentionally offset by salvage value

## 12. Salvage and Duplicate Value

Duplicate gear must remain valuable.

Recommended salvage outputs:

- blue: seasonal dust
- purple: seasonal dust + affix shards
- gold: seasonal dust + polish cores
- red: seasonal dust + set embers
- prismatic: seasonal dust + prismatic thread

This is especially important because:

- all weapon families can drop in every dungeon
- off-class drops are intentional, not a mistake
- the system depends on duplicates having long-term value

## 13. Implementation Notes

Recommended item fields:

- `item_id`
- `template_id`
- `season_band`
- `slot`
- `weapon_style`
- `rarity`
- `set_id`
- `class_restriction`
- `base_stats`
- `affixes`
- `drop_source_type`
- `drop_source_id`

Recommended API requirements:

- every dungeon response should show set identity and reward summary directly
- every item response should expose slot, rarity, set membership, and weapon style explicitly
- result payloads should separate rating-based equipment rewards from in-combat material drops

## 14. Open Follow-Up Tasks

- decide whether every dungeon should have one boss material and one generic seasonal material, or multiple unique materials
- define whether block is a formal combat stat or an event-style defensive keyword in V1
