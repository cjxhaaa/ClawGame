# Equipment, Dungeon, and Loot Framework

## 1. Design Goals

This module defines the first complete dungeon-driven equipment framework for ClawGame.

Goals:

- equipment progression should be primarily sourced from dungeon drops
- a new baseline gear band should appear every `20` levels
- bots should clearly understand what dungeon they should farm for each level bracket
- only weapons are class-restricted
- all other wearable slots are shared across all classes
- rarity, affix count, and set effects should be explicit and machine-readable
- late-season play should naturally shift from leveling routes into dungeon and material routes

## 2. Equipment Slots and Wear Rules

Each character equips the following slots:

- `weapon`
- `head`
- `wrist`
- `chest`
- `boots`
- `ring`
- `necklace`

Rules:

- `weapon` is class-restricted
- all other slots are universal and can be worn by any class
- only one item may occupy each slot
- red and prismatic set pieces count by equipped set name, not by slot family
- gear score should be derived from item tier, rarity, affix count, and upgrade level later

## 3. Tier Structure

Equipment is divided into five progression tiers tied to season level.

| Tier | Recommended Level Range | Dungeon Band | Core Purpose |
| --- | --- | --- | --- |
| T1 | 1-20 | novice dungeon band | onboarding and stable early stats |
| T2 | 21-40 | low-mid dungeon band | transition out of starter routes |
| T3 | 41-60 | mid dungeon band | first serious build shaping |
| T4 | 61-80 | advanced dungeon band | pre-cap optimization |
| T5 | 81-100 | capstone dungeon band | max-level farming and set chasing |

Recommended equip rule:

- bots can wear any tier they meet the level requirement for
- dungeon drop tables heavily favor the matching band
- lower-tier gear remains useful mainly for temporary gaps and salvage materials

## 4. Rarity and Affix Rules

The game uses five quality grades.

| Quality | Color | Affix Count | Set Bonus | Intended Identity |
| --- | --- | --- | --- | --- |
| Blue | blue | 0 | no | stable baseline drop |
| Purple | purple | 1 | no | efficient progression piece |
| Gold | gold | 2 | no | premium farm target before set chase |
| Red | red | 3 | yes | set-entry item |
| Prismatic | rainbow | 4 | yes | apex chase item |

Rules:

- blue gear only has base slot stats
- every rarity step above blue adds `+1` extra affix
- red and prismatic gear can belong to named sets
- red and prismatic sets support `2-piece`, `4-piece`, and `6-piece` effects
- prismatic items should not introduce a separate set line in V1; they should be the highest-quality form of the same dungeon set family

## 5. Slot Stat Identity

Each slot should have a predictable baseline role.

### 5.1 Weapon

Main purpose:

- primary offense slot
- strongest source of attack or healing scaling

Main stat directions:

- warrior weapons: `physical_attack`, small `max_hp`
- mage weapons: `magic_attack`, small `max_mp`
- priest weapons: `magic_attack`, `healing_power`, small `max_mp`

### 5.2 Head

Main purpose:

- balanced survivability slot

Main stat directions:

- `max_hp`
- `physical_defense`
- `magic_defense`

### 5.3 Wrist

Main purpose:

- precision, offensive support, and build shaping

Main stat directions:

- `physical_attack` or `magic_attack`
- `speed`
- niche combat utility

### 5.4 Chest

Main purpose:

- largest defensive armor slot

Main stat directions:

- `max_hp`
- `physical_defense`
- `magic_defense`

### 5.5 Boots

Main purpose:

- movement tempo and tactical survivability

Main stat directions:

- `speed`
- `max_hp`
- one defensive stat

### 5.6 Ring

Main purpose:

- offensive specialization

Main stat directions:

- `physical_attack`
- `magic_attack`
- `healing_power`

### 5.7 Necklace

Main purpose:

- resource and sustain specialization

Main stat directions:

- `max_mp`
- `healing_power`
- `magic_defense`
- occasional `speed`

## 6. Affix Pool

Base affix pool for V1:

- `+max_hp`
- `+max_mp`
- `+physical_attack`
- `+magic_attack`
- `+physical_defense`
- `+magic_defense`
- `+speed`
- `+healing_power`

Optional phase-two affixes:

- `+accuracy`
- `+status_mastery`
- `+status_resistance`

Affix rules:

