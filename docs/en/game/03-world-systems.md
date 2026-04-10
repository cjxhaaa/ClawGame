## 9. Equipment System

Module scope:

- this chapter keeps the player-facing world and progression summary
- detailed numeric systems should live in the dedicated deep-dive modules under `06-19`
- backend contracts, API surface, and website delivery concerns should stay in `04-bot-platform.md` and `05-website-ops-and-delivery.md`

This section is a high-level summary.

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
- civilians may equip any weapon or armor piece; their restriction is skill access rather than equipment access
- after profession choice, weapon type must match the chosen class-compatible weapon families
- if a profession change makes the currently equipped weapon incompatible, the weapon is automatically unequipped to inventory

### 9.3 Item rarity

V1 item rarity uses the five-grade dungeon loot model:

- Blue
- Purple
- Gold
- Red
- Prismatic

### 9.4 Starter gear

Every new adventurer receives:

- basic civilian chest item
- basic boots
- 100 starting gold
- no profession starter weapon at creation

When a character reaches level `10`, the Adventurers Guild may offer profession changes among `civilian`, `warrior`, `mage`, and `priest`. Each change costs `800` gold. Moving from `civilian` into a promoted class grants one class-aligned starter weapon.

## 10. Economy

Boundary note:

- this chapter explains economy intent and player-facing rules
- pricing formulas, payout models, and backend persistence details should not be duplicated here when a dedicated module already exists

### 10.1 Currency

V1 uses one soft currency:

- `gold`

### 10.2 Gold sources

- guild quest rewards
- dungeon clear rewards
- dungeon loot sold to shops
- arena weekly rewards
- arena prize-pool payouts
- arena betting payouts

### 10.3 Gold sinks

- consumables
- equipment enhancement
- skill upgrades
- fast travel fee between distant regions
- guild quest reroll fee
- arena betting stake

### 10.4 Enhancement

V1 enhancement is intentionally simple:

- every equipment slot can be enhanced
- enhancement levels: `+0` to `+20`
- enhancement never destroys the item
- enhancement consumes gold and enhancement materials
- enhancement success is deterministic in V1
- enhancement is bound to equipment slots rather than individual item instances
- changing to another item in the same slot keeps that slot's enhancement level
- enhancement only scales the equipped item's base stat package for that slot
- passive affixes are not multiplied by enhancement

Reason:

- low emotional volatility
- easier economy tuning

### 10.5 Arena weekly cycle and rewards

The arena should use a weekly loop built from weekday rating play, a Saturday elimination bracket, and temporary weekly title rewards.

Weekly cycle:

- Monday to Friday: rating qualifier ladder
- Saturday: top-64 elimination tournament
- Sunday: rest day, result presentation, and title payout

Qualifier ladder rules:

- every Bot starts with `1000` arena rating
- every Bot has `3` free rating challenges per day
- extra challenges may be purchased with gold
- each Bot may buy up to `10` extra challenges per day
- the gold cost for extra challenges should increase with each purchase
- the candidate list should show `5` randomly selected opponents from a nearby rating band
- each rating challenge auto-resolves under a `10`-round cap
- if the round cap is reached with both sides still alive, the challenger loses; rating challenges do not draw

Rating change rules:

- when the challenger wins, the challenger gains rating and the defender loses rating
- when the challenger loses, the challenger loses nothing and the defender remains unchanged
- one successful challenge should transfer between `10` and `30` rating points
- upsetting a stronger opponent should land near the upper end
- beating a clearly weaker opponent should land near the lower end

Saturday elimination qualification:

- after Friday closes, the top `64` by rating qualify
- if the live field is below `64`, NPC entrants fill the remaining slots
- the full top-64 bracket resolves on Saturday
- each Saturday arena duel also uses a `10`-round cap
- if the round cap is reached with both duelists still alive, the duelist with higher remaining HP wins
- if remaining HP is also tied at the cap, use lower `character_id` as the stable tiebreaker so the bracket always advances

Weekly reward direction:

- weekly arena should not rely on large direct gold payouts as its primary reward
- title rewards begin at the top `32`
- each title lasts `2` weeks
- only one arena title may be active at a time
- a newly earned title overwrites the previous one and refreshes the duration

Title bands:

- top 32
- top 16
- top 8
- top 4
- runner-up
- champion

Title stat direction:

- the reward should mainly be a broad base-stat bonus
- primary base stats should receive the largest boost
- probability and outcome-modifier stats should scale more conservatively
- the champion title should be meaningfully stronger than runner-up without destabilizing dungeon or progression balance

Recommended title stat gradients:

- `top 32`
  - `max_hp / physical_attack / magic_attack / physical_defense / magic_defense / healing_power` `+2%`
  - `speed` `+1%`
  - `crit_rate` `+0.5%`
  - `crit_damage` `+1%`
  - `block_rate / precision / evasion_rate` `+0.4%`
  - `physical_mastery / magic_mastery` `+1%`
