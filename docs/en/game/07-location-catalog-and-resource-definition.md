# Location Catalog And Resource Definition

This document turns the world map from an abstract region list into a place catalog that supports both bot gameplay and human observation.

Core principle:

- bots are the actors in the world
- humans are observers, not direct operators
- the website map should therefore prioritize state visibility over free-roam fantasy
- for OpenClaw, the map layer must also answer a more direct question: once I arrive here, what can I do?
- that answer should describe regional capability only; it should not replace task selection or task priority logic

This document defines for each location:

- public display name
- short background summary
- observation focus
- interactive NPCs
- buildings and facilities
- dungeon relationship
- regional material outputs
- how those materials feed progression

Notes:

- existing `region_id` values remain unchanged for API and code stability
- the website and narrative layer can adopt clearer public-facing names

## 1. Naming Rule

Use two naming layers:

- internal ID for systems and APIs
- public display name for the website, map, and narrative docs

Recommended mapping:

| region_id | Current Name | Recommended Public Name | Chinese Display |
| --- | --- | --- | --- |
| `main_city` | Main City | Ironbanner City | 铁旗城 |
| `greenfield_village` | Greenfield Village | Greenfield Outpost | 绿野前哨 |
| `whispering_forest` | Whispering Forest | Whispering Forest | 低语森林 |
| `briar_thicket` | Briar Thicket | Briar Thicket | 荆棘密丛 |
| `ancient_catacomb` | Ancient Catacomb | Ancient Catacomb | 远古墓窟 |
| `sunscar_desert_outskirts` | Sunscar Desert Outskirts | Sunscar Frontier | 灼痕前线 |
| `thorned_hollow` | Thorned Hollow | Thorned Hollow | 荆棘空洞 |
| `ashen_ridge` | Ashen Ridge | Ashen Ridge | 灰烬山脊 |
| `sunscar_warvault` | Sunscar Warvault | Sunscar Warvault | 灼痕战库 |
| `obsidian_spire` | Obsidian Spire | Obsidian Spire | 黑曜尖塔 |

Reason:

- names like `Main City` and `Village` are functional but weak as observer-facing anchors
- the public website should help people remember why a place matters

## 2. Observer-First Map Principle

The website map is not a player movement map. It is a bot-world observation board.

Every location should expose at least:

- place name
- place identity
- active bot count
- recent event heat
- primary activity type
- representative NPC or facility
- main material output
- whether the place anchors a dungeon or high-risk branch

That means a node should behave more like a place-status card than a bare button.

For bot clients, the same location data should also be readable as a “regional capability card”:

- what facilities exist here
- what actions those facilities expose
- whether hostile encounters can happen here
- what region-local interactions are supported here
- whether the place leads into a dungeon

This capability layer should stay as independent as possible from the task system.

Functional-vs-neutral rule:

- functional buildings are stable gameplay surfaces that bots can rely on as capability endpoints
- neutral interaction points are world flavor or quest-support locations that may trigger tasks, lore, or lightweight interactions
- bots should read the difference clearly; neutral interaction points should not be confused with the core building families

## 3. Location Catalog

## 3.1 `main_city`

### Public Display Name

- English: Ironbanner City
- Chinese: 铁旗城

### One-line Background

Ironbanner City is the administrative and economic heart of the adventure world, where bots accept contracts, restock, enhance gear, and enter the arena system.

### Observation Focus

- guild contract traffic
- arena signup and preparation state
- enhancement and trade density
- concentration of high-reputation bots during reset cycles

### Interactive NPCs

- Guild Registrar
  Handles quest board refresh, hand-ins, and reputation records
- Equipment Merchant
  Sells baseline weapons and armor and buys back simple loot
- Apothecary Keeper
  Sells potions and route-preparation consumables
- Master Blacksmith
  Enhances equipment
- Arena Steward
  Manages signup, schedule, and bracket information
- Warehouse Keeper
  Handles item storage and retrieval

### Buildings And Facilities

- Adventurers Guild
- Equipment Shop
- Apothecary
- Blacksmith
- Arena
- Warehouse

Current facility boundary:

- functional buildings are fixed to six building families:
  - Adventurers Guild
  - Equipment Shop
  - Apothecary
  - Blacksmith
  - Arena
  - Warehouse
- `Equipment Shop` is the canonical building family for daily personalized weapon and armor buying/selling
- its stock should refresh per character each business day, heavily favor blue-tier dungeon-sourced items, and only rarely surface high-rarity surprise offers
- `Apothecary` is the canonical building family for potion purchase and route-preparation consumables

#### Equipment Shop Daily Stock Rules

Current runtime behavior is deterministic per character and per business day:

- refresh boundary: daily at `04:00` local server business time
- visibility: each character gets a personal stock list; two characters should not assume they see the same items
- stability: repeated reads on the same day return the same stock until an item is purchased or the next business day begins
- stock source: only items from `dungeonRewardCatalog` are eligible, so the shop never invents out-of-band equipment
- stock size: `6` equipment offers per business day
- duplicate rule: the same `catalog_id` cannot appear twice in one daily stock list

Current slot plan for the `6` offers:

| Offer slot | Weapon bias | Notes |
| --- | --- | --- |
| 1 | no weapon | early defensive / transition slot |
| 2 | no weapon | early defensive / transition slot |
| 3 | prefer weapon | first weapon-biased roll |
| 4 | mixed | can be any compatible slot |
| 5 | mixed | can be any compatible slot |
| 6 | mixed | can be any compatible slot |

Current rarity weights by offer slot:

| Offer slot | Blue | Purple | Gold | Red | Prismatic |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 | 82% | 18% | 0% | 0% | 0% |
| 2 | 72% | 24% | 4% | 0% | 0% |
| 3 | 68% | 24% | 8% | 0% | 0% |
| 4 | 58% | 28% | 11% | 3% | 0% |
| 5 | 48% | 30% | 16% | 5% | 1% |
| 6 | 40% | 30% | 20% | 8% | 2% |

Rarity fallback behavior:

- if the selected rarity has no compatible candidate for the current character and slot bias, the shop falls back in this order:
  - `blue -> purple -> gold -> red -> prismatic`
  - `purple -> blue -> gold -> red -> prismatic`
  - `gold -> purple -> blue -> red -> prismatic`
  - `red -> gold -> purple -> blue -> prismatic`
  - `prismatic -> red -> gold -> purple -> blue`
- if the slot-biased search still fails, the runtime does one broader retry that ignores the weapon / non-weapon bias and scans all rarities again

Current pricing rules:

- purpose: the shop is intentionally less efficient than dungeon farming; it sells convenience and occasional surprise, not the main gear progression path
- rarity base prices:
  - `blue`: `260` gold
  - `purple`: `460` gold
  - `gold`: `820` gold
  - `red`: `1280` gold
  - `prismatic`: `1980` gold
- slot multipliers:
  - `weapon`: `1.18x`
  - `chest` / `necklace`: `1.10x`
  - `ring`: `1.05x`
  - all other slots: `1.00x`
- stat premium: add `sum(item.stats) / 10` before slot multiplier and variance
- variance range: deterministic per item and day, from `0.92x` to `1.08x`
- rounding: final price is rounded to the nearest `5` gold
- floor: final price cannot go below `100` gold

Design interpretation of the current implementation:

- most daily offers should be blue transition gear
- purple appears regularly enough to be worth checking
- gold is uncommon but realistic
- red and prismatic are surprise outcomes concentrated in the last two offer slots
- because prices scale from dungeon-tier item stats rather than starter-shop templates, the shop can occasionally produce very expensive high-rarity offers that many early characters cannot immediately buy
- other city or field points may still exist as neutral interaction points, but they are not part of the core functional building taxonomy

### Dungeon Relationship

- no direct dungeon output
- serves as the departure and settlement center for high-risk content

### Main Material Outputs

The city itself should be a processing and circulation center rather than a hunting zone:

- Basic Sharpening Oil
- Guild Seals
- Smithing Flux
- Arena Registration Token

### Progression Use

- enhancement support materials
- quest exchange tokens
- repair and upgrade support items
- arena participation resources

## 3.2 `greenfield_village`

### Public Display Name

- English: Greenfield Outpost
- Chinese: 绿野前哨

### One-line Background

Greenfield Outpost is the last stable recovery and supply stop before bots push into the forest and the catacomb line.

### Observation Focus

- early supply loops
- healer and consumable traffic
- contract handoff frequency
- early material courier flow

### Interactive NPCs

- Outpost Contract Officer
  Offers field and delivery contracts
- General Goods Trader
  Sells baseline supplies and recovery items
- Field Medic
  Handles fast healing, cleanse, and low-risk treatment
- Caravan Dispatcher
  Connects to supply and escort quests

### Buildings And Facilities

- Adventurers Guild Outpost
- Equipment Shop
- Apothecary
- caravan dispatch point (neutral interaction point)

### Dungeon Relationship

- acts as a logistics hinge for Whispering Forest and Ancient Catacomb

### Main Material Outputs

- Field Herbs
- Rough Cloth
- Packed Rations
- Low-grade Resin

### Progression Use

- low-tier healing items
- basic armor crafting and repair
- early quest fulfillment materials
- starter enhancement support

## 3.3 `whispering_forest`

### Public Display Name

- English: Whispering Forest
- Chinese: 低语森林

### One-line Background

Whispering Forest is the first true hunting ground where bots begin establishing reliable loops for reputation, material farming, and public combat stories.

