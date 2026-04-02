## 9. Equipment System

This section is now a high-level summary only.

The detailed specification has moved to:

- `09-equipment-dungeon-and-loot-framework.md`

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

The baseline rarity direction is now superseded by the new five-grade dungeon loot model:

- Blue
- Purple
- Gold
- Red
- Prismatic

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

- restore HP for gold
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

- 3 `normal` quests
- 2 `hard` quests
- 1 `nightmare` quest
- the current repo still uses `common / uncommon / challenge` as the temporary board-pool split and should later migrate to an explicit `difficulty`
- the business-day reset happens at `04:00 Asia/Shanghai`
- quest states currently include `available`, `accepted`, `completed`, `submitted`, and `expired`
- `GET /me/quests` returns the whole active board rather than a paged quest list

### 13.2 Quest difficulty

Quest difficulty should be a separate axis from quest rarity or quest source.

The three planned tiers are:

- `normal`
- `hard`
- `nightmare`

Their intended roles:

- `normal`: one region, one objective, one loop; suitable for stable daily throughput
- `hard`: cross-region, cross-facility, or multi-step but still rule-driven
- `nightmare`: requires information interpretation and a small story-like procedure rather than a single combat check

The increase in difficulty should come not only from combat numbers but also from procedural complexity:

- `normal` is direct action
- `hard` is combined action
- `nightmare` is a conditional mini-scenario

### 13.3 Quest types

V1 templates:

- defeat `N` enemies in a region
- defeat a named elite in a dungeon
- collect `N` materials from a region encounter pool
- deliver purchased supplies to an outpost
- clear a dungeon without defeat

Current implementation notes:

- field curios can also generate a `curio_followup_delivery` quest
- this follow-up quest does not consume one of the fixed daily template slots
- when generated, it is inserted into the board and auto-accepted immediately

Quest content should not be restricted to pure combat. The daily pool should also cover:

- combat clearing
- gathering and recovery
- transport and delivery
- investigation and evidence collection
- story reasoning
- multi-step handoff flows

### 13.4 Daily quest pool planning

#### `normal` daily quests

Role:

- give OpenClaw low-overhead, reliable daily loops

Suitable content:

- defeat a set number of enemies in one region
- gather a set number of materials in one region
- deliver standard supplies to one outpost
- visit one facility and finish a standard interaction

Examples:

- `Clear Forest Route`: defeat a target count of enemies in `whispering_forest`
- `Gather Shrine Herbs`: collect a target amount of materials in `whispering_forest`
- `Deliver Village Provisions`: move supplies from `main_city` to `greenfield_village`
- `Repair Outpost Gear`: visit a smith in a target settlement and finish a repair order

#### `hard` daily quests

Role:

- make the bot handle cross-region and cross-step tasks while keeping the logic explicit

Suitable content:

- obtain an item first, then deliver it
- complete a handoff between two facilities
- recover a clue in the field and report back to a hub
- clear a dungeon or elite objective and then return for hand-in

Examples:

- `Reinforce the Forest Outpost`: pick up supplies in the city, then deliver them to the forest outpost
- `Desert Contract Relay`: obtain the frontier receipt in the desert, then register it back in the city
- `Catacomb Threat Report`: clear the `ancient_catacomb` objective, then return to the guild with the report
- `Recovered Caravan Ledger`: gain a ledger from a field `curio`, then verify it at the village outpost

#### `nightmare` daily quests

Role:

- make OpenClaw solve daily missions that include story interpretation and procedural judgment
- the point is not just bigger monsters, but small executable mystery-like flows

Suitable content:

- infer which NPC or building should receive an item based on clues
- visit several locations in order to reconstruct what happened
- handle branch conditions where the wrong step leads to partial completion or retry
- read log or description text to infer the next action

Examples:

- `Echoes Beneath the Shrine`:
- recover torn pages from a forest `curio`
- collect testimony in `greenfield_village`
- decide whether the final evidence should be handed to the guild or the temple

- `Beacon Without a Sender`:
- recover a damaged beacon core in the desert
- decode it in the city workshop
- decide whether to redirect supplies to the frontier or archive the incident

- `The Wrong Crate`:
- verify crate numbers in storage
- confirm the missing shipment at an outpost
- infer which cargo batch was swapped and deliver it to the correct destination

Key requirements for `nightmare` dailies:

- at least 3 steps
- at least 1 step that requires a text-clue judgment
- at least 1 cross-region or cross-facility move
- completion should depend on the correct procedure, not on a single combat result

### 13.5 Quest constraints

- a quest can be active or completed once per daily board
- only `available` quests can be accepted in the current implementation
- quest state must advance through its defined state machine and cannot skip to submit
- rerolling replaces all incomplete quests on the board
- `submitted` and `completed` quests are preserved across reroll
- other non-finished quests are marked `expired` before replacement quests are appended

Additional process constraint for future daily quests:

- completion should not depend only on “entering a region”; it must also support step, clue, and handoff target validation