- affixes should roll from a slot-appropriate pool
- the same affix should not duplicate on the same item in V1
- blue items never roll affixes
- prismatic items always roll four affixes with at least one high-value affix

## 7. Dungeon Bands and Equipment Themes

The following dungeons define the first five equipment bands.

## 7.1 T1 Dungeon: Ancient Catacomb

Level band:

- recommended for levels `1-20`

Theme:

- sealed tomb corridors
- undead sentries
- necromancer rites

Set family name:

- `Gravewake`

Narrative identity:

- gear made from grave iron, candle soot cloth, and burial wards reclaimed from the catacomb

### T1 Weapon Names

| Class | Item Name | Theme |
| --- | --- | --- |
| Warrior | Gravewake Bastionblade | tomb guard sword-shield weapon |
| Warrior | Gravewake Reaper Axe | executioner axe recovered from burial halls |
| Mage | Gravewake Ash Staff | ashwood conduit from ritual braziers |
| Mage | Gravewake Bone Codex | bone-bound spellbook from a necromancer archive |
| Priest | Gravewake Chapel Scepter | reliquary scepter reclaimed from a ruined chapel |
| Priest | Gravewake Vigil Tome | prayer tome sealed with funeral wax |

### T1 Universal Armor Names

| Slot | Item Name | Flavor |
| --- | --- | --- |
| Head | Gravewake Hood | stitched from burial cloth and ward thread |
| Wrist | Gravewake Shackles | iron bands once used on the dead |
| Chest | Gravewake Vestment | robe or coat lined with grave sigils |
| Boots | Gravewake Marchers | boots hardened by tomb dust |
| Ring | Gravewake Seal | seal ring from catacomb wardens |
| Necklace | Gravewake Reliquary | bone charm holding dim holy residue |

### T1 Baseline Slot Stats

| Slot | Main Stats |
| --- | --- |
| Weapon | +18 primary attack, +6 secondary scaling |
| Head | +60 HP, +6 P.Def, +6 M.Def |
| Wrist | +8 primary attack, +2 speed |
| Chest | +90 HP, +10 P.Def, +10 M.Def |
| Boots | +45 HP, +3 speed, +4 P.Def |
| Ring | +10 primary attack |
| Necklace | +40 MP, +8 healing or +6 magic attack |

## 7.2 T2 Dungeon: Thorned Hollow

Level band:

- recommended for levels `21-40`

Theme:

- overgrown ruin under Whispering Forest
- venomous roots
- cursed beast altars

Set family name:

- `Briarbound`

Narrative identity:

- gear woven from thorn hide, root resin, and druidic ward fragments

### T2 Weapon Names

| Class | Item Name |
| --- | --- |
| Warrior | Briarbound Wardblade |
| Warrior | Briarbound Ravager Axe |
| Mage | Briarbound Sap Staff |
| Mage | Briarbound Hex Grimoire |
| Priest | Briarbound Bloom Scepter |
| Priest | Briarbound Oath Tome |

### T2 Universal Armor Names

| Slot | Item Name |
| --- | --- |
| Head | Briarbound Crown |
| Wrist | Briarbound Bracers |
| Chest | Briarbound Carapace |
| Boots | Briarbound Rootwalkers |
| Ring | Briarbound Thornband |
| Necklace | Briarbound Heartvine Pendant |

### T2 Baseline Slot Stats

| Slot | Main Stats |
| --- | --- |
| Weapon | +34 primary attack, +10 secondary scaling |
| Head | +110 HP, +10 P.Def, +10 M.Def |
| Wrist | +14 primary attack, +4 speed |
| Chest | +160 HP, +16 P.Def, +16 M.Def |
| Boots | +80 HP, +5 speed, +7 M.Def |
| Ring | +18 primary attack or +14 healing |
| Necklace | +80 MP, +12 healing or +12 magic attack |

## 7.3 T3 Dungeon: Sunscar Warvault

Level band:

- recommended for levels `41-60`

Theme:

- buried desert armory
- heat-forged automata
- scorched military relics

Set family name:

- `Sunscar`

Narrative identity:

- gear salvaged from a ruined war vault beneath the frontier sands

### T3 Weapon Names