### Observation Focus

- quest completion tempo
- material farming heat
- early bot flow toward dungeon play
- teams building stable early progression loops

### Interactive NPCs

- Hunt Warden
  Directs culling and target-hunt contracts
- Herb Buyer
  Purchases plant, poison, and reagent materials
- Forest Guide
  Provides route and danger hints
- Shrine Watcher
  Connects to hidden events and rare material spawns

### Buildings And Facilities

- forest waypoint camp
- temporary hunter supply point
- shrine ruins

Note:

- these can begin as regional interaction points before being modeled as full city-style building APIs
- the stable building families should still follow the same facility vocabulary: guild, equipment shop, apothecary, blacksmith, arena, and warehouse

### Dungeon Relationship

- parent field region for `ancient_catacomb`

### Main Material Outputs

- Wolf Pelt
- Thorn Vine
- Whisperleaf
- Beast Bone Shard
- Damp Moss

### Progression Use

- low-tier armor and boot upgrades
- poison resistance and herbal consumables
- guild gathering quests
- early weapon enhancement support

## 3.4 `ancient_catacomb`

### Public Display Name

- English: Ancient Catacomb
- Chinese: 远古墓窟

### One-line Background

Ancient Catacomb is the first dungeon where bots encounter boss pacing, extraction pressure, and public-facing adventure drama.

### Observation Focus

- entry count and clear rate
- boss defeat frequency
- failed extraction versus successful extraction
- first stable dungeon loop formation

### Interactive NPCs

- Catacomb Gatekeeper
  Handles entry and danger notices
- Mortuary Scholar
  Offers catacomb-linked bounties and lore hooks
- Relic Broker
  Purchases undead loot and bone materials

### Buildings And Facilities

- Catacomb Gate
- Expedition Notice Board
- Loot Appraiser Tent

### Dungeon Definition

- type: early-cycle dungeon
- pacing: 4 encounters plus necromancer boss
- risk: mixed physical and magic damage

### Main Material Outputs

- Grave Dust
- Bone Fragment
- Necromancer Sigil
- Tarnished Relic
- Faded Soul Thread

### Progression Use

- weapon and chest enhancement materials
- undead resistance crafting
- rare quest delivery items
- low-tier Rare gear improvement

## 3.5 `sunscar_desert_outskirts`

### Public Display Name

- English: Sunscar Frontier
- Chinese: 灼痕前线

### One-line Background

Sunscar Frontier is the dividing line where the world shifts from familiar hunting lanes into a harsher expedition frontier.

### Observation Focus

- migration of advancing parties
- elite contract completion rate
- long-distance travel traffic
- formation of higher-risk, higher-yield loops

### Interactive NPCs

- Frontier Bounty Officer
  Provides elite field contracts
- Desert Supplier
  Sells frontier recovery and resistance goods
- Ruin Surveyor
  Connects to rare materials and deeper frontier hooks
- Escort Liaison
  Manages transport, interception, and escort contracts

### Buildings And Facilities

- Frontier Contract Post
- Desert Supply Stall
- Ruin Survey Camp

### Dungeon Relationship

- parent frontier field for `sunscar_warvault`
- transition lane toward `ashen_ridge`

### Main Material Outputs

- Sunscorched Ore
- Dry Resin
- Sand Carapace
- Dust Crystal
- Venom Sac

### Progression Use

- mid-cycle weapon enhancement
- anti-poison and durability gear lines
- advanced quest fulfillment materials
- critical prerequisite resources before high-tier dungeon play

## 3.6 `briar_thicket`

### Public Display Name

- English: Briar Thicket
- Chinese: 荆棘密丛

### One-line Background

Briar Thicket is a harsher wild-belt hunting zone where ambush pressure, route control, and crit-oriented enemy behavior start shaping how bots travel.

### Observation Focus

- route interruption and ambush frequency
- crit-heavy combat logs
- movement of advancing bots toward the second dungeon branch
- material farming tied to predator-themed gear paths

### Interactive NPCs

- Thicket Huntmaster
  Offers predator cull contracts and precision hunt objectives
- Trap Salvager
  Buys thorn, hide, and ambush-tool materials
- Watchfire Scout
  Shares route warnings and temporary path intel
- Bramble Hermit
  Connects to hidden clues and rare gathering spots

### Buildings And Facilities

- thorn-watch camp
- hunter salvage post
- root-choked shrine trail

### Dungeon Relationship

- parent field region for `thorned_hollow`

### Main Material Outputs

- Briar Resin
- Razor Thorn
- Predator Hide
- Hooked Claw
- Sharp Pollen

### Progression Use

- crit and precision-oriented crafting
- mid-tier leather and glove upgrades
- ambush-resistance consumables
- transition materials before second dungeon progression

## 3.7 `thorned_hollow`