- `top 16`
  - primary base stats `+3%`
  - `speed` `+1.5%`
  - `crit_rate` `+0.8%`
  - `crit_damage` `+1.5%`
  - `block_rate / precision / evasion_rate` `+0.6%`
  - `physical_mastery / magic_mastery` `+1.5%`
- `top 8`
  - primary base stats `+4%`
  - `speed` `+2%`
  - `crit_rate` `+1.1%`
  - `crit_damage` `+2%`
  - `block_rate / precision / evasion_rate` `+0.8%`
  - `physical_mastery / magic_mastery` `+2%`
- `top 4`
  - primary base stats `+5%`
  - `speed` `+2.5%`
  - `crit_rate` `+1.4%`
  - `crit_damage` `+2.5%`
  - `block_rate / precision / evasion_rate` `+1%`
  - `physical_mastery / magic_mastery` `+2.5%`
- `runner-up`
  - primary base stats `+6.5%`
  - `speed` `+3%`
  - `crit_rate` `+1.8%`
  - `crit_damage` `+3%`
  - `block_rate / precision / evasion_rate` `+1.2%`
  - `physical_mastery / magic_mastery` `+3%`
- `champion`
  - primary base stats `+9%`
  - `speed` `+4%`
  - `crit_rate` `+2.5%`
  - `crit_damage` `+4%`
  - `block_rate / precision / evasion_rate` `+1.6%`
  - `physical_mastery / magic_mastery` `+4%`

## 11. World Map

Reading note:

- use this chapter for the top-level map, building, quest, dungeon, arena, and world-boss loop
- use `06-world-map-definition.md` and `07-location-catalog-and-resource-definition.md` when editing geography, route layout, regional identity, or place-by-place content

### 11.1 Region list

V1 regions:

- Main City
- Greenfield Village
- Whispering Forest
- Briar Thicket
- Sunscar Desert Outskirts
- Ashen Ridge
- Ancient Catacomb
- Thorned Hollow
- Sunscar Warvault
- Obsidian Spire

### 11.2 Region unlocks

| Region | Access Requirement | Type |
| --- | --- | --- |
| Main City | default | safe hub |
| Greenfield Village | default | safe hub |
| Whispering Forest | default | field |
| Briar Thicket | default | field |
| Ancient Catacomb | default | dungeon |
| Thorned Hollow | default | dungeon |
| Sunscar Desert Outskirts | default | field |
| Ashen Ridge | default | field |
| Sunscar Warvault | default | dungeon |
| Obsidian Spire | default | dungeon |

### 11.3 Travel rules

- travel is menu-based, not free-roam
- travel consumes no time but may consume gold for long-distance fast travel
- all regions expose a list of interactable facilities and available actions

## 12. Buildings and Interactions

Section role:

- this section explains which places and interaction surfaces exist in the world loop
- detailed formulas, reward tables, runtime state rules, and API contracts should live in the dedicated modules and backend specs

### 12.1 Main City

- `Adventurers Guild`: quests, profession changes, route planning, and skill progression
- `Equipment Shop`: equipment browsing, buying, and selling
- `Apothecary`: consumable purchase
- `Blacksmith`: enhancement and salvage
- `Arena Hall`: arena state, signup, bracket viewing, and betting entry
- `Warehouse`: storage-oriented utility surface when enabled

### 12.2 Greenfield Village

- `Adventurers Guild Outpost`: local quest and progression support
- `Equipment Shop`: early-region trading
- `Apothecary`: early sustain item access
- `Caravan Dispatch Point`: travel and delivery support

### 12.3 Building actions

- Guild surfaces handle quest listing, submission, reroll, profession changes, and growth planning.
- Shop surfaces handle browse, buy, and direct liquidation selling.
- Apothecary surfaces handle consumable purchase only.
- Blacksmith surfaces handle deterministic slot-bound enhancement and salvage.
- Arena surfaces handle schedule viewing, signup, bracket access, and betting.
- Warehouse surfaces are utility-facing, not a separate primary progression loop.

Reference modules:

- equipment and enhancement: `09-equipment-dungeon-and-loot-framework.md`
- bot/platform routing and backend rules: `04-bot-platform.md`

### 12.4 World-boss matching and rewards

- World boss is an asynchronous `6`-participant auto-resolved raid rather than a live co-op party feature.
- Matchmaking forms one shared boss attempt from the active queue while the current boss season window is open.
- Reward tiers are based on total team damage, not last hit or personal placement.
- Every valid participant in the same completed raid receives the same tier package.
- The system is meant to provide long-tail reforge progression without replacing dungeons as the main gear source.

Reference modules:

- reforge and equipment direction: `09-equipment-dungeon-and-loot-framework.md`
- bot/platform and backend behavior: `04-bot-platform.md` and backend specs

### 12.5 Arena betting

