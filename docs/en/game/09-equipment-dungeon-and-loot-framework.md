# Equipment, Dungeon, and Loot Framework

## 1. Design Goals

This module defines the revised dungeon-driven equipment framework for ClawGame.

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

- civilians may equip any weapon style and any non-weapon slot item
- promoted characters use class and weapon-style restrictions when equipping weapons
- civilians are not bound to any class weapon before profession choice and may freely equip any weapon style
- drops themselves are not class-filtered
- all non-weapon slots are universal
- only one item may occupy each slot
- all current seasonal dungeon gear pieces count by `set_id`

## 3. Shared Dungeon Loot Structure

The four dungeons are parallel seasonal farms rather than a strict low-to-high ladder.

System rules:

- all four dungeons drop the current seasonal dungeon gear band
- base stat budget is shared across all four dungeons
- base reward economy is shared across all four dungeons:
  - same gold baseline
  - same material rules
  - same rating reward count and quality logic
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

### 3.1 Assist-template Party Entry

Dungeons should support solo entry plus optional assist-template party entry.

#### Party size rules

| Topic | Rule |
| --- | --- |
| Allowed party size | `1-3` total members |
| Entry owner | One bot always acts as the run owner and reward owner |
| Additional members | Up to `2` additional slots may be filled by borrowed assist snapshots |
| Solo support | Every dungeon must remain clearable as a solo run |
| Party role | Assist-template entry improves team composition and clear stability, but does not replace solo access |

#### Borrowed snapshot rules

| Topic | Rule |
| --- | --- |
| Snapshot source | Each borrowed slot uses one combat snapshot captured from the selected bot's current battle-ready state at entry time |
| Snapshot timing | The snapshot is locked when the run starts and remains fixed for the full run |
| Snapshot contents | Snapshot should include current class, stat panel, equipment summary, equipped skills, potion loadout, and power snapshot |
| Runtime isolation | Snapshot use does not consume the borrowed bot's own dungeon quota, potions, durability, or inventory resources |
| Reward participation | Snapshot allies do not count as independent reward owners |

#### Borrowing cost and social payout rules

| Topic | Rule |
| --- | --- |
| Stranger borrow cost | `150` gold per borrowed template slot |
| Friend borrow cost | `75` gold per borrowed template slot |
| Stranger borrow limit | Each bot may borrow from the same stranger bot at most `1` time per day |
| Payout per borrowed snapshot | `50` gold granted to the borrowed bot |
| Daily payout cap | Each borrowed bot may receive at most `1000` gold per day from being borrowed as a snapshot |
| Reward owner | Only the run owner receives dungeon clear rewards, loot, and rating rewards |
| Economic direction | Borrowing is a tactical convenience and social incentive, not a way to duplicate dungeon rewards across accounts |

#### Difficulty and reward rules

| Topic | Rule |
| --- | --- |
| Difficulty tuning | Dungeons should scale by actual member count so solo, duo, and trio entries remain valid |
| Reward direction | Party size should improve clear stability and rating potential rather than multiply the reward table directly |
| Rating logic | Higher-member entries may more reliably reach high ratings, but rating thresholds remain shared |
| System boundary | This is asynchronous assist-template dungeon play, not synchronous co-op or manual live party control |

### 3.2 Localization Display Names

UI should expose localized display names while preserving stable English ids and template-direction keys in data.

#### Dungeon display names

| English name | Chinese display name |
| --- | --- |
| `Ancient Catacomb` | 远古墓窟 |
| `Thorned Hollow` | 荆棘空谷 |
| `Sunscar Warvault` | 日痕战库 |
| `Obsidian Spire` | 黑曜尖塔 |

#### Set display names

| set_id / English set name | Chinese display name |
| --- | --- |
| `Gravewake Bastion` | 墓醒壁垒 |
| `Briarbound Sight` | 棘界猎眼 |
| `Sunscar Assault` | 日蚀锋袭 |
| `Nightglass Arcanum` | 夜镜秘典 |

#### Weapon-style display names

| weapon_style | Chinese display name |
| --- | --- |
| `sword_shield` | 剑盾 |
| `great_axe` | 巨斧 |
| `staff` | 法杖 |
| `spellbook` | 咒书 |
| `scepter` | 权杖 |
| `holy_tome` | 圣典 |

#### Template-direction display names

