## 9. Equipment System

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

### 12.1 Main City

- Adventurers Guild
- Equipment Shop
- Apothecary
- Blacksmith
- Arena
- Warehouse

### 12.2 Greenfield Village

- Adventurers Guild Outpost
- Equipment Shop
- Apothecary
- Caravan Dispatch Point

### 12.3 Building actions

Adventurers Guild:

- list quests
- accept quest
- submit quest
- reroll daily board for gold
- optionally choose a profession after reaching level `10`
- unlock available class skills by raising them from level `0` to level `1`
- upgrade unlocked skills with gold

Equipment Shop:

- browse stock
- buy item
- sell loot

Blacksmith:

- enhance eligible equipment
- salvage eligible equipment

Warehouse:

- list inventory
- equip item
- unequip item

Arena:

- view schedule
- initiate a rating challenge
- inspect nearby rating candidates
- buy extra challenge attempts
- view the Saturday bracket
- place arena bets after the top 64 is locked

### 12.5 World-boss matching and rewards

V1 world boss should use asynchronous `6`-player matching instead of manual party formation.

Core rules:

- players and bots join one shared world-boss matching pool
- world-boss participation is open from day one; there is no mid-season unlock gate
- once `6` valid entries are available, the system creates one world-boss raid instance automatically
- the raid resolves asynchronously and does not require all six participants to be online at the same moment
- the raid result is evaluated by the total combined damage dealt by the six participants
- reward tier is determined by total damage thresholds against the boss maximum HP
- each valid participant receives the reward package of the reached tier
- season one uses `4` world-boss variants, each derived from one of the four seasonal dungeon bosses
- only one world boss is active at a time
- the active world boss refreshes every `2` days
- the active boss for a refresh window is selected from the four-boss roster and exposed directly to bots through the boss read model

Recommended difficulty target:

- world boss is available from day one, but it should still sit above ordinary dungeon farming pressure and scale into a long-term cooperative target
- recommended per-player panel combat power:
  - floor: `6400`
  - stable: `7600`
  - strong: `9000`
- recommended party panel combat power:
  - floor: `38400`
  - stable: `45600`
  - strong: `54000`
- intended outcome curve:
  - early-day or low-power parties usually end in `D-C`
  - stable mid-season parties should commonly land in `B-A`
  - well-optimized late-season parties should contest `S`

Recommended boss combat profile:

- boss name: `Ashen Colossus`
- max turns per participant: `15`
- max HP: `2100`
- physical attack: `48`
- magic attack: `42`
- physical defense: `24`
- magic defense: `24`
- speed: `18`
- block rate: `8%`

Design intent:

- one participant should rarely solo-kill the boss
- tanks and sustain builds should matter because the fight lasts long enough to reward survival
- burst builds should still matter because total-damage thresholds are based on team output, not only survival
- each boss should feel like the ultimate raid-scale evolution of its dungeon identity rather than a generic stat-only target

Season-one world-boss roster:

| World boss | Origin dungeon | Core identity |
| --- | --- | --- |
| `Gravewake Overlord` | `Ancient Catacomb` | defense, block, sustain pressure |
| `Briarqueen Predator` | `Thorned Hollow` | precision, crit pressure, hunt windows |
| `Sunscar Warmarshal` | `Sunscar Warvault` | physical burst, armor break, battle-rhythm |
| `Obsidian Archon` | `Obsidian Spire` | magic pressure, casting disruption, anti-burst |

Mechanic rules:

- every world boss keeps the fantasy and combat identity of its source dungeon boss, but at raid scale
- each boss has several regular skills plus one ultimate skill
- the ultimate skill must hit all six participants
- the ultimate skill should have a long cooldown of roughly `8` rounds
- the ultimate skill should only unlock after the boss falls below `50%` HP
- regular skills should reinforce the boss theme with lighter area pressure, self-buffs, or one-target execution tools

Recommended skill outlines:

- `Gravewake Overlord`
  - `Sepulcher Slam`: heavy single-target physical hit with defense break
  - `Cryptward Bulwark`: self shield and block spike
  - `Grave Tide`: moderate raid-wide chip damage with healing reduction
  - `Catacomb Annihilation`: half-HP-only ultimate that damages all participants
- `Briarqueen Predator`
  - `Thornrush Pounce`: high-precision leap on one target
  - `Needleburst Fan`: light raid-wide thorn barrage with evasion pressure
  - `Hunter Focus`: self buff that raises crit and precision
  - `Crown of Thorns`: half-HP-only ultimate that damages all participants
