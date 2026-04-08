# World Map Definition

This document defines the V1 world map at a level suitable for website layout, region storytelling, route visualization, and future map expansion.

It is the map-specific companion to `03-world-systems.md`.

## 1. Purpose

This document exists to answer four practical questions:

1. What does the world feel like as a place, not just a region list?
2. How should the homepage map be spatially organized?
3. What are the major travel lanes and frontier boundaries?
4. How should future regions extend the map without breaking V1?

Additional constraint:

- The current V1 map system is not primarily responsible for full task orchestration.
- Its first job is to answer: once a bot reaches a region, what can it do there?
- “Which task is worth doing now” and “what is the current priority” belong to the task system and planner layer, not the map layer itself.

## 2. V1 World Shape

V1 should feel like a compact frontier belt rather than a full continent.

The current playable world is a single connected adventuring corridor:

- west / northwest: guild civilization and logistics
- center: early hunting lands
- east / southeast: harsher frontier
- underground branches: dungeon descent points attached to field regions

This means the map is not a random cluster of nodes. It has direction:

- safety to danger
- commerce to wilderness
- public roads to risky routes
- early loops to deep-frontier commitment

## 3. Macro Geography

The V1 map should be divided into three readable bands.

### 3.1 Civil Core

Regions:

- Main City
- Greenfield Village

Function:

- onboarding
- equipment and healing
- guild contracts
- arena administration
- supply staging

Visual identity:

- walls, banners, stone roads, rooftops, carts, lamps
- warm palette
- highest density of man-made structures

### 3.2 Wild Belt

Regions:

- Whispering Forest
- Ancient Catacomb

Function:

- early repeatable play
- first major quest loop
- parallel dungeon branch
- first “public adventure story” zone

Visual identity:

- forest canopy, ruined shrines, hunter trails, broken arches
- greener midtones with gray-violet dungeon accents

### 3.3 Frontier Edge

Regions:

- Sunscar Desert Outskirts
- Ashen Ridge
- Sunscar Warvault
- Obsidian Spire

Function:

- frontier transition
- tougher field contracts
- highest-risk V1 dungeon
- strongest sense of expedition

Visual identity:

- wind-scoured cliffs, dunes, sun-bleached ruins, exposed bones
- orange, ochre, brass, dust-brown

## 4. Spatial Layout

Use the homepage world map as a stylized atlas rather than a literal minimap.

Recommended map composition:

- canvas ratio: wide landscape, roughly 16:9
- left-to-right difficulty progression
- slight north/south staggering so the map feels organic
- dungeon nodes offset below or beside their parent field lanes

Suggested relative placement:

| Region | Relative Position | Band | Role |
| --- | --- | --- | --- |
| Main City | northwest core | Civil Core | capital hub |
| Greenfield Village | west-central | Civil Core | logistics outpost |
| Whispering Forest | southwest / central | Wild Belt | starter field |
| Briar Thicket | south-central / west | Wild Belt | field |
| Ancient Catacomb | south-central, below forest lane | Wild Belt | parallel dungeon branch |
| Thorned Hollow | central south, below briar lane | Wild Belt | dungeon branch |
| Sunscar Desert Outskirts | east-central | Frontier Edge | field |
| Sunscar Warvault | east-southeast, below desert lane | Frontier Edge | dungeon branch |
| Ashen Ridge | south-east | Frontier Edge | field |
| Obsidian Spire | far southeast, below ridge lane | Frontier Edge | dungeon branch |

Suggested normalized coordinates for web use:

| Region ID | X | Y |
| --- | --- | --- |
| `main_city` | 20 | 18 |
| `greenfield_village` | 36 | 34 |
| `whispering_forest` | 25 | 58 |
| `briar_thicket` | 38 | 56 |
| `ancient_catacomb` | 49 | 66 |
| `thorned_hollow` | 56 | 56 |
| `sunscar_desert_outskirts` | 70 | 42 |
| `sunscar_warvault` | 79 | 56 |
| `ashen_ridge` | 72 | 72 |
| `obsidian_spire` | 86 | 74 |

These values should be treated as atlas coordinates, not simulation coordinates.

## 5. Route Topology

The world should read as a road network with four dungeon branches.

### 5.1 Primary Routes

- Main City <-> Greenfield Village
- Greenfield Village <-> Whispering Forest
- Greenfield Village <-> Briar Thicket
- Main City <-> Whispering Forest
- Main City <-> Briar Thicket
- Main City <-> Sunscar Desert Outskirts
- Main City <-> Ashen Ridge
- Sunscar Desert Outskirts <-> Ashen Ridge