| template direction | Chinese display name |
| --- | --- |
| `physical` | 物理 |
| `magic` | 魔法 |
| `healing` | 治疗 |
| `universal` | 通用 |
| `physical-guard` | 物防 |
| `magic-guard` | 魔防 |
| `sustain` | 续航 |
| `caster` | 施法 |

## 4. Quality and Affix Rules

The game uses five quality grades:

| Quality | Color | Extra Affixes | Set Bonus | Identity |
| --- | --- | --- | --- | --- |
| Blue | blue | 1 | no | stable baseline drop |
| Purple | purple | 2 | no | efficient transition piece |
| Gold | gold | 3 | no | high-value non-set piece |
| Red | red | 4 | yes | standard set chase piece |
| Prismatic | rainbow | 4 | yes | premium chase piece |

Rules:

- every item has fixed main affixes tied to slot identity
- every item also rolls slot-aligned random secondary affixes
- blue items start with `1` extra random affix
- extra-affix count scales by quality from Blue through Prismatic
- all current seasonal dungeon items belong to a named dungeon set family
- all set lines support `2-piece`, `4-piece`, and `6-piece`
- prismatic is the apex version of the same dungeon set family rather than a separate set line
- prismatic extra affixes always roll at max value; other qualities roll within a controlled range

### 4.1 Three-Layer Item Stat Structure

Each item should be understood as having three stat layers:

1. fixed main affixes
2. slot-aligned random secondary affixes
3. fully random extra affixes

Design intent:

- main affixes make the slot readable at a glance
- secondary affixes preserve slot identity while still allowing meaningful variance
- extra affixes create the long-tail chase and can break away from slot expectations

Recommended rule:

- every item has `1` fixed main-affix package
- every item rolls `1` random secondary affix from the slot-aligned pool
- extra affix count depends on quality: Blue `1`, Purple `2`, Gold `3`, Red `4`, Prismatic `4`

### 4.2 Fixed Main Affixes By Slot

The main-affix package should be deterministic for the item template and should closely match slot identity.

Recommended slot direction:

- `weapon`: main offensive package, centered on `physical_attack`, `magic_attack`, or `healing_power`
- `head`: balanced survivability package, centered on `max_hp`, `physical_defense`, `magic_defense`
- `chest`: largest defensive package, centered on `max_hp`, `physical_defense`, `magic_defense`
- `boots`: tempo package, centered on `speed`, `max_hp`, and one defense stat
- `ring`: offensive specialization package, centered on `physical_attack`, `magic_attack`, or `healing_power`
- `necklace`: sustain and utility package, centered on `max_hp`, `healing_power`, `magic_defense`, and occasional `speed`

Rule:

- the same template should always keep the same main-affix package
- dungeon choice should not change the main-affix structure of a slot

### 4.2.1 Fixed Main-Affix Baseline Values

Because the first live season uses one shared seasonal gear band, the slot templates below should be treated as the baseline main-affix packages for that full season.

Recommended main-affix templates:

| Slot | Template direction | Fixed main-affix package |
| --- | --- | --- |
| `weapon` | physical | `+30 physical_attack` |
| `weapon` | magic | `+30 magic_attack` |
| `weapon` | healing | `+22 healing_power`, `+10 magic_attack` |
| `head` | universal | `+90 max_hp`, `+7 physical_defense`, `+7 magic_defense` |
| `chest` | universal | `+150 max_hp`, `+12 physical_defense`, `+12 magic_defense` |
| `boots` | physical-guard | `+60 max_hp`, `+4 speed`, `+5 physical_defense` |
| `boots` | magic-guard | `+60 max_hp`, `+4 speed`, `+5 magic_defense` |
| `ring` | physical | `+16 physical_attack` |
| `ring` | magic | `+16 magic_attack` |
| `ring` | healing | `+16 healing_power` |
| `necklace` | sustain | `+80 max_hp`, `+10 healing_power`, `+6 magic_defense` |
| `necklace` | caster | `+72 max_hp`, `+8 magic_attack`, `+6 magic_defense`, `+2 speed` |

Budget rule:

- the fixed main-affix package should provide about `58%-62%` of a normal item's total non-set stat budget before enhancement
- weapon and chest should remain the largest single-slot contributors
- ring and necklace should stay below weapon impact even on high-roll items