### 13.6 Quest rewards

Every quest grants:

- gold
- reputation

Challenge quests may additionally grant:

- enhancement materials
- guaranteed Rare item

Current implementation notes:

- the stable reward loop today is gold plus reputation
- submitting a quest can immediately rank the character up when the reputation threshold is crossed
- challenge-specific bonus item rewards are not yet a fully implemented loop
- reward values should stay open for now and be filled in after the broader economy and progression numbers are locked

### 13.7 Current progression triggers

The quest system is already wired into the three main gameplay loops:

- `deliver_supplies` and `curio_followup_delivery` complete automatically when travel reaches the target region
- `kill_region_enemies` and `collect_materials` progress from resolved field encounters using enemy and material counts
- `kill_dungeon_elite` and `clear_dungeon` complete when the matching dungeon run resolves successfully

This keeps map and quest responsibilities intentionally separate:

- the map tells OpenClaw what actions are available in the current region
- the quest system decides which accepted objectives advance after those actions resolve

### 13.8 Current scope and known gaps

The current quest system is intentionally basic and aims to provide a stable growth loop rather than a full narrative framework.

Already supported:

- personal daily quest boards
- accept, complete, submit, and reroll flow
- automatic progress updates from travel, field, and dungeon resolution
- curio-seeded delivery follow-up quests

Not yet fully supported:

- multi-stage quest chains
- explicit prerequisite trees
- region-level quest priority recommendation
- a dedicated strategy layer for quest planning
- an explicit `quest_difficulty`
- text-clue judgment and multi-step state handling for `nightmare` daily content

### 13.9 Quest framework design principles

To keep future quest additions cheap, the system should evolve into a template-driven framework.

Core principles:

- new quests should mostly be added by defining templates, step structures, and content data
- all quest types should share one state machine instead of inventing a new flow per quest family
- quest progression should be driven by standard triggers rather than scattered custom logic inside travel, field, and dungeon handlers
- `nightmare` quests must support multi-step progression, branching, clue interpretation, and incorrect paths
- the map answers “what can be done here”; the quest system decides whether that action advances a quest

Recommended layering:

- template layer: defines the quest goal, step structure, difficulty, and generation rules
- instance layer: records the concrete quest rolled for one character
- runtime layer: records current step, discovered clues, completed nodes, and branch choices

### 13.10 Recommended quest-step model

The framework should support more than just numeric progress bars. Step types should include:

- arrive at a region
- interact with a building
- resolve a specific field approach
- finish a dungeon objective or dungeon segment
- deliver an item or evidence object
- discover or read a clue
- submit a branch choice
- return to a target NPC or building with a conclusion

For `normal` quests, 1 to 2 steps are usually enough.

For `hard` quests, 2 to 3 steps are the typical target.

For `nightmare` quests, the recommended baseline is:

- 3 or more steps
- at least 1 clue-discovery step
- at least 1 choice or inference step
- incorrect procedures are allowed, but they should not break the system

### 13.11 Recommended quest generation rules

Daily boards should be generated from a fixed framework rather than pure randomness:

- fill the board by fixed `normal / hard / nightmare` slots first
- then apply per-tag limits and de-duplication inside each difficulty band
- one board should not be overloaded with the same content type such as three pure combat tasks
- `nightmare` tasks should be drawn from templates that support reasoning, branching, and narrative recovery

Recommended tags:

- `combat`
- `gather`
- `delivery`
- `investigation`
- `handoff`
- `dungeon`
- `story`

Goal:

- each day should give OpenClaw both stable loops and a small amount of higher-cognition content
- adding new quests should mostly mean “add templates”, not “rewrite the framework”

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
- signup stays open each day until `09:00` Asia/Shanghai

### 15.2 Format

- daily qualifier cycle
- signup locks at `09:00`, after which all signed entrants enter the qualifier pool
- qualifier rounds are resolved automatically in repeated 1v1 elimination waves until the live field becomes a 64-player main bracket
- if a qualifier round has an odd entrant count, the bracket may assign a deterministic bye so the round can still complete cleanly
- if signups are below `64`, pre-seeded NPC entrants are added until the main bracket reaches 64
- NPC strength is based on the median power band of the signed-up entrants
- registration ordering is only used as a stable display tiebreaker, not as a way to discard extra entrants

### 15.3 Match rules

- arena uses the same battle engine as PvE
- qualifier duels are fully simulated by the server
- every qualifier duel and every main-bracket duel produces a battle report
- battle reports are queryable from both arena tournament views and the participating bot's own arena battle history
- once the 64-player bracket begins, each elimination round resolves every `5 minutes`
- the final resolves after the bracket schedule completes, after which the champion is published
- no manual intervention after signup

### 15.4 Rewards

- top 1, 2, 4, 8 receive gold and unique title strings
- rankings page stores the latest completed tournament snapshot

### 15.5 V1 limitations

- no betting
- no live tactical input
- no replay UI beyond event log and battle summary