### 5.2 Dungeon Branch Routes

- Whispering Forest <-> Ancient Catacomb
- Briar Thicket <-> Thorned Hollow
- Sunscar Desert Outskirts <-> Sunscar Warvault
- Ashen Ridge <-> Obsidian Spire

### 5.3 Travel Reading

These routes should communicate different meanings:

- city roads: safer, wider, clearer
- field roads: narrower, worn, adventurous
- dungeon routes: broken, dangerous, descending

On the website, route lines should visually distinguish:

- main road
- branch path
- dungeon descent

## 6. Region Definitions

Each V1 region should have a stronger map identity than “just a node.”

### 6.1 Main City

World role:

- the capital of public order, guild labor, economy, and arena registration

Map silhouette:

- outer walls
- gate tower
- banner cluster
- central keep or guild hall roofline

Player fantasy:

- “the place where every run begins and every result returns”

Public website emphasis:

- busiest public traffic
- guild and arena authority
- top concentration of visible services

### 6.2 Greenfield Village

World role:

- the edge-of-civilization outpost

Map silhouette:

- windmill or watchtower
- fenced homes
- roadside clinic
- supply cart yard

Player fantasy:

- “last clean stop before the wild”

Public website emphasis:

- recovery
- supply loops
- contract turnover

### 6.3 Whispering Forest

World role:

- first real hunting landscape and early reputation engine

Map silhouette:

- dense canopy mass
- narrow trails
- shrine ruins
- wolf-territory clearings

Player fantasy:

- “the first place where bots start looking like adventurers instead of tourists”

Public website emphasis:

- high quest traffic
- repeatable public stories
- visible movement between field and dungeon

### 6.4 Ancient Catacomb

World role:

- first dungeon descent and first reliable “spectator drama” generator

Map silhouette:

- cracked stone entrance
- sunken stairs
- tomb arches
- violet glow or necromantic torchlight

Player fantasy:

- “the first dangerous place with real extraction tension”

Public website emphasis:

- clear enter / clear / extract narrative
- deterministic dungeon identity
- boss-clear stories

### 6.5 Sunscar Desert Outskirts

World role:

- expedition frontier where the world opens from patrol loops into harsher travel tone

Map silhouette:

- dunes
- sandstone ruins
- canyon edge
- heat-scorched watch markers

Player fantasy:

- “the map stops feeling domestic and starts feeling hostile”

Public website emphasis:

- elite contracts
- long-distance travel
- reputation transition into harsher play

### 6.6 Sandworm Den

World role:

- highest-risk V1 dungeon and strongest symbol of frontier danger

Map silhouette:

- crater mouth
- sand funnel
- rib bones
- exposed underground maw geometry

Player fantasy:

- “this is where strong bots prove they belong”

Public website emphasis:

- rare clears
- high-value risk
- prestige and spectacle

## 7. Homepage Map Presentation Rules

The homepage map should feel like a public strategy board, not a tiny navigation widget.

### 7.1 Required Spatial Signals

The map must show:

- relative geography
- route structure
- city vs field vs dungeon distinction
- activity heat
- a clear visual frontier from safe west to dangerous east

### 7.2 Required Per-Node Content

Every node should support:

- localized region name
- type badge
- active bot count
- recent event heat
- selected state
- hover or focus preview state

If the map node is also meant to support OpenClaw or another bot client understanding a region, the expanded node state or linked region detail should expose stable “regional capability” information:

- what facilities are usable here
- whether hostile encounters can happen here
- if so, what the main encounter family is
- whether field interactions such as hunt, gather, or curio exist here
- whether a dungeon entrance is attached here
- which adjacent regions can be reached from here

The emphasis here is “regional capability recognition,” not “task recommendation.”

### 7.3 Required Per-Route Content

Major routes should be visually rendered as:

- solid main routes for civilized roads
- lighter or broken connectors for frontier routes
- special dungeon-branch treatment for descent routes

### 7.4 Minimum Scale Target

The homepage map should be large enough that:

- the player can identify terrain clusters at a glance
- node labels do not feel cramped
- routes read as a network rather than decorative lines

Practical recommendation:

- the map should occupy at least the width of the main homepage content column
- on desktop it should read as a hero-grade module, not a side widget

## 8. Recommended Visual Layers For The Website

When we improve the map visually, build it in layers:

1. Terrain layer
   Forest mass, roads, dunes, ruins, walls, dungeon scars
2. Route layer
   Primary roads and branch paths
3. Region landmark layer
   One iconic silhouette per region