### 4.3 Slot-Aligned Random Secondary Affixes

Secondary affixes should still lean toward what the slot is naturally good at, but they should remain random within that local identity.

Recommended secondary-affix pools:

| Slot | Random secondary-affix pool |
| --- | --- |
| `weapon` | `physical_attack`, `magic_attack`, `healing_power`, `accuracy`, `crit_chance`, `crit_damage`, `speed` |
| `head` | `max_hp`, `physical_defense`, `magic_defense`, `speed`, `healing_power` |
| `chest` | `max_hp`, `physical_defense`, `magic_defense`, `healing_power` |
| `boots` | `speed`, `max_hp`, `physical_defense`, `magic_defense`, `accuracy` |
| `ring` | `physical_attack`, `magic_attack`, `healing_power`, `accuracy`, `crit_chance`, `crit_damage`, `speed` |
| `necklace` | `max_hp`, `healing_power`, `magic_attack`, `magic_defense`, `speed`, `accuracy` |

Rules:

- secondary affixes are where slot identity remains visible after the fixed main-affix package
- this layer should feel more coherent than the fully random extra-affix layer

### 4.3.1 Random Secondary-Affix Value Ranges

Each item rolls exactly one random secondary affix from its slot-aligned pool.

Recommended secondary-affix ranges:

| Slot | Secondary affix | Recommended range |
| --- | --- | --- |
| `weapon` | `physical_attack` | `+8 to +12` |
| `weapon` | `magic_attack` | `+8 to +12` |
| `weapon` | `healing_power` | `+8 to +12` |
| `weapon` | `accuracy` | `+4% to +7%` |
| `weapon` | `crit_chance` | `+2.0% to +3.2%` |
| `weapon` | `crit_damage` | `+4.0% to +6.5%` |
| `weapon` | `speed` | `+1 to +2` |
| `head` | `max_hp` | `+35 to +55` |
| `head` | `physical_defense` | `+3 to +5` |
| `head` | `magic_defense` | `+3 to +5` |
| `head` | `speed` | `+1` |
| `head` | `healing_power` | `+4 to +7` |
| `chest` | `max_hp` | `+50 to +80` |
| `chest` | `physical_defense` | `+4 to +6` |
| `chest` | `magic_defense` | `+4 to +6` |
| `chest` | `healing_power` | `+5 to +8` |
| `boots` | `speed` | `+2 to +4` |
| `boots` | `max_hp` | `+30 to +50` |
| `boots` | `physical_defense` | `+3 to +5` |
| `boots` | `magic_defense` | `+3 to +5` |
| `boots` | `accuracy` | `+3% to +6%` |
| `ring` | `physical_attack` | `+6 to +10` |
| `ring` | `magic_attack` | `+6 to +10` |
| `ring` | `healing_power` | `+6 to +10` |
| `ring` | `accuracy` | `+4% to +7%` |
| `ring` | `crit_chance` | `+2.0% to +3.5%` |
| `ring` | `crit_damage` | `+5.0% to +8.0%` |
| `ring` | `speed` | `+1 to +2` |
| `necklace` | `max_hp` | `+30 to +50` |
| `necklace` | `healing_power` | `+6 to +10` |
| `necklace` | `magic_attack` | `+6 to +10` |
| `necklace` | `magic_defense` | `+3 to +5` |
| `necklace` | `speed` | `+1 to +2` |
| `necklace` | `accuracy` | `+3% to +6%` |

Budget rule:

- the random secondary affix should contribute about `16%-20%` of a normal item's total non-set stat budget before enhancement
- even a high-roll secondary affix should not outweigh the slot's fixed main-affix package

### 4.4 Global Extra-Affix Pool

Extra affixes should come from one shared global pool to preserve openness and chase value.

Recommended global extra-affix pool:

- `max_hp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`
- `accuracy`
- `crit_chance`
- `crit_damage`

Design rule:

- do not hard-code affix validity by class
- the same affix can be stronger or weaker depending on the build, and that is an intended part of the sandbox
- bots should evaluate affixes by current build goals rather than by static "invalid stat" labels

Rules:

- all equipment slots roll extra affixes from the same global affix pool
- slot identity is expressed through the main-affix and secondary-affix layers, not through extra-affix restrictions
- a single item should not roll the exact same affix twice
- slots may roll both flat-value and percent-value variants later, but the first live version should keep naming and evaluation unified at the system layer

