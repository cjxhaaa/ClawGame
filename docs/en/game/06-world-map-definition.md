# World Map Definition

This document defines the V1 world map at a level suitable for website layout, region storytelling, route visualization, and future map expansion.

It is the map-specific companion to `03-world-systems.md`.

## 1. Purpose

This document exists to answer four practical questions:

1. What does the world feel like as a place, not just a region list?
2. How should the homepage map be spatially organized?
3. What are the major travel lanes and frontier boundaries?
4. How should future regions extend the map without breaking V1?

## 2. V1 World Shape

V1 should feel like a compact frontier belt rather than a full continent.

The current playable world is a single connected adventuring corridor:

- west / northwest: guild civilization and logistics
- center: low-rank hunting lands
- east / southeast: harsher mid-rank frontier
- underground branches: dungeon descent points attached to field regions

This means the map is not a random cluster of nodes. It has direction:

- safety to danger
- commerce to wilderness
- public roads to risky routes
- low-rank loops to high-rank commitment

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

- low-rank repeatable play
- first major quest loop
- first dungeon branch
- first “public adventure story” zone

Visual identity:

- forest canopy, ruined shrines, hunter trails, broken arches
- greener midtones with gray-violet dungeon accents

### 3.3 Frontier Edge

Regions:

- Sunscar Desert Outskirts
- Sandworm Den

Function:

- mid/high-rank transition
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
| Ancient Catacomb | south-central, below forest lane | Wild Belt | starter dungeon branch |
| Sunscar Desert Outskirts | east-central | Frontier Edge | mid-rank field |
| Sandworm Den | southeast, below desert lane | Frontier Edge | high-rank dungeon branch |

Suggested normalized coordinates for web use:

| Region ID | X | Y |
| --- | --- | --- |
| `main_city` | 20 | 18 |
| `greenfield_village` | 36 | 34 |
| `whispering_forest` | 25 | 58 |
| `ancient_catacomb` | 49 | 66 |
| `sunscar_desert_outskirts` | 70 | 42 |
| `sandworm_den` | 82 | 70 |

These values should be treated as atlas coordinates, not simulation coordinates.

## 5. Route Topology

The world should read as a road network with two dungeon branches.

### 5.1 Primary Routes

- Main City <-> Greenfield Village
- Greenfield Village <-> Whispering Forest
- Main City <-> Whispering Forest
- Main City <-> Sunscar Desert Outskirts
- Whispering Forest <-> Sunscar Desert Outskirts

### 5.2 Dungeon Branch Routes

- Whispering Forest <-> Ancient Catacomb
- Sunscar Desert Outskirts <-> Sandworm Den

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

- first real hunting landscape and low-rank reputation engine

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

- mid-rank frontier where the world opens from patrol loops into expedition tone

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

## 9. Region Data Extensions Recommended For Frontend

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
- deeper below Sandworm Den: endgame dungeon chain

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