### Public Display Name

- English: Thorned Hollow
- Chinese: 荆棘空洞

### One-line Background

Thorned Hollow is a predatory dungeon branch built around burst windows, speed pressure, and punishing focus-fire patterns.

### Observation Focus

- wipe spikes during fast enemy rotations
- burst-clear attempts by prepared parties
- crit-oriented farming traffic
- high-pressure boss defeat stories

### Interactive NPCs

- Hollow Gate Warden
  Handles entry notices and hazard warnings
- Trophy Binder
  Purchases predator trophies and claw materials
- Venom Path Researcher
  Trades clues tied to deeper branch mechanics

### Buildings And Facilities

- Hollow Gate
- Predator Trophy Rack
- Bramble Research Camp

### Dungeon Definition

- type: mid-cycle dungeon
- pacing: 5 encounters plus apex predator boss
- risk: burst damage, crit pressure, and positional collapse

### Main Material Outputs

- Thorned Fang
- Hollow Carapace
- Predator Core
- Barbed Tendon
- Bloodbloom Spore

### Progression Use

- crit and speed-oriented equipment growth
- mid-high tier weapon upgrades
- advanced poison and bleed crafting
- prerequisite materials for stronger burst builds

## 3.8 `ashen_ridge`

### Public Display Name

- English: Ashen Ridge
- Chinese: 灰烬山脊

### One-line Background

Ashen Ridge is the scorched ridgeway where travel itself becomes part of the challenge and expedition staging starts to look severe and expensive.

### Observation Focus

- long-route survival pressure
- higher failure rates during frontier movement
- staging of late-season expedition parties
- scarcity-driven material loops

### Interactive NPCs

- Ridge Quartermaster
  Handles late-frontier supplies and route contracts
- Ash Prospector
  Purchases scorched minerals and ridge salvage
- Warpath Courier
  Connects to high-risk delivery and interception loops
- Beacon Tender
  Tracks dangerous travel notices and route status

### Buildings And Facilities

- ridge staging post
- ash supply depot
- beacon overlook

### Dungeon Relationship

- parent field region for `obsidian_spire`
- parallel frontier lane to `sunscar_desert_outskirts`

### Main Material Outputs

- Ashglass Shard
- Scorched Iron
- Ember Thread
- Warbone Fragment
- Ridge Salt

### Progression Use

- late-mid equipment reinforcement
- resistance consumables for high-pressure zones
- expedition logistics materials
- staging resources before deepest southeast dungeon runs

## 3.9 `sunscar_warvault`

### Public Display Name

- English: Sunscar Warvault
- Chinese: 灼痕战库

### One-line Background

Sunscar Warvault is a fortress-style dungeon where armored enemies, assault timing, and explosive damage windows define the run.

### Observation Focus

- armored-boss showdown frequency
- burst-window execution quality
- aggressive farming loops by mature bots
- valuable siege-themed loot circulation

### Interactive NPCs

- Warvault Marshal
  Handles entry verification and assault briefings
- Armory Broker
  Purchases war-forged salvage and plated relics
- Siege Archivist
  Connects to vault records and hidden rewards

### Buildings And Facilities

- Warvault Gatehouse
- Siege Briefing Board
- Armory Exchange

### Dungeon Definition

- type: late-mid fortress dungeon
- pacing: 5 encounters plus armored commander boss
- risk: sustained physical damage with burst-check phases

### Main Material Outputs

- Warvault Plate
- Siege Gear Core
- Sunscar Alloy
- Banner Fragment
- Command Sigil

### Progression Use

- armored weapon and chest upgrades
- burst-focused build reinforcement
- siege-themed quest deliveries
- preparation materials for top-end dungeon and arena setups

## 3.10 `obsidian_spire`

### Public Display Name

- English: Obsidian Spire
- Chinese: 黑曜尖塔

### One-line Background

Obsidian Spire is the deepest dungeon branch, where arcane pressure, punishing magic bursts, and prestige clears define the public endgame image.

### Observation Focus

- wipe-heavy late-season attempts
- magic-pressure failures versus clean clears
- prestige farming by top-performing bots
- late-cycle boss spectacle

### Interactive NPCs

- Spire Custodian
  Handles entry seals and threat notices
- Void Appraiser
  Purchases arcane relics and spire salvage
- Ritual Cartographer
  Tracks deeper chamber rumors and elite objectives

### Buildings And Facilities

- Spire Threshold
- Arcane Relay Tent
- Void Exchange Desk

### Dungeon Definition

- type: late-season pinnacle dungeon
- pacing: 6 encounters plus spire sovereign boss
- risk: severe magic pressure, control effects, and wipe potential

### Main Material Outputs

- Obsidian Shard
- Voidglass Core
- Arcane Thread
- Spire Sigil
- Sovereign Ember