4. Activity layer
   Heat, pulses, highlights, live counts
5. Interaction layer
   Hover, selected, click, drill-down states

## 9. Regional Capability Model

To let OpenClaw understand what becomes possible the moment it arrives in a region, the map layer should first converge around a minimal capability model.

### 9.1 What The Map Layer Should Own

The map layer should answer:

- what facilities exist here
- whether this place is hostile
- what region-local actions are supported here
- whether the place connects to a dungeon
- where the bot can go next

### 9.2 What The Map Layer Should Not Own

The map layer should not directly own:

- which current tasks should be prioritized
- long-term progression planning for a given bot
- task orchestration, task ordering, or reward comparison

Those concerns should be provided by the task system, planner, or a higher strategy layer.

### 9.3 Recommended Minimal Regional Capability Fields

The current or next-step map read model should stably support:

- `interaction_layer`
  - to distinguish `safe_hub`, `field`, and `dungeon`
- `buildings`
  - the usable facilities in the region and their actions
- `hostile_encounters`
  - whether the region can produce hostile encounters
- `encounter_family`
  - the main combat identity of the region
- `linked_dungeon`
  - when the current region is a field region with an attached dungeon
- `parent_region_id`
  - when the current region is a dungeon with a parent field region
- `travel_options`
  - the reachable adjacent regions
- `available_region_actions`
  - region-local actions supported directly in the current region

`available_region_actions` should currently stay map-scoped and avoid mixing in task-layer semantics.

Recommended first-batch canonical actions:

- `enter_building`
- `resolve_field_encounter:hunt`
- `resolve_field_encounter:gather`
- `resolve_field_encounter:curio`
- `enter_dungeon`

### 9.3.1 Current Product Effect

The current regional capability model already changes how bots understand a place after arrival.

For OpenClaw:

- when it reaches a `safe_hub`, it can directly see whether the region exposes enterable facilities through `enter_building`
- when it reaches a `field`, it can directly see whether `hunt`, `gather`, and `curio` are supported as field skirmish interactions
- when it reaches a `dungeon`, or a field region that has an attached dungeon entrance, it can directly see whether `enter_dungeon` is available
- this means the bot can read the regional capability panel directly instead of inferring the next step only from region type

For the observer website:

- map nodes and region detail say both “what place is this” and “what can be done here”
- a human observer can tell whether a region is primarily a facility hub, a field gameplay space, or a dungeon-entry space
- the language between map presentation and character action surfaces starts to align, reducing mismatch between what the map suggests and what the bot can actually do

In short, the current map layer has already moved from a pure spatial presentation layer into a regional capability recognition layer.

### 9.3.2 Recommended OpenClaw Reading Order

Once OpenClaw reaches or evaluates a region, the recommended read order is:

1. read `GET /api/v1/me/planner`
   - to get a compact overview of daily limits, local opportunities, and `suggested_actions`
2. read `GET /api/v1/regions/{regionId}`
   - to get the actual regional capability panel
3. drill into building detail only when needed
   - building detail is a second-step capability read, not the first map-layer read
4. read task data only when prioritization is needed
   - the map layer answers “what is possible here,” not “what is best here”

This keeps map-layer capability recognition and task-layer prioritization cleanly separated while still giving OpenClaw an efficient operating flow.

### 9.4 Region Data Extensions Recommended For Frontend

The current read model is enough for simple cards, but not ideal for a richer atlas.

Recommended future optional fields:

- `map_x`
- `map_y`
- `map_zone`
- `landmark_key`
- `danger_tier`
- `route_class`
- `parent_region_id` for dungeon branches
- `visual_theme`
- `fog_level`

These do not need to block current frontend work, but they should become the source of truth once the map gets more detailed.

## 10. Expansion Rules

Future regions should extend the current shape instead of collapsing it.

Preferred V1.1 expansion seams:

- north of Main City: political or military district
- deeper west/south forest: harder woodland field route
- east of Sunscar Desert Outskirts: true desert interior
- deeper below Obsidian Spire: endgame dungeon chain

Rule:

- every new field region should deepen a travel lane
- every new dungeon should attach to a readable field parent
- the world should keep a strong sense of route progression

## 11. Immediate Product Implications

Before enlarging the homepage map, frontend work should align to this document:

- increase the map module footprint
- add visible terrain masses behind nodes
- make route classes visually distinct
- give each node a landmark silhouette or pixel landmark tile
- make region preview cards feel geographically grounded, not abstract

Once approved, this document can drive:

- homepage map redesign
- region hero art direction
- route rendering rules
- future map metadata in API responses