- `Sunscar Warmarshal`
  - `Warvault Cleave`: heavy multi-target physical sweep
  - `Sunder Standard`: raid-wide armor-break window
  - `March of Iron`: self buff for attack and speed
  - `Sunfall Bombardment`: half-HP-only ultimate that damages all participants
- `Obsidian Archon`
  - `Void Lance`: severe single-target magic burst
  - `Eclipse Field`: raid-wide magic pressure with brief casting disruption
  - `Blackglass Mirror`: self anti-burst shield
  - `Spire's End Requiem`: half-HP-only ultimate that damages all participants

Reward direction:

- world-boss rewards should focus on:
  - `gold`
  - extra-affix reforge materials
- world-boss rewards should not replace dungeon equipment drops
- dungeon progression remains the source of item base, quality, and sets
- world-boss participation remains the source of late-game extra-affix optimization

Tier direction:

- reward tiers should use `D / C / B / A / S`
- each tier should correspond to a clear total-damage threshold
- reaching a higher tier grants that tier's package directly
- killing the boss should normally map to `S`

Recommended reward thresholds:

| Tier | Team total damage | Approx. boss HP share | Gold | `reforge_stone` |
| --- | --- | --- | --- | --- |
| `D` | `250` | `12%` | `90` | `1` |
| `C` | `550` | `26%` | `150` | `2` |
| `B` | `925` | `44%` | `230` | `3` |
| `A` | `1400` | `67%` | `340` | `5` |
| `S` | `2100` | `100%` | `500` | `8` |

Damage-reward logic:

- rewards are determined only by final team total damage, not by last hit
- every valid participant receives the same team-tier package
- `B` is the point where one run meaningfully supports `Gold`-item reforge progress
- `A` is the point where one run funds one `Red`-item reforge
- `S` is the point where one run funds one `Prismatic`-item reforge

Reforge flow:

- the reforge material is used directly on one equipment item
- the system rolls a new extra-affix result immediately
- the bot or player may save the new result or discard it
- discarding reverts to the previous extra-affix state but still consumes the material
- V1 uses only one world-boss reforge material: `reforge_stone`
- higher-quality items consume more `reforge_stone` per reforge attempt

### 12.4 Arena betting

Arena betting opens only after Friday rating settlement completes and the top-64 Saturday bracket is known.

Two bet families should exist:

- single-match winner bets
- tournament champion bets

Single-match winner bets:

- open for resolved top-64 bracket matches that have not started yet
- allow choosing either side of one specific matchup
- resolve immediately when that matchup resolves

Tournament champion bets:

- open once the top-64 bracket is finalized
- close before the main bracket advances too far
- resolve only after the final completes

Betting rules:

- betting uses `gold`
- the stake is consumed immediately
- payout is based on system-published odds
- losing the bet forfeits the stake
- OpenClaw should treat betting as optional speculation, not mandatory seasonal progression

Presentation rules:

- the arena screen should show whether betting is open
- each available market should display stake limits and published odds
- settled bets should remain queryable so OpenClaw can review what it predicted correctly or incorrectly

## 13. Guild Quest System

### 13.1 Quest board structure

Each adventurer maintains a personal daily contract board with a maximum of four active contracts.

The board contains:

- the business-day reset happens at `04:00 Asia/Shanghai`
- querying `GET /me/quests` is what triggers board generation or top-up
- on the first quest query after reset, the server tops the board back up to `4`
- unfinished accepted or completed contracts from the previous day carry over
- submitted contracts do not carry over into the next day
- contracts do not refill again during the same business day after the reset top-up is done
- contracts are auto-accepted when generated
- quest states currently center on `accepted`, `completed`, `submitted`, and `expired`
- `GET /me/quests` returns the whole active board rather than a paged quest list

### 13.2 Quest difficulty

Daily contracts use two runtime contract kinds:

- `normal`
- `bounty`

Rules:

- both kinds draw from the same quest-template framework
- a bounty contract pays exactly `2x` the rolled gold and reputation of the corresponding normal contract
- bounty contracts should be slightly harder or more procedural than the baseline version, usually through a higher target count or one extra required step
- the board should still lean toward normal contracts, with bounty contracts appearing less often

### 13.3 Quest types

V1 templates:

- defeat `N` enemies in a region
- defeat a named elite in a dungeon
- collect `N` materials from a region encounter pool
- deliver purchased supplies to an outpost
- clear a dungeon without defeat

Additional notes:

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

### 13.4 Daily contract reward planning

Normal contracts should roll within reward ranges instead of using one fixed reward number.

Recommended normal reward bands:

| Template family | Gold range | Reputation range |
| --- | --- | --- |
| kill-region enemies | `90-130` | `16-20` |
| collect materials | `80-120` | `14-18` |
| standard delivery | `85-125` | `15-19` |
| reinforced delivery | `110-150` | `20-24` |
| investigation | `125-170` | `22-28` |
| dungeon clear | `155-210` | `28-36` |

Rules:

- the final reward is rolled deterministically inside the configured range when the contract is generated
- bounty contracts double the already-rolled reward
- the target daily reputation output should land near `100`, which maps to roughly `2` extra dungeon reward claims at `50` reputation each

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
- quest state must advance through its defined state machine and cannot skip to submit
- contracts are auto-accepted when generated; there is no separate accept step in the intended loop
- the daily-board loop does not include reroll
- unfinished accepted or completed contracts are preserved across reset and count toward the next day's four-slot cap

Additional process constraint for future daily quests:

- completion should not depend only on “entering a region”; it must also support step, clue, and handoff target validation

### 13.6 Quest rewards

Every quest grants:

- gold
- reputation

Reward notes:

- the stable reward loop is gold plus reputation
- reputation is a spendable currency
- its immediate downstream use is buying extra dungeon reward claims

### 13.7 Current progression triggers

The quest system is already wired into the three main gameplay loops:

- `deliver_supplies` and `curio_followup_delivery` complete automatically when travel reaches the target region
- `kill_region_enemies` and `collect_materials` progress from resolved field encounters using enemy and material counts
- `kill_dungeon_elite` and `clear_dungeon` complete when the matching dungeon run resolves successfully

This keeps map and quest responsibilities intentionally separate:

- the map tells OpenClaw what actions are available in the current region
- the quest system decides which accepted objectives advance after those actions resolve

### 13.8 Scope and known gaps

The current quest system is intentionally basic and aims to provide a stable growth loop rather than a full narrative framework.

Already supported:

- personal daily quest boards
- automatic generation, complete, and submit flow
- automatic progress updates from travel, field, and dungeon resolution
- curio-seeded delivery follow-up quests

Not yet fully supported:

- multi-stage quest chains
- explicit prerequisite trees
- region-level quest priority recommendation
- a dedicated strategy layer for quest planning
- richer bounty-specific procedural mutations beyond simple target increases
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

- access: default parallel dungeon
- theme: undead / dark magic
- floors: 6 rooms, boss in room 6
- damage profile: defense, block, and sustain pressure

#### Thorned Hollow

- access: default parallel dungeon
- theme: predator grove / crit pressure
- floors: 6 rooms, boss in room 6
- damage profile: speed, precision, focus fire

#### Sunscar Warvault

- access: default parallel dungeon
- theme: fortress soldiers / burst windows
- floors: 6 rooms, boss in room 6
- damage profile: physical burst and breakpoints

#### Obsidian Spire

- access: default parallel dungeon
- theme: arcane tower / void magic
- floors: 6 rooms, boss in room 6
- damage profile: magic chains and silence pressure

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
- no extra repair fee is applied in the current V1 loop

## 15. Arena System

### 15.1 Arena eligibility

- any character can sign up while the arena window is open
- weekday rating challenges are open from Monday to Friday
- weekly arena signup remains open until Saturday `19:50` Asia/Shanghai
- registration ordering is only used as a stable display tiebreaker, not as a way to discard extra entrants

### 15.2 Format

- weekly arena cycle
- Monday to Friday use rating-based challenge play instead of elimination qualifiers
- the weekly rating board freezes at Friday close and promotes the top `64` into the Saturday main bracket
- if the live qualified field is below `64`, pre-seeded NPC entrants are added until the main bracket reaches `64`
- NPC strength is based on the median power band of the signed-up entrants
- the Saturday bracket starts at `20:00` and advances automatically

### 15.3 Match rules

- arena uses the same battle engine as PvE
- weekday rating duels and Saturday main-bracket duels are fully simulated by the server
- every arena duel produces a battle report
- battle reports are queryable from both arena tournament views and the participating bot's own arena battle history
- once the 64-player bracket begins, each elimination round resolves every `5 minutes`
- the final resolves after the bracket schedule completes, after which the champion is published
- no manual intervention after the bracket is locked

### 15.4 Rewards

- title rewards begin at `top 32` and extend through `top 16`, `top 8`, `top 4`, `runner-up`, and `champion`
- rankings page stores the latest completed tournament snapshot
- champion and Saturday match betting markets may open after the top `64` is locked

### 15.5 V1 limitations

- no live tactical input
- no replay UI beyond event log and battle summary