| Class | Item Name |
| --- | --- |
| Warrior | Sunscar Bulwark Edge |
| Warrior | Sunscar Siege Axe |
| Mage | Sunscar Ember Rod |
| Mage | Sunscar Signal Codex |
| Priest | Sunscar Standard Scepter |
| Priest | Sunscar Covenant Tome |

### T3 Universal Armor Names

| Slot | Item Name |
| --- | --- |
| Head | Sunscar Helmguard |
| Wrist | Sunscar Command Wraps |
| Chest | Sunscar Warplate |
| Boots | Sunscar Dune Striders |
| Ring | Sunscar Officer Ring |
| Necklace | Sunscar Medalion |

### T3 Baseline Slot Stats

| Slot | Main Stats |
| --- | --- |
| Weapon | +54 primary attack, +16 secondary scaling |
| Head | +170 HP, +15 P.Def, +15 M.Def |
| Wrist | +22 primary attack, +6 speed |
| Chest | +240 HP, +24 P.Def, +22 M.Def |
| Boots | +120 HP, +7 speed, +10 P.Def |
| Ring | +28 primary attack or +22 healing |
| Necklace | +130 MP, +18 healing or +18 magic attack |

## 7.4 T4 Dungeon: Obsidian Spire

Level band:

- recommended for levels `61-80`

Theme:

- volcanic black tower
- void priests
- obsidian mirrors and curse engines

Set family name:

- `Nightglass`

Narrative identity:

- gear forged from black glass shards and sealed abyssal script

### T4 Weapon Names

| Class | Item Name |
| --- | --- |
| Warrior | Nightglass Bastion Fang |
| Warrior | Nightglass Cataclysm Axe |
| Mage | Nightglass Rift Staff |
| Mage | Nightglass Mirror Codex |
| Priest | Nightglass Halo Scepter |
| Priest | Nightglass Liturgy Tome |

### T4 Universal Armor Names

| Slot | Item Name |
| --- | --- |
| Head | Nightglass Visor |
| Wrist | Nightglass Bindings |
| Chest | Nightglass Aegis |
| Boots | Nightglass Tread |
| Ring | Nightglass Vow Ring |
| Necklace | Nightglass Eclipse Pendant |

### T4 Baseline Slot Stats

| Slot | Main Stats |
| --- | --- |
| Weapon | +78 primary attack, +24 secondary scaling |
| Head | +250 HP, +22 P.Def, +22 M.Def |
| Wrist | +32 primary attack, +9 speed |
| Chest | +340 HP, +34 P.Def, +30 M.Def |
| Boots | +170 HP, +9 speed, +14 M.Def |
| Ring | +40 primary attack or +32 healing |
| Necklace | +190 MP, +26 healing or +26 magic attack |

## 7.5 T5 Dungeon: Sandworm Den

Level band:

- recommended for levels `81-100`

Theme:

- colossal sand tunnels
- venom pressure
- matriarch ambush patterns

Set family name:

- `Dunescourge`

Narrative identity:

- gear crafted from sandworm shell, venom sacs, and matriarch spinal crystal

### T5 Weapon Names

| Class | Item Name |
| --- | --- |
| Warrior | Dunescourge Matriarch Blade |
| Warrior | Dunescourge Mawsplitter |
| Mage | Dunescourge Venomspire |
| Mage | Dunescourge Brood Ledger |
| Priest | Dunescourge Dawn Scepter |
| Priest | Dunescourge Lifebind Tome |

### T5 Universal Armor Names

| Slot | Item Name |
| --- | --- |
| Head | Dunescourge Crownshell |
| Wrist | Dunescourge Coilguards |
| Chest | Dunescourge Carapace Mail |
| Boots | Dunescourge Burrowstep Boots |
| Ring | Dunescourge Fang Ring |
| Necklace | Dunescourge Heartspine Chain |

### T5 Baseline Slot Stats

| Slot | Main Stats |
| --- | --- |
| Weapon | +108 primary attack, +34 secondary scaling |
| Head | +360 HP, +30 P.Def, +30 M.Def |
| Wrist | +44 primary attack, +11 speed |
| Chest | +470 HP, +45 P.Def, +40 M.Def |
| Boots | +230 HP, +12 speed, +18 P.Def |
| Ring | +56 primary attack or +42 healing |
| Necklace | +260 MP, +36 healing or +36 magic attack |

## 8. Set Effects

Only red and prismatic items can roll as set pieces.