- Arena betting opens after Friday qualification is locked and the Saturday bracket is known.
- V1 only needs simple markets such as single-match winner and tournament champion.
- Betting is an optional spectator-facing `gold` sink, not a required progression loop.
- The goal is to make the public bracket more watchable without creating a deep prediction economy.

## 13. Guild Quest System

Section role:

- this section keeps the quest loop readable at a product level
- detailed template graphs, multi-step runtime rules, and persistence concerns should live in quest-specific docs and backend specs

### 13.1 Quest board structure

- Each adventurer maintains one personal daily contract board.
- The board should combine reliable daily progress with a smaller number of more valuable goals.
- Reset, top-up, and carry-over behavior should remain explicit and machine-readable.
- Exact generation weights and slot composition should live in quest config assets instead of being maintained here.

### 13.2 Quest difficulty

- The daily board should use a small set of difficulty bands so bots can trade off time, travel, and reward value.
- Lower-friction contracts sustain routine progress.
- Higher-friction contracts justify more travel, combat risk, or procedural effort.

### 13.3 Quest types

- Region combat
- Dungeon combat
- Material collection
- Delivery and travel
- Investigation and multi-step progression
- Optional follow-up tasks seeded from world interactions such as curios

### 13.4 Daily contract reward planning

- Rewards should be useful on the same day they are earned.
- Lower-friction quests mainly sustain gold and steady reputation.
- Harder quests justify travel, dungeon pressure, or multi-step reasoning with better returns.
- Daily contracts should complement dungeon farming, not replace it.

### 13.5 Quest constraints

- A quest should complete once per board rather than becoming infinitely repeatable.
- State transitions should stay explicit and machine-readable.
- Completion should support step, clue, and handoff validation instead of relying only on a location check.

### 13.6 Quest rewards

- Every quest grants gold and reputation.
- Harder quests may add stronger progression support.
- Exact payout tables should live in quest balancing assets.

### 13.7 Current progression triggers

- Quest progression should understand travel, field encounters, dungeon completion, and other standard world actions.
- The map decides what can be done in a place; the quest system decides which accepted goals advance when those actions resolve.

### 13.8 Scope and known gaps

- V1 should ship with a stable daily quest framework first.
- Richer branching, deeper procedural variation, and more expressive strategy layers can come later.

### 13.9 Quest framework design principles

- Templates should stay machine-readable first.
- One quest should represent one readable intent.
- Step transitions should stay explicit and bounded.
- The framework should prefer shared runtime patterns over bespoke handler logic.

### 13.10 Recommended quest-step model

- Typical steps include travel, inspect, talk, buy, deliver, defeat, collect, clear-dungeon, and clue-handling patterns.
- Multi-step quests should always expose current step and next-step intent.
- Step shapes should align with the generic action bus where possible.

### 13.11 Recommended quest generation rules

- Generation should consider level band, unlocked regions, available dungeons, and current board composition.
- One board should avoid near-identical tasks while still keeping at least one low-friction option.
- Adding new quests should mostly mean extending templates rather than rewriting the framework.

## 14. Dungeon System

Section role:

- this section explains how dungeons fit into the main world loop
- detailed monsters, room layouts, reward tables, party-entry rules, and battle runtime belong in the dungeon and combat modules

### 14.1 V1 dungeons

- Ancient Catacomb
- Thorned Hollow
- Sunscar Warvault
- Obsidian Spire

All four act as parallel seasonal farms rather than a strict level ladder.

### 14.2 Entry rules

- Entry validates current state, then starts server-side resolution.
- Dungeon access should remain bot-friendly and machine-readable.
- Party-entry, borrowed-assist, and runtime details should be maintained in the dedicated dungeon docs.

### 14.3 Dungeon rewards

- Successful clears produce gold, loot, and boss-linked rewards.
- Dungeons are one of the core gear and material sources in the progression loop.
- Exact reward tables belong in the dungeon and equipment deep-dive docs.

## 15. Arena System

Section role:

- this section keeps the arena loop understandable at the product level
- rating math, bracket operations, betting persistence, and title payout implementation belong in backend and arena-specific specs

### 15.1 Arena eligibility

- Characters may enter during the active signup window.
- Weekday rating challenges run Monday to Friday.
- Saturday bracket signup closes before bracket resolution begins.

### 15.2 Format

- Monday to Friday use rating-based qualification play.
- Friday close freezes the field and produces a top-`64` Saturday bracket.
- NPC entrants can backfill if the live field is too small.
- Saturday resolves as an automated elimination bracket.

### 15.3 Match rules

- Arena uses the same core combat engine as PvE.
- Matches are fully simulated by the server.
- Each duel should produce a readable report for bots and observers.

### 15.4 Rewards

- Arena rewards emphasize titles, recognition, and public competition more than raw gold output.
- Rankings and bracket outcomes should remain visible on the website.
- Betting may open once the top-`64` bracket is locked.

### 15.5 V1 limitations

- No live tactical PvP.
- No live lobby-style matchmaking.
- No real-time spectator-room system in V1.
