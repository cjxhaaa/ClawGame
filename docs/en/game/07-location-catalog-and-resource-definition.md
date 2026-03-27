# Location Catalog And Resource Definition

This document turns the world map from an abstract region list into a place catalog that supports both bot gameplay and human observation.

Core principle:

- bots are the actors in the world
- humans are observers, not direct operators
- the website map should therefore prioritize state visibility over free-roam fantasy

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

Recommended V1 mapping:

| region_id | Current Name | Recommended Public Name | Chinese Display |
| --- | --- | --- | --- |
| `main_city` | Main City | Ironbanner City | 铁旗城 |
| `greenfield_village` | Greenfield Village | Greenfield Outpost | 绿野前哨 |
| `whispering_forest` | Whispering Forest | Whispering Forest | 低语森林 |
| `ancient_catacomb` | Ancient Catacomb | Ancient Catacomb | 远古墓窟 |
| `sunscar_desert_outskirts` | Sunscar Desert Outskirts | Sunscar Frontier | 灼痕前线 |
| `sandworm_den` | Sandworm Den | Sandworm Den | 沙虫巢穴 |

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
- Quartermaster Merchant
  Sells baseline weapons, armor, and supplies
- Temple Acolyte
  Restores HP/MP and removes status effects
- Master Blacksmith
  Repairs and enhances equipment
- Arena Steward
  Manages signup, schedule, and bracket information
- Warehouse Keeper
  Handles item storage and retrieval

### Buildings And Facilities

- Adventurers Guild
- Weapon Shop
- Armor Shop
- Temple
- Blacksmith
- Arena Hall
- Warehouse

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

- low-rank supply loops
- healer and consumable traffic
- contract handoff frequency
- early material courier flow

### Interactive NPCs

- Outpost Contract Officer
  Offers low-rank field and delivery contracts
- General Goods Trader
  Sells baseline supplies and recovery items
- Field Medic
  Handles fast healing, cleanse, and low-risk treatment
- Caravan Dispatcher
  Connects to supply and escort quests

### Buildings And Facilities

- Quest Outpost
- General Store
- Field Healer

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
- teams building stable low-rank progression loops

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

- type: low-rank dungeon
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

Sunscar Frontier is the dividing line where the world shifts from familiar hunting lanes into a harsher mid-rank frontier.

### Observation Focus

- migration of mid-rank parties
- elite contract completion rate
- long-distance travel traffic
- formation of higher-risk, higher-yield loops

### Interactive NPCs

- Frontier Bounty Officer
  Provides elite field contracts
- Desert Supplier
  Sells mid-rank recovery and resistance goods
- Ruin Surveyor
  Connects to rare materials and deeper frontier hooks
- Escort Liaison
  Manages transport, interception, and escort contracts

### Buildings And Facilities

- Frontier Contract Post
- Desert Supply Stall
- Ruin Survey Camp

### Dungeon Relationship

- parent field region for `sandworm_den`

### Main Material Outputs

- Sunscorched Ore
- Dry Resin
- Sand Carapace
- Dust Crystal
- Venom Sac

### Progression Use

- mid-tier weapon enhancement
- anti-poison and durability gear lines
- advanced quest fulfillment materials
- critical prerequisite resources before high-tier dungeon play

## 3.6 `sandworm_den`

### Public Display Name

- English: Sandworm Den
- Chinese: 沙虫巢穴

### One-line Background

Sandworm Den is one of the most dangerous and watchable dungeons in V1, serving as a power threshold for high-rank bots.

### Observation Focus

- high-rank entries and defeats
- rare clear appearances
- boss fight heat
- material monopolies by top-performing bots

### Interactive NPCs

- Deep Den Watcher
  Handles entry permissions and warnings
- Venom Alchemist
  Purchases poison and worm-derived materials
- Elite Appraiser
  Evaluates rare drops

### Buildings And Facilities

- Den Entrance Camp
- Venom Lab Cart
- Elite Loot Exchange

### Dungeon Definition

- type: high-rank dungeon
- pacing: 5 encounters plus matriarch boss
- risk: high physical pressure with medium-high poison damage

### Main Material Outputs

- Sandworm Fang
- Hardened Carapace Plate
- Toxic Gland
- Deep Desert Core
- Matriarch Spine Shard

### Progression Use

- high-tier equipment enhancement
- rare weapon and armor crafting
- poison and penetration-oriented build progression
- endgame-zone and arena gearing prerequisites

## 4. Map Node State Definition

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

This helps a human observer immediately answer:

- what place is this
- what is happening there right now
- why are bots going there
- why do the materials from there matter

## 5. Direct Frontend Implication

In the next homepage map redesign, nodes should evolve into place-status cards.

Minimum visible content:

- place name
- one-line background
- current main activity
- activity heat
- signature material

Expanded or selected state:

- NPCs
- buildings and facilities
- dungeon relationship
- progression use

## 6. Recommended Next Steps

Once this catalog is accepted, frontend work should continue in three directions:

1. Turn homepage nodes into place cards rather than small buttons
2. Add `NPC / facilities / material output / dungeon relationship` sections to region detail pages
3. Reserve read-model fields such as `primary_activity`, `material_focus`, `risk_level`, and `landmark_key`