Implication:

- a defensive chest may still roll offensive extra affixes
- an offensive ring may still roll survivability extra affixes
- this randomness is intentional and helps preserve long-tail chase value across the full season

### 4.5 Quality Gates For Affixes

The first live season should keep affix complexity layered by quality.

Recommended gates:

| Quality | Allowed affix scope |
| --- | --- |
| Blue | fixed main affixes + 1 slot-aligned random secondary affix + 1 extra affix from the shared global pool |
| Purple | Blue structure + 2 extra affixes from the shared global pool with a stronger roll range |
| Gold | Blue structure + 3 extra affixes from the shared global pool; may start rolling `accuracy`, `crit_chance`, `crit_damage` |
| Red | Blue structure + 4 extra affixes from the full shared global pool |
| Prismatic | Blue structure + 4 extra affixes from the full shared global pool, all at max roll value |

Extension rule:

- niche affixes such as `boss_damage`, `guard_power`, `defense_ignore`, or `skill_effect` should remain optional future additions rather than part of the first mandatory launch pool
- if such affixes are introduced later, prefer gating them to `Gold` and above, or even `Red` and above, to keep early gear legible

Roll-value rule:

- Blue, Purple, Gold, and Red extra affixes should roll within a predefined min-max range
- Prismatic extra affixes should always use the maximum value of that range
- this keeps prismatic clearly aspirational without increasing the affix-count ceiling beyond `4`

### 4.6 Extra-Affix Roll Ranges

Roll ranges should stay readable and narrow enough that item quality matters more than lottery variance.

Recommended roll-position ranges:

| Quality | Roll position within affix range |
| --- | --- |
| Blue | `70%-82%` |
| Purple | `78%-88%` |
| Gold | `84%-93%` |
| Red | `90%-97%` |
| Prismatic | `100%` |

Interpretation:

- each extra affix has its own stat-specific min and max value
- the quality determines where inside that min-max band the roll can land
- prismatic always lands on the max value for every extra affix

Recommended examples for percentage-style extra affixes:

| Affix | Full range | Blue | Purple | Gold | Red | Prismatic |
| --- | --- | --- | --- | --- | --- | --- |
| `max_hp%` | `3.0%-6.0%` | `4.0%-4.4%` | `4.3%-4.6%` | `4.5%-4.8%` | `4.7%-4.8%` | `6.0%` |
| `physical_attack%` | `2.5%-4.5%` | `3.9%-4.1%` | `4.1%-4.3%` | `4.2%-4.4%` | `4.3%-4.4%` | `4.5%` |
| `magic_attack%` | `2.5%-4.5%` | `3.9%-4.1%` | `4.1%-4.3%` | `4.2%-4.4%` | `4.3%-4.4%` | `4.5%` |
| `speed%` | `1.5%-3.0%` | `2.6%-2.7%` | `2.7%-2.8%` | `2.8%-2.9%` | `2.8%-3.0%` | `3.0%` |
| `crit_chance%` | `1.5%-3.5%` | `2.9%-3.1%` | `3.0%-3.2%` | `3.2%-3.4%` | `3.3%-3.4%` | `3.5%` |
| `crit_damage%` | `3.0%-7.0%` | `5.8%-6.3%` | `6.1%-6.5%` | `6.4%-6.7%` | `6.6%-6.9%` | `7.0%` |

Recommended examples for flat-value extra affixes:

| Affix | Full range | Blue | Purple | Gold | Red | Prismatic |
| --- | --- | --- | --- | --- | --- | --- |
| `max_hp` | `50-110` | `92-99` | `96-103` | `100-106` | `104-108` | `110` |
| `physical_attack` | `6-16` | `13-14` | `13-14` | `14-15` | `15` | `16` |
| `magic_attack` | `6-16` | `13-14` | `13-14` | `14-15` | `15` | `16` |
| `physical_defense` | `5-12` | `10-11` | `10-11` | `10-11` | `11` | `12` |
| `magic_defense` | `5-12` | `10-11` | `10-11` | `10-11` | `11` | `12` |
| `healing_power` | `6-16` | `13-14` | `13-14` | `14-15` | `15` | `16` |

Operational rule:

- the exact min-max table should be authored per affix family in data tables, not hand-written per item template
- if balance testing shows too much frustration, narrow the Blue-to-Red ranges before reducing affix counts
- prismatic should remain rare because its value comes from both quality and perfect rolls

Calibration note:

- with the main-affix, secondary-affix, and quality-base values in this spec, a representative full Red set should stay near the `31%-33%` endgame share target
- a representative full Prismatic set should stay near the `33%-35%` endgame share target rather than blowing past it
### 4.7 Shared Affix Logic Across All Four Dungeons

The four seasonal dungeons should not have different extra-affix bias tables.

Rules:

- all four dungeons use the same fixed-main-affix rules by slot
- all four dungeons use the same slot-aligned secondary-affix rules
- all four dungeons use the same global extra-affix pool without dungeon-specific bias
- dungeon choice changes set family and set effects, not the affix bias on the dropped item
- bots may still prefer one dungeon because of the set bonus, but should not expect a different affix pool there

### 4.8 Extra-Affix Reforge

World-boss participation introduces a dedicated reforge loop for the extra-affix layer.

Core rules:

- reforge materials are consumed by applying them directly to one equipment item
- reforge only changes the extra-affix layer
- fixed main affixes never change through reforge
- the slot-aligned random secondary affix never changes through reforge
- one reforge attempt always consumes the required material immediately
- after the reroll result is generated, the bot or player may either save the new result or discard it
- discarding the new result restores the previous extra-affix set, but the spent material is not refunded

Current runtime behavior:

- `POST /api/v1/items/{itemId}/reforge` creates one pending preview rather than overwriting the item immediately
- the pending preview stores `material_key`, `material_quantity`, `previous_affixes`, `preview_affixes`, and `created_at`
- `save` replaces the item's current `extra_affixes` with `preview_affixes` and then recomputes final stats
- `discard` clears the pending preview and leaves the item's current affixes unchanged
- Reforge cost is quality-based and currently resolved as `1 / 2 / 3 / 5 / 8` for `blue / purple / gold / red / prismatic`

Design intent:

- dungeons remain responsible for farming item base, quality, and set identity
- world-boss participation remains responsible for late-game extra-affix optimization
- reforge should create a meaningful keep-or-revert decision without risking permanent loss of the whole item

Recommended material direction:

- the system uses only one reforge material: `reforge_stone`
- one reforge attempt rerolls all extra affixes on the chosen item once
- reforge should not be split into multiple stone families yet

Recommended reforge-stone cost by item quality:

| Item quality | Reforge-stone cost per attempt |
| --- | --- |
| `Blue` | `1` |
| `Purple` | `2` |
| `Gold` | `3` |
| `Red` | `5` |
| `Prismatic` | `8` |

Recommended world-boss reforge-stone rewards:

| Reward tier | Total-damage threshold | Gold | `reforge_stone` |
| --- | --- | --- | --- |
| `D` | `>= 250 / 2100` | `90` | `1` |
| `C` | `>= 550 / 2100` | `150` | `2` |
| `B` | `>= 925 / 2100` | `230` | `3` |
| `A` | `>= 1400 / 2100` | `340` | `5` |
| `S` | `>= 2100 / 2100` | `500` | `8` |

Balance intent:

- `Blue` and `Purple` items should be inexpensive to iterate on
- `Gold` items should feel like regular optimization targets once bots stabilize world-boss participation
- `Red` items should require deliberate spending decisions
- `Prismatic` items should remain expensive enough that perfect-endgame optimization stays slow

Pending-result presentation:

- before confirmation, the UI and bot read model should expose:
  - previous extra affixes
  - newly rolled extra affixes
  - a save action
  - a discard action
- save makes the rerolled extra affixes permanent
- discard keeps the previous extra affixes and closes the pending result

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
- all quality results use the active dungeon's `set_id`
- every quality pool includes all `6` weapon styles alongside the dungeon's armor and accessory pieces
- dropped weapons are class-agnostic at acquisition time and only enforce weapon-style compatibility when equipped
- difficulty changes efficiency, not the base stat budget of an item at the same quality

Runtime note:

- inventory items expose `set_id`
- inventory responses expose `equipped_set_bonuses`
- active `2 / 4 / 6` piece thresholds currently apply direct stat snapshots so set identity has immediate combat impact

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
- define whether block is a formal combat stat or an event-style defensive keyword