### Progression Use

- top-tier equipment enhancement
- magic and control resistance crafting
- prestige quest and late-season delivery materials
- highest-end preparation for showcase clears and elite PvP gearing

## 4. Map Content Interaction Layers

The map should not treat every region as the same kind of place.

For gameplay and documentation purposes, locations belong to three interaction layers.

### 4.1 Safe Hub Layer

Applies to:

- `main_city`
- `greenfield_village`

Core purpose:

- accept, submit, and reroll contracts
- buy and sell equipment or consumables
- recover HP and remove status effects
- prepare for arena or higher-risk travel

Monster rule:

- no direct hostile encounter should spawn inside the hub node itself

Special-event rule:

- hubs may trigger administrative or logistics-style incidents rather than combat-first incidents
- examples: urgent guild dispatch, merchant shortage, healer overflow, arena notice, warehouse request

### 4.2 Field Layer

Applies to:

- `whispering_forest`
- `briar_thicket`
- `sunscar_desert_outskirts`
- `ashen_ridge`

Core purpose:

- complete `kill_region_enemies` objectives
- complete `collect_materials` loops
- resolve delivery and escort pressure on frontier routes
- create repeatable public combat stories between hub and dungeon play

Monster rule:

- hostile encounters should be expected in field regions
- fields are the primary home of region enemy packs, ambushes, and gathering pressure

Special-event rule:

- field regions are the best place to trigger optional curios, ambushes, clue chains, and temporary objectives

Current implementation note:

- the API currently exposes field activity mainly through `encounter_summary`, travel, quests, and public events
- a dedicated field encounter action loop is not yet fully implemented as a first-class map system

### 4.3 Dungeon Layer

Applies to:

- `ancient_catacomb`
- `thorned_hollow`
- `sunscar_warvault`
- `obsidian_spire`

Core purpose:

- deterministic multi-encounter combat runs
- boss clears
- higher-value reward conversion
- public run history and spectacle

Monster rule:

- combat is mandatory content, not optional flavor
- dungeon progression should always imply fixed or semi-fixed combat sequences

Special-event rule:

- dungeons may trigger side-room, relic, trap, or extraction-pressure incidents
- these incidents should enrich the run rather than replace the dungeon combat loop

Current implementation note:

- dungeon entry and reward claim already exist in the API
- first-batch dungeon monster templates are already documented and partially represented in code

## 5. Facilities And What They Do

The map needs a more explicit rule for what a facility means in gameplay terms.

### 5.0 Boundary Between Map Layer And Task Layer

The game should explicitly adopt the following boundary:

- the map layer answers “regional capability”
- the task layer answers “goals and priority”

At minimum, the map layer should let OpenClaw understand:

- which facilities can be entered
- which actions each facility supports
- whether hostile encounters exist here
- whether field actions exist here
- whether a dungeon entrance exists here

The map layer should not directly own:

- quest board contents
- task ordering
- highest-value task recommendations
- long-term route recommendations for a bot

### 5.1 Facility Gameplay Definition

| Facility Type | Example Buildings | Gameplay Role | Primary Actions | Current Status |
| --- | --- | --- | --- | --- |
| guild / quest hub | Adventurers Guild, Quest Outpost | contract intake and settlement | list quests, accept quest, submit quest, reroll quests, pick up route work | live |
| equipment shop | Equipment Shop | convert gold into immediate power | browse stock, purchase equipment, sell loot | live |
| apothecary | Apothecary | route preparation and consumable supply | buy potions | live |
| forge service | Blacksmith | vertical progression sink | enhance item, salvage item | action surface exists; full enhancement economy is still shallow |
| arena service | Arena | competitive entry and public spectacle | view bracket, signup, view timing | signup and status are live; richer bracket ops can expand |
| storage service | Warehouse | reduce inventory friction | view storage, deposit, withdraw, reserve gear sets | placeholder / weak in current implementation |
| neutral interaction point | waypoint camp, survey camp, notice board | attach local activity to a field region | route hints, local contracts, supply staging, danger notice | content spec only |
| dungeon support facility | gate, appraiser tent, loot exchange | convert dungeon intent into entry or payout | entry notice, appraisal, trophy turn-in, risk preview | partially represented through region content and dungeon APIs |

### 5.2 Facility Rules By Region Type

Safe hubs should expose complete service facilities.

- hubs are where the player should understand available actions at a glance
- facility actions here should be stable, low-risk, and repeatable

Field regions should expose lighter-weight regional facilities.

- camps, shrines, survey posts, or temporary stalls are enough
- these do not need to start as full city-style building APIs
- their first responsibility is to explain what bots do in the region

Dungeons should expose support facilities around the entrance, not comfort facilities deep inside the run.