Each dungeon set uses the same progression shape:

### 8.1 Gravewake Set

- `2-piece`: +8% max HP, +6% magic defense
- `4-piece`: after entering battle, gain a tomb ward shield equal to 12% max HP for 2 turns
- `6-piece`: once per battle, when dropping below 35% HP, cleanse one negative status and restore 10% HP

### 8.2 Briarbound Set

- `2-piece`: +6% speed, +8% status resistance
- `4-piece`: damaging a poisoned or rooted enemy grants +8% primary attack for 2 turns
- `6-piece`: the first control or poison effect applied each battle gains +1 turn duration

### 8.3 Sunscar Set

- `2-piece`: +10% physical attack and +10% magic attack
- `4-piece`: first offensive skill each battle deals +18% increased damage
- `6-piece`: after defeating an elite or boss phase, restore 12% HP and 12% MP

### 8.4 Nightglass Set

- `2-piece`: +12% magic defense and +10% healing power
- `4-piece`: when casting a non-basic skill, gain 8% damage reduction for 2 turns
- `6-piece`: once every 4 turns, the next magic or holy skill costs 0 MP and gains +15% effect value

### 8.5 Dunescourge Set

- `2-piece`: +10% max HP and +10% speed
- `4-piece`: against boss targets, gain +12% damage and +12% healing output
- `6-piece`: after receiving lethal damage once per battle, survive at 1 HP, gain a 15% HP shield, and cleanse poison

## 9. Loot Probability Model

Drop rates should be strict enough to support long-term farming without making upgrades feel impossible.

## 9.1 Standard Encounter Chest

| Quality | Drop Rate |
| --- | --- |
| Blue | 72.00% |
| Purple | 20.00% |
| Gold | 6.50% |
| Red | 1.30% |
| Prismatic | 0.20% |

## 9.2 Dungeon Final Boss Drop

| Quality | Drop Rate |
| --- | --- |
| Blue | 48.00% |
| Purple | 30.00% |
| Gold | 16.00% |
| Red | 5.00% |
| Prismatic | 1.00% |

## 9.3 High-Difficulty Boss Drop

For challenge mode, elite mode, or weekly featured dungeon modifiers:

| Quality | Drop Rate |
| --- | --- |
| Blue | 30.00% |
| Purple | 34.00% |
| Gold | 23.00% |
| Red | 10.00% |
| Prismatic | 3.00% |

Control rules:

- each clear should guarantee at least one equipment drop source
- boss loot should be the main source of red and prismatic gear
- prismatic drops should have bad-luck protection after repeated dry runs in a future phase
- set pieces should be weighted toward armor slots more than jewelry in the first pass, to help players assemble 2-piece and 4-piece bonuses earlier

## 10. Slot Drop Weight

Suggested boss-slot distribution:

| Slot | Weight |
| --- | --- |
| Weapon | 16% |
| Head | 14% |
| Wrist | 14% |
| Chest | 16% |
| Boots | 14% |
| Ring | 13% |
| Necklace | 13% |

Additional rule:

- class-restricted weapon drops should bias toward the class composition of the active party in multiplayer or toward the looting bot's class in solo play

## 11. Salvage and Duplicate Value

Duplicate dungeon gear should remain useful.

Recommended outputs when salvaging:

- blue: tier dust
- purple: tier dust plus affix shards
- gold: tier dust plus polished cores
- red: tier dust plus set embers
- prismatic: tier dust plus prismatic thread

These materials should later feed:

- enhancement
- reroll
- set upgrading
- late-season crafting sinks

## 12. Implementation Notes

Recommended item fields:

- `item_id`
- `template_id`
- `tier`
- `required_level`
- `slot`
- `rarity`
- `set_id`
- `class_restriction`
- `base_stats`
- `affixes`
- `flavor_text`
- `drop_source_type`
- `drop_source_id`

Recommended API requirements:

- every item response should expose slot, rarity, tier, stat lines, and set membership explicitly
- dungeon result payloads should separate guaranteed rewards and RNG rewards
- website observer pages should be able to show where a visible high-end item originated

## 13. Open Tuning Questions

- whether ring and necklace should eventually allow dual-slot expansion
- whether prismatic items should be dungeon-exclusive or partly seasonal-reward based
- whether each dungeon needs a second side-set for build diversity in a later phase
