## 9. Equipment System

### 9.1 Equipment slots

- head
- chest
- necklace
- ring
- boots
- weapon

### 9.2 Equipment rules

- only one item per slot
- items are bound to the adventurer account in V1
- equipping or unequipping is out-of-combat only
- weapon type must match class-compatible weapon families

### 9.3 Item rarity

- Common
- Rare
- Epic

V1 item power should come mostly from:

- main stat package
- one passive affix at most

Examples of passive affixes:

- `+max_hp`
- `+physical_attack`
- `+magic_attack`
- `+healing_power`
- `+speed`
- `+physical_defense`
- `+magic_defense`

No proc-based or on-hit affixes in V1.

### 9.4 Starter gear

Every new adventurer receives:

- class-compatible starter weapon
- cloth or armor chest item based on class
- basic boots
- 100 starting gold

## 10. Economy

### 10.1 Currency

V1 uses one soft currency:

- `gold`

### 10.2 Gold sources

- guild quest rewards
- dungeon clear rewards
- dungeon loot sold to shops
- arena weekly rewards

### 10.3 Gold sinks

- consumables
- equipment repair fee after dungeon or arena defeat
- equipment enhancement
- fast travel fee between distant regions
- guild quest reroll fee

### 10.4 Enhancement

V1 enhancement is intentionally simple:

- only weapons and chest items can be enhanced
- enhancement levels: `+0` to `+5`
- enhancement never destroys the item
- enhancement cost scales by rarity and level
- failure only consumes gold and materials

Reason:

- low emotional volatility
- easier economy tuning

## 11. World Map

### 11.1 Region list

V1 regions:

- Main City
- Greenfield Village
- Whispering Forest
- Sunscar Desert Outskirts
- Ancient Catacomb
- Sandworm Den

### 11.2 Region unlocks

| Region | Access Requirement | Type |
| --- | --- | --- |
| Main City | default | safe hub |
| Greenfield Village | default | safe hub |
| Whispering Forest | default | field |
| Ancient Catacomb | default | dungeon |
| Sunscar Desert Outskirts | Mid rank | field |
| Sandworm Den | High rank | dungeon |

### 11.3 Travel rules

- travel is menu-based, not free-roam
- travel consumes no time but may consume gold for long-distance fast travel
- all regions expose a list of interactable facilities and available actions

## 12. Buildings and Interactions

### 12.1 Main City

- Adventurers Guild
- Weapon Shop
- Armor Shop
- Temple
- Blacksmith
- Arena Hall
- Warehouse

### 12.2 Greenfield Village

- Quest Outpost
- General Store
- Field Healer

### 12.3 Building actions

Adventurers Guild:

- list quests
- accept quest
- submit quest
- reroll daily board for gold

Weapon Shop / Armor Shop:

- browse stock
- buy item
- sell loot

Temple / Field Healer:

- restore HP/MP for gold
- remove status effects

Blacksmith:

- repair item durability
- enhance eligible equipment

Warehouse:

- list inventory
- equip item
- unequip item

Arena Hall:

- view schedule
- sign up
- view bracket

## 13. Guild Quest System

### 13.1 Quest board structure

At daily reset, each adventurer receives a personal quest board.

The board contains:

- 3 common quests
- 2 uncommon quests
- 1 challenge quest

### 13.2 Quest types

V1 templates:

- defeat `N` enemies in a region
- defeat a named elite in a dungeon
- collect `N` materials from a region encounter pool
- deliver purchased supplies to an outpost
- clear a dungeon without defeat

### 13.3 Quest constraints

- a quest can be active or completed once per daily board
- abandoned quests count against the daily completion cap only if already completed
- rerolling replaces all incomplete quests on the board

### 13.4 Quest rewards

Every quest grants:

- gold
- reputation

Challenge quests may additionally grant:

- enhancement materials
- guaranteed Rare item

## 14. Dungeon System

### 14.1 V1 dungeons

#### Ancient Catacomb

- access: Low rank
- theme: undead / dark magic
- floors: 3 encounters plus boss
- damage profile: mixed physical and magic

#### Sandworm Den

- access: High rank
- theme: desert beast / poison
- floors: 4 encounters plus boss
- damage profile: physical and poison pressure

### 14.2 Entry rules

- each entry consumes one daily dungeon charge
- entry is blocked when no charge remains
- abandoning a run still consumes the charge

### 14.3 Dungeon rewards

On successful clear:

- clear gold
- loot table roll
- boss drop roll
- possible reputation bonus if linked to quest

On failure:

- partial loot only if at least one encounter was cleared
- repair fee increases on damaged items

## 15. Arena System

### 15.1 Arena eligibility

- Mid rank and above can sign up
- signup closes Saturday `19:50` Asia/Shanghai

### 15.2 Format

- single-elimination tournament
- bracket seeding uses:
  1. adventurer rank
  2. current equipment score
  3. registration timestamp

### 15.3 Match rules

- arena uses the same battle engine as PvE
- all matches are fully simulated by the server
- no manual intervention after signup

### 15.4 Rewards

- top 1, 2, 4, 8 receive gold and unique title strings
- rankings page stores the latest completed tournament snapshot

### 15.5 V1 limitations

- no betting
- no live tactical input
- no replay UI beyond event log and battle summary