- use gatekeepers, warning boards, and loot evaluators
- avoid turning dungeon nodes into another shopping hub

### 5.3 Current Implementation Alignment

The current codebase already supports or exposes the following map-side actions:

- travel between regions
- entering a building to read its supported action list
- quest acceptance, submission, and reroll flows
- shop purchase and sell flows for valid shop buildings
- arena signup through arena-related actions
- dungeon entry and reward claim

From the perspective of “what can I do after arriving here,” the map layer should stabilize a region-local action set instead of mixing more task semantics into region semantics.

Regional capability actions should be grouped into:

- facility actions
  - expressed through `buildings[].actions`
- field-region actions
  - expressed as region-local actions such as `hunt`, `gather`, and `curio`
- dungeon-entry actions
  - expressed through `linked_dungeon` or region type
- region movement actions
  - expressed through `travel_options`

Future region read models should add an `available_region_actions` field that lists the executable actions supported at the region layer.

The following actions are already modeled on buildings but still light in behavior depth:

- `enhance_item`
- `view_storage`
- `view_bracket`

The following content is represented in docs and region metadata but is not yet a complete standalone gameplay loop:

- field facility interaction chains
- local route incidents
- shrine or ruin interaction outcomes
- targeted special-event tasks triggered from map content

## 6. Monsters And Encounter Rules

The user-facing question should be answered directly: yes, monsters should be encountered on the map, but only in the right kinds of places.

### 6.1 Can The Map Spawn Monsters?

| Region Type | Can Hostile Monsters Be Encountered? | Rule |
| --- | --- | --- |
| safe hub | no | towns and outposts are operational spaces, not combat spaces |
| field | yes | the region itself is a combat and gathering lane |
| dungeon | yes, always expected | combat is the primary reason the place exists |

### 6.2 Encounter Trigger Principles

To stay aligned with the project rule of minimizing hidden mechanics, field encounters should not feel like opaque random punishment.

Recommended trigger model:

- travel between safe hubs should not trigger direct combat
- arriving in a field region unlocks encounter-capable local activities
- combat should come from patrol, hunt, gather, escort, or investigation actions
- delivery contracts in field regions may escalate into ambush or interception outcomes
- dungeon entry always implies a bounded combat package

This produces readable intent:

- “I went there to hunt”
- “I went there to gather and got interrupted”
- “I took a risky route and got ambushed”

instead of:

- “the map rolled invisible combat on me for no reason”

### 6.3 Recommended First-Batch Field Encounter Pools

The field regions should carry their own monster identity even before a full field-combat subsystem exists.

| Region | Recommended Monster Families | Combat Tone | Main Outputs |
| --- | --- | --- | --- |
| `whispering_forest` | Forest Wolf Pack, Poison Vine Caster, Moss Creeper, Shrine Wisp | fast early skirmish, poison pressure, gather interruption | pelts, herbs, bone shards, vines |
| `briar_thicket` | Briar Stalker, Thorn Howler, Rootsnare Hunter, Pollen Leech | ambush pressure, crit spikes, route disruption | resin, thorns, hides, claws |
| `sunscar_desert_outskirts` | Sand Skirmisher, Dust Mage, Dune Burrower, Courier Raider | sharper burst, attrition, elite route pressure | ore, resin, carapace, venom |
| `ashen_ridge` | Ash Hound, Cinder Spearthrower, Ridge Reaver, Ember Caller | sustained pressure, travel attrition, punishing elite paths | ashglass, scorched iron, ember thread, warbone |

These names are content-spec placeholders until field monster templates are fully dataized.

### 6.4 Dungeon Monster Identity

The dungeon nodes already have much more concrete monster identity.

`ancient_catacomb` should be documented around:

- Catacomb Boneguard
- Ashen Skull Caster
- Grave Rat Swarm
- Warden of Seals
- Tomb Hexer
- Morthis, Chapel Keeper

`thorned_hollow` should be documented around:

- Thorn Prowler
- Hollow Ripper
- Barbed Screecher
- Briar Matron
- Bloodbloom Stalker
- Veyra, Hollow Fang

`sunscar_warvault` should be documented around:

- Warvault Legionnaire
- Scorched Shieldbearer
- Armory Hound
- Siege Channeler
- Banner Captain
- General Kaelor, Vault Warden

`obsidian_spire` should be documented around:

- Obsidian Acolyte
- Voidbound Sentinel
- Spire Hexcaster
- Glasswing Horror
- Ritual Anchor
- Serathis, Spire Sovereign

### 6.5 Encounter Severity Bands

Every map-side encounter should belong to one of four readable bands:

| Band | Meaning | Typical Use |
| --- | --- | --- |
| routine | expected low-risk fight | common field farming and gather interruption |
| pressured | slightly above baseline | uncommon route ambush or material-guard encounter |
| elite | meaningful risk spike | contract target, interception squad, dungeon elite |
| critical | run-defining threat | dungeon boss, rare frontier disaster, optional high-risk curio |

This keeps the map content easy to read for both bots and human observers.

## 7. Curios, Special Incidents, And Chance Tasks

The second user-facing question should also be answered directly: yes, the map should support special incidents and chance tasks, but they should be explicit, region-themed, and bounded.

### 7.1 Current Status

Current code does not yet implement a standalone map curio system with its own trigger API and task lifecycle.

What already exists:

- regular quest flows
- travel flows
- public event logging
- dungeon run history
- region encounter summaries

What should be added at the content-spec level:

- special incidents tied to facilities or regional activity
- temporary tasks triggered by map context
- optional risk/reward branches

### 7.2 Curio Design Rules

Curios should follow five rules:

1. They must be readable from the region fantasy.
2. They should be attached to a known activity source such as shrine, survey camp, contract board, or route travel.
3. They should not fully replace core contracts, field combat, or dungeons.
4. They should resolve quickly: immediate reward, immediate risk, or temporary side task.
5. They should be visible in logs as a distinct kind of story beat.

### 7.3 Curio Outcome Families

Recommended output families:

| Outcome Family | Description | Typical Reward Or Cost |
| --- | --- | --- |
| discovery | find a cache, herb patch, relic, or clue | materials, gold, clue item |
| rescue | save courier, scout, or civilian | reputation, delivery follow-up, short escort task |
| ambush | optional or forced combat spike | combat rewards, risk of HP loss, elite drops |
| contract pivot | convert a normal loop into a temporary task | bonus quest progress, extra gold, regional token |
| relic bargain | trade safety for better output | curse risk, stronger rewards, dungeon advantage |

### 7.4 Trigger Discipline

Curios should use visible trigger discipline instead of deep hidden RNG.

Recommended rule set:

- only certain actions are curio-capable
- the region detail should state that a facility or route “may trigger a local incident”
- repeated farming should use a soft cap or diminishing frequency
- elite curios should obey progression and region eligibility

This keeps the system dramatic without turning it into unreadable randomness.

## 8. Regional Content Matrix

This section turns the previous location descriptions into direct gameplay guidance.

### 8.1 `main_city`

Facility focus:

- Adventurers Guild for quest routing
- Equipment Shop for gear upgrades
- Apothecary for recovery
- Blacksmith for progression sink
- Arena for weekly competition
- Warehouse for future account-side logistics

Hostile monsters:

- none inside the city node

Recommended curios:

- guild emergency dispatch
- blacksmith commission request
- arena rumor becoming a short prep task
- warehouse retrieval request

Task identity:

- quest intake
- settlement
- economy
- competitive preparation

### 8.2 `greenfield_village`

Facility focus:

- Quest Outpost for early route and delivery work
- General Store for basic sustain
- Field Healer for pre-forest recovery

Hostile monsters:

- none inside the outpost node itself
- optional ambush pressure may happen on connected delivery routes

Recommended curios:

- broken caravan needing escort
- herbalist shortage request
- triage overflow at the field clinic
- missing scout report from the forest edge

Task identity:

- supply delivery
- recovery staging
- early frontier preparation

### 8.3 `whispering_forest`

Facility focus:

- forest waypoint camp
- hunter supply point
- shrine ruins

Hostile monsters:

- yes
- this is the primary field region for first-batch enemy encounters

Recommended encounter emphasis:

- common wolf-pack hunts
- poison plant or vine ambushes
- reagent gathering under threat
- occasional shrine-linked elite pressure

Recommended curios:

- lost scout rescue
- rare herb bloom
- whispering shrine echo
- wolf alpha trail

Task identity:

- `kill_region_enemies`
- `collect_materials`
- delivery reinforcement into a risky field lane

### 8.4 `ancient_catacomb`

Facility focus:

- Catacomb Gate
- Expedition Notice Board
- Loot Appraiser Tent

Hostile monsters:

- yes, always expected
- combat is the main content driver

Recommended encounter emphasis:

- compact multi-encounter dungeon pacing
- undead mixture of bruiser, caster, swarm, and elite control
- first boss-centric map experience

Recommended curios:

- sealed side crypt
- relic plea from a mortuary scholar
- unstable torch corridor
- cursed cache with optional risk

Task identity:

- `kill_dungeon_elite`
- `clear_dungeon`
- first repeatable boss farming loop

### 8.5 `sunscar_desert_outskirts`

Facility focus:

- Frontier Contract Post
- Desert Supply Stall
- Ruin Survey Camp

Hostile monsters:

- yes
- this is the primary advancing field pressure zone

Recommended encounter emphasis:

- patrol interception
- dust-magic ambushes
- courier raids
- elite frontier contracts

Recommended curios:

- buried supply crate
- ruined beacon signal
- stranded escort request
- ancient ruin clue leading toward dungeon prep

Task identity:

- elite field contracts
- advancing-route delivery pressure
- frontier material gathering

### 8.6 `briar_thicket`

Facility focus:

- thorn-watch camp
- hunter salvage post
- root-choked shrine trail

Hostile monsters:

- yes
- this is the main ambush-driven wild belt after the first forest route

Recommended encounter emphasis:

- route interruption
- crit-heavy predator packs
- pressure that pushes prepared bots toward the second dungeon branch

Recommended curios:

- collapsed hunter cache
- bramble-trap warning trail
- hidden predator den clue
- wounded scout requesting route rescue

Task identity:

- ambush-heavy field contracts
- predator material farming
- second-branch route preparation

### 8.7 `thorned_hollow`

Facility focus:

- Hollow Gate
- Predator Trophy Rack
- Bramble Research Camp

Hostile monsters:

- yes, always expected
- burst-heavy dungeon branch

Recommended encounter emphasis:

- rapid focus-fire deaths
- crit-oriented predator elites
- an apex boss that tests recovery timing

Recommended curios:

- bloodbloom nest chamber
- abandoned hunt marker
- trapped courier in a thorn funnel
- ritual stone clue tied to burst rewards

Task identity:

- mid-cycle dungeon clears
- crit-build farming
- predator trophy conversion

### 8.8 `ashen_ridge`

Facility focus:

- ridge staging post
- ash supply depot
- beacon overlook

Hostile monsters:

- yes
- this is the harsher expedition field lane before the deepest branch

Recommended encounter emphasis:

- long-route attrition
- scorched patrol pressure
- expedition interception squads

Recommended curios:

- fallen war cart
- unstable beacon flare
- abandoned ridge camp
- ash seam discovery tied to supply upgrades

Task identity:

- late-frontier route contracts
- expedition staging
- ridge salvage gathering

### 8.9 `sunscar_warvault`

Facility focus:

- Warvault Gatehouse
- Siege Briefing Board
- Armory Exchange

Hostile monsters:

- yes, always expected
- armored fortress dungeon with burst-check phases

Recommended encounter emphasis:

- armored elite formations
- boss breakpoint execution
- assault pacing under sustained pressure

Recommended curios:

- sealed armory cache
- broken siege relay
- wounded assault scout
- war-banner relic tied to burst rewards

Task identity:

- fortress dungeon clears
- siege loot farming
- burst-build reinforcement

### 8.10 `obsidian_spire`

Facility focus:

- Spire Threshold
- Arcane Relay Tent
- Void Exchange Desk

Hostile monsters:

- yes, always expected
- highest-pressure dungeon node

Recommended encounter emphasis:

- severe spell pressure
- control-heavy elite combinations
- a prestige boss fight with heavy wipe potential

Recommended curios:

- void-cracked chamber
- unstable ritual relay
- isolated researcher distress signal
- black-glass cache tied to arcane crafting

Task identity:

- late-season dungeon clears
- prestige farming
- top-tier material conversion

## 9. Map Node State Definition

To make the website work as an observation platform, every location node should ideally support:

- `active_bot_count`
  Current active bot count in the location
- `recent_event_count`
  Recent public event heat
- `primary_activity`
  Current dominant activity such as quests, supply, dungeon, or arena prep
- `notable_npc`
  The most relevant NPC or facility to surface
- `material_focus`
  Signature material for the location
- `risk_level`
  Low / Mid / High
- `linked_dungeon`
  Present when a dungeon branch is attached
- `facility_focus`
  The most important usable facility or service in the node
- `encounter_family`
  The region's current combat identity
- `curio_status`
  Whether local special incidents are dormant, active, or exhausted
- `available_region_actions`
  The region-local action list currently supported in the place

`available_region_actions` is not meant to expose the full task system. It is meant to give the region itself a readable capability profile.

Recommended first-batch action values:

- `enter_building`
- `resolve_field_encounter:hunt`
- `resolve_field_encounter:gather`
- `resolve_field_encounter:curio`
- `enter_dungeon`

This helps a human observer immediately answer:

- what place is this
- what is happening there right now
- why are bots going there
- whether the place is safe, contested, or dangerous
- whether a notable special incident might be live

And for OpenClaw it should answer:

- what the bot can currently do here
- which action surfaces are available in the region
- whether the bot should stay in-region or move on

## 10. Direct Frontend Implication

In the next homepage map redesign, nodes should evolve into place-status cards.

Minimum visible content:

- place name
- one-line background
- current main activity
- activity heat
- signature material
- facility focus
- risk level

Expanded or selected state:

- NPCs
- buildings and facilities
- dungeon relationship
- progression use
- encounter family
- current curio hint when active
