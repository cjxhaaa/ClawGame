# Class Skill System

## 1. Design Goals

This document defines the class skill system used by dungeon auto-combat.

Core rules:

- the game provides `6` universal skills that any character may use
- each class now has `3` clear combat tracks for skill categorization and recommended builds
- each class also keeps a small class-specific shared skill pool
- each dungeon run may equip up to `4` active skills
- the basic attack is always available and does not consume a slot
- skills do not consume MP and are gated only by cooldowns
- dungeon combat is fully auto-acted using explainable priority rules
- characters start as `civilian` and choose one profession route at level `10`
- players and bots should be able to mix unlocked universal and class skills freely rather than being locked to one route id
- non-basic skills are unlocked directly through gold-based upgrades at the Adventurers Guild
- skill upgrades consume gold and increase skill effect values by a controlled percentage
- skill upgrades use a fixed `10`-level cap instead of a character-level scaling cap

The three routes per class are:

- Warrior: `tank`, `physical_burst`, `magic_burst`
- Mage: `single_burst`, `aoe_burst`, `control`
- Priest: `healing_support`, `curse`, `summon`

## 2. Skill Shape

Every skill should define:

- `skill_id`
- `name`
- `class`
- `route_id`
- `weapon_style`
- `action_type`
- `target_type`
- `damage_type`
- `skill_tier`
- `cooldown_rounds`
- `priority`
- `skill_power`
- `atk_ratio`
- `def_ratio`
- `heal_ratio`
- `shield_ratio`
- `status_payloads`
- `ai_condition`

Track rule:

- the old idea of a hard `route_id` should be softened into a classification tag or `skill_track`
- tracks exist to describe intended playstyle and help bot recommendations
- tracks should not hard-lock what skills may be equipped together
- universal skills may use `class = universal` and `route_id = universal`
- `weapon_style` may be used as an optional flavor or compatibility tag later, but it should not be the main limiter of build freedom in this version

## 2.1 Localization Display Names

UI, battle logs, and bot summaries should expose localized display names while keeping English ids for data tables and backend responses.

### Route display names

| class | route_id | Chinese display name |
| --- | --- | --- |
| Warrior | `tank` | 坦克 |
| Warrior | `physical_burst` | 物理爆发 |
| Warrior | `magic_burst` | 魔法爆发 |
| Mage | `single_burst` | 单体爆发 |
| Mage | `aoe_burst` | 群体爆发 |
| Mage | `control` | 控场 |
| Priest | `healing_support` | 治疗辅助 |
| Priest | `curse` | 诅咒 |
| Priest | `summon` | 召唤 |

### Skill display names

| skill_id | Chinese display name |
| --- | --- |
| `Strike` | 重击 |
| `Quickstep` | 迅步 |
| `Pocket Sand` | 扬沙 |
| `Emergency Roll` | 紧急翻滚 |
| `Signal Flare` | 信号照明弹 |
| `Field Tonic` | 野战提神剂 |
| `Tripwire Kit` | 绊索装置 |
| `Guard Stance` | 守御姿态 |
| `War Cry` | 战吼 |
| `Intercept` | 拦截 |
| `Shield Bash` | 盾击 |
| `Fortified Slash` | 固守斩 |
| `Bulwark Field` | 壁垒领域 |
| `Linebreaker` | 裂阵斩 |
| `Cleave` | 劈斩 |
| `Blood Roar` | 血怒咆哮 |
| `Execution Rush` | 处决突进 |
| `Rending Arc` | 裂刃圆弧 |
| `Runic Brand` | 符文烙印 |
| `Arcsteel Surge` | 奥钢奔流 |
| `Spellrend Wave` | 裂法波 |
| `Astral Breaker` | 星界破灭 |
| `Arc Bolt` | 奥术飞弹 |
| `Arc Veil` | 奥术帷幕 |
| `Focus Pulse` | 聚能脉冲 |
| `Disrupt Ray` | 扰乱射线 |
| `Hex Mark` | 咒印 |
| `Seal Fracture` | 封印裂解 |
| `Detonate Sigil` | 引爆秘印 |
| `Star Pierce` | 星芒穿刺 |
| `Flame Burst` | 烈焰爆裂 |
| `Meteor Shard` | 陨星碎片 |
| `Chain Script` | 连锁咒文 |
| `Ember Field` | 余烬领域 |
| `Frost Bind` | 寒霜束缚 |
| `Gravity Knot` | 引力结 |
| `Silencing Prism` | 缄默棱镜 |
| `Time Lock` | 时滞锁 |
| `Smite` | 惩击 |
| `Restore` | 恢复术 |
| `Sanctuary Mark` | 圣护印记 |
| `Purge` | 净除 |
| `Grace Field` | 恩泽领域 |
| `Purifying Wave` | 净化之潮 |
| `Prayer of Renewal` | 复苏祷言 |
| `Bless Armor` | 圣佑护甲 |
| `Judged Weakness` | 弱点审判 |
| `Seal of Silence` | 沉默封印 |
| `Wither Prayer` | 枯败祷言 |
| `Judgment` | 裁决 |
| `Sanctified Blow` | 圣化打击 |
| `Lantern Servitor` | 灯灵侍从 |
| `Censer Guardian` | 香炉守卫 |
| `Choir Invocation` | 圣咏召唤 |

## 3. Loadout Rules

- each dungeon entry may equip up to `4` active skills
- the basic attack is always available and does not consume a slot
- the same skill cannot be equipped twice
- equipped skills may be mixed from unlocked universal skills and any unlocked class-specific skills
- classes should not be hard-locked to a single route before a dungeon
- bots should be able to save multiple free-mix loadout presets
- routes remain recommended build labels, not equip restrictions

Cooldown tier rules:

- `normal`: `1` round cooldown
- `advanced`: `2` round cooldown
- `ultimate`: `3` round cooldown
- basic attack: `0` round cooldown

## 4. Skill Unlock And Upgrade Rules

### 4.1 Civilian Stage and Unlock Rules

- the basic attack is innate and available at character creation
- every new character starts as a `civilian`
- all non-basic skills start at level `0`
- a skill at level `0` is locked and cannot be equipped or used
- upgrading a skill from level `0` to level `1` unlocks that skill for use
- skill unlocks and upgrades are handled at the `Adventurers Guild`
- unlock access should respect career stage, class ownership, and level requirements

Recommended unlock structure:

| Stage | Unlockable content |
| --- | --- |
| character creation | basic attack only |
| civilian level `1-9` | the `6` universal skills may be unlocked to level `1` |
| level `10+` after profession choice | class-specific skills for the chosen class may be unlocked |

Design rule:

- once a universal or class skill reaches level `1`, it may be mixed freely with other unlocked skills that the character is allowed to use
- route labels should still exist for UI guidance, bot recommendation, and balancing analysis

### 4.2 Profession Route Choice

When a civilian reaches level `10`, it must choose one profession route at the Adventurers Guild before unlocking any class-specific skills.

Profession-choice rules:

- the profession route determines the character's class identity
- the choice is intended to be stable for the season in V1
- the route grants one recommended starter weapon family at promotion time
- route labels guide onboarding and bot planning, but do not hard-lock later skill loadouts

Route-to-class mapping:

| Route | Class | Recommended starter weapon |
| --- | --- | --- |
| `tank` | `warrior` | `sword_shield` |
| `physical_burst` | `warrior` | `great_axe` |
| `magic_burst` | `warrior` | `sword_shield` |
| `single_burst` | `mage` | `spellbook` |
| `aoe_burst` | `mage` | `staff` |
| `control` | `mage` | `spellbook` |
| `healing_support` | `priest` | `holy_tome` |
| `curse` | `priest` | `scepter` |
| `summon` | `priest` | `holy_tome` |

### 4.3 Upgrade Rules

- every skill has its own upgrade level
- upgrading a skill costs `gold`
- each skill level increases the skill's effective output by a controlled percentage
- the skill-level cap is fixed and does not scale with character level

Skill-cap rule:

- every non-basic skill starts at level `0`
- level `0` means locked and unusable
- upgrading to level `1` unlocks the skill
- every skill may be upgraded up to level `10`

Recommended upgrade effect:

- each skill level grants `+2%` linear growth to that skill's core effect value
- "core effect value" means damage, healing, shield strength, summon output, or the main numeric payload of the skill
- the recommended formula is:
  - `effective_value = base_value * (1 + 0.02 * skill_level)`
- example:
  - if a skill deals damage equal to `50%` of attack, then at skill level `1` it should deal `50% * 1.02`
  - the same skill at skill level `10` should deal `50% * 1.20`
- control duration and hard crowd-control chance should generally not scale linearly with every level unless explicitly tagged to do so

Economy rule:

- gold is the default upgrade currency in V1
- upgrade costs should scale by current skill level
- skill-upgrade cost should not depend on skill tier in this version

Recommended gold-cost table for upgrading to the next level:

| Upgrade target | Gold cost |
| --- | --- |
| `0 -> 1` | `120` |
| `1 -> 2` | `180` |
| `2 -> 3` | `300` |
| `3 -> 4` | `480` |
| `4 -> 5` | `750` |
| `5 -> 6` | `1140` |
| `6 -> 7` | `1680` |
| `7 -> 8` | `2400` |
| `8 -> 9` | `3360` |
| `9 -> 10` | `4620` |

Balance intent:

- all skills should follow the same upgrade curve so build choice is shaped by gameplay role, not by hidden cost bias
- unlocking a new skill should be cheaper than pushing an already-used skill deeper into the curve
- early levels should still feel affordable enough for broad experimentation
- later levels should remain a meaningful long-tail gold sink
- the total cost to max one skill is intentionally meaningful, but should remain lower than full endgame gear optimization

## 5. Universal Skill Pool

Universal skills are available to all characters and form the entire non-basic skill pool during the civilian stage.

### 5.1 Universal skills

- `Quickstep`
  - type: `self_buff`
  - target: `self`
  - effect: gain `speed_up` and `accuracy_up` for 2 rounds
  - tier: `normal`
  - cooldown: `1`

- `Pocket Sand`
  - type: `utility_attack`
  - target: `single_enemy`
  - effect: deal light physical damage and apply `accuracy_down` for 2 rounds
  - tier: `normal`
  - cooldown: `1`

- `Emergency Roll`
  - type: `defensive_mobility`
  - target: `self`
  - effect: gain `evasion_up` and heavy reduction against the next single-target hit this round
  - tier: `advanced`
  - cooldown: `2`

- `Signal Flare`
  - type: `battlefield_setup`
  - target: `all_enemies`
  - effect: apply `revealed` for 2 rounds so allied attacks gain a small hit-bonus against affected targets
  - tier: `advanced`
  - cooldown: `2`

- `Field Tonic`
  - type: `tempo_support`
  - target: `self`
  - effect: reduce one equipped skill cooldown by `1` and gain a short `speed_up`
  - tier: `advanced`
  - cooldown: `2`

- `Tripwire Kit`
  - type: `trap`
  - target: `single_enemy`
  - effect: place a trap that triggers on the target's next action, dealing light damage and applying `speed_down`
  - tier: `ultimate`
  - cooldown: `3`

Design intent:

- universal skills should feel practical and flavorful without replacing class-defining skills
- they should provide onboarding utility, tempo control, and light setup rather than hard identity payoffs
- they should remain situationally useful even after profession choice

## 6. Auto-battle Selection Logic

Player-side entities choose actions in this order:

1. survival skills if self HP is low
2. healing, shielding, or summon-refresh if allies are endangered
3. AoE or battlefield-control skills when multiple enemies are active
4. boss-specific break, curse, or route payoff skills when conditions are met
5. highest-value damage or summon-empower skills if available
6. basic attack if everything is on cooldown

## 7. Warrior Skill System

Warrior routes are built around frontline control, physical execution, and magic-infused burst.

### 7.1 Shared Basic Attack

- `Strike`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline physical attack
  - tier: `basic`
  - cooldown: `0`

### 7.2 Warrior Shared Active Skills

- `Guard Stance`
  - type: `shield`
  - target: `self`
  - effect: gain a max-HP-based shield and 2 rounds of `defense_up`
  - tier: `advanced`
  - cooldown: `2`

- `War Cry`
  - type: `buff`
  - target: `all_allies`
  - effect: team gains 2 rounds of `attack_up`
  - tier: `ultimate`
  - cooldown: `3`

- `Intercept`
  - type: `debuff`
  - target: `single_enemy`
  - effect: light damage plus strong taunt threat
  - tier: `normal`
  - cooldown: `1`

### 7.3 Tank Track

Route identity:

- front-line stability
- enemy control
- shield and defense layering

Route skills:

- `Shield Bash`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-low physical damage, 30% chance to apply `stun`
  - tier: `normal`
  - cooldown: `1`

- `Fortified Slash`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium physical damage plus 2 rounds of `defense_up` to self
  - tier: `advanced`
  - cooldown: `2`

- `Bulwark Field`
  - type: `shield`
  - target: `all_allies`
  - effect: small team-wide shield
  - tier: `ultimate`
  - cooldown: `3`

- `Linebreaker`
  - type: `skill_attack`
  - target: `front_enemy`
  - effect: strong front-target hit plus `defense_down`
  - tier: `advanced`
  - cooldown: `2`

### 7.4 Physical Burst Track

Route identity:

- direct physical damage
- execution windows
- speed and attack spikes

Route skills:

- `Cleave`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: medium AoE physical damage
  - tier: `normal`
  - cooldown: `1`

- `Blood Roar`
  - type: `buff`
  - target: `self`
  - effect: 2 rounds of `attack_up` and `speed_up`
  - tier: `advanced`
  - cooldown: `2`

- `Execution Rush`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: high single-target damage with bonus damage versus low-HP enemies
  - tier: `ultimate`
  - cooldown: `3`

- `Rending Arc`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: medium-low AoE damage plus `attack_down`
  - tier: `ultimate`
  - cooldown: `3`

### 7.5 Magic Burst Track

Route identity:

- magic-infused blade attacks
- anti-armor pressure
- burst windows through magic conversion

Route skills:

- `Runic Brand`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `magic_vulnerable`
  - tier: `normal`
  - cooldown: `1`

- `Arcsteel Surge`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-high magic damage based on weapon force
  - tier: `advanced`
  - cooldown: `2`

- `Spellrend Wave`
  - type: `skill_attack`
  - target: `front_enemy`
  - effect: front-row magic slash plus `magic_defense_down`
  - tier: `advanced`
  - cooldown: `2`

- `Astral Breaker`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: very high magic burst against branded or debuffed targets
  - tier: `ultimate`
  - cooldown: `3`

## 8. Mage Skill System

Mage routes are built around precision nukes, large-scale AoE pressure, and battlefield control.

### 8.1 Shared Basic Attack

- `Arc Bolt`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline magic attack
  - tier: `basic`
  - cooldown: `0`

### 8.2 Mage Shared Active Skills

- `Arc Veil`
  - type: `shield`
  - target: `self`
  - effect: gain a spell shield and 2 rounds of `magic_defense_up`
  - tier: `advanced`
  - cooldown: `2`

- `Focus Pulse`
  - type: `buff`
  - target: `self`
  - effect: 2 rounds of `magic_attack_up`
  - tier: `normal`
  - cooldown: `1`

- `Disrupt Ray`
  - type: `debuff`
  - target: `single_enemy`
  - effect: medium damage plus `speed_down`
  - tier: `advanced`
  - cooldown: `2`

### 8.3 Single Burst Track

Route identity:

- single-target deletion
- debuff-assisted nukes
- boss burn windows

Route skills:

- `Hex Mark`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `vulnerable`
  - tier: `normal`
  - cooldown: `1`

- `Seal Fracture`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `magic_defense_down`
  - tier: `advanced`
  - cooldown: `2`

- `Detonate Sigil`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: high burst damage against already-debuffed targets
  - tier: `ultimate`
  - cooldown: `3`

- `Star Pierce`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: heavy single-target magic damage with bonus crit rate
  - tier: `advanced`
  - cooldown: `2`

### 8.4 AoE Burst Track

Route identity:

- room clearing
- burn and splash damage
- fast backline wipe pressure

Route skills:

- `Flame Burst`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: AoE magic damage with a chance to apply `burn`
  - tier: `advanced`
  - cooldown: `2`

- `Meteor Shard`
  - type: `skill_attack`
  - target: `front_enemy`
  - effect: high front-target damage
  - tier: `ultimate`
  - cooldown: `3`

- `Chain Script`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: low AoE damage with bonus damage against debuffed targets
  - tier: `advanced`
  - cooldown: `2`

- `Ember Field`
  - type: `debuff`
  - target: `all_enemies`
  - effect: applies `vulnerable` for 2 rounds
  - tier: `ultimate`
  - cooldown: `3`

### 8.5 Control Track

Route identity:

- action denial
- speed manipulation
- boss phase disruption

Route skills:

- `Frost Bind`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium damage plus `speed_down`
  - tier: `normal`
  - cooldown: `1`

- `Gravity Knot`
  - type: `debuff`
  - target: `all_enemies`
  - effect: applies short `slow` and mild `accuracy_down`
  - tier: `advanced`
  - cooldown: `2`

- `Silencing Prism`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `silence` or cast-delay pressure
  - tier: `advanced`
  - cooldown: `2`

- `Time Lock`
  - type: `debuff`
  - target: `front_enemy`
  - effect: heavy control attempt with large `speed_down` and break utility
  - tier: `ultimate`
  - cooldown: `3`

## 9. Priest Skill System

Priest routes are built around sustain utility, curse application, and creature-based support.

### 9.1 Shared Basic Attack

- `Smite`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline holy damage
  - tier: `basic`
  - cooldown: `0`

### 9.2 Priest Shared Active Skills

- `Restore`
  - type: `heal`
  - target: `single_ally`
  - effect: single-target heal
  - tier: `normal`
  - cooldown: `1`

- `Sanctuary Mark`
  - type: `shield`
  - target: `single_ally`
  - effect: shield one ally and apply 2 rounds of `defense_up`
  - tier: `advanced`
  - cooldown: `2`

- `Purge`
  - type: `cleanse`
  - target: `single_ally`
  - effect: remove one negative status and heal slightly
  - tier: `normal`
  - cooldown: `1`

### 9.3 Healing Support Track

Route identity:

- stable sustain
- cleanse and regen
- teamwide defense support

Route skills:

- `Grace Field`
  - type: `heal`
  - target: `all_allies`
  - effect: applies team-wide `regen`
  - tier: `advanced`
  - cooldown: `2`

- `Purifying Wave`
  - type: `cleanse`
  - target: `all_allies`
  - effect: small team heal and removes one damage-over-time layer
  - tier: `ultimate`
  - cooldown: `3`

- `Prayer of Renewal`
  - type: `heal`
  - target: `all_allies`
  - effect: medium team heal
  - tier: `ultimate`
  - cooldown: `3`

- `Bless Armor`
  - type: `buff`
  - target: `all_allies`
  - effect: apply 2 rounds of `defense_up` to the whole team
  - tier: `advanced`
  - cooldown: `2`

### 9.4 Curse Track

Route identity:

- weakening enemies
- anti-boss utility
- holy damage payoff against cursed targets

Route skills:

- `Judged Weakness`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `attack_down`
  - tier: `advanced`
  - cooldown: `2`

- `Seal of Silence`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `silence`
  - tier: `advanced`
  - cooldown: `2`

- `Wither Prayer`
  - type: `debuff`
  - target: `all_enemies`
  - effect: applies short `healing_down` and mild `defense_down`
  - tier: `ultimate`
  - cooldown: `3`

- `Judgment`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-high holy damage with bonus damage versus cursed or debuffed targets
  - tier: `ultimate`
  - cooldown: `3`

### 9.5 Summon Track

Route identity:

- calls support entities
- creates persistent value across rounds
- mixes minor healing, shielding, and damage through summons

Route skills:

- `Sanctified Blow`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-low holy damage and a small heal to the lowest-HP ally
  - tier: `normal`
  - cooldown: `1`

- `Lantern Servitor`
  - type: `summon`
  - target: `self_side`
  - effect: summons a support spirit that provides minor healing each round
  - tier: `advanced`
  - cooldown: `2`

- `Censer Guardian`
  - type: `summon`
  - target: `self_side`
  - effect: summons a guardian construct that grants small shields or intercept value
  - tier: `advanced`
  - cooldown: `2`

- `Choir Invocation`
  - type: `summon`
  - target: `self_side`
  - effect: summons a major holy entity that pulses heal or holy damage for 2 rounds
  - tier: `ultimate`
  - cooldown: `3`

## 10. Recommended Loadout Examples

### Warrior Tank

- `Guard Stance`
- `Intercept`
- `Shield Bash`
- `Bulwark Field`

### Warrior Physical Burst

- `War Cry`
- `Cleave`
- `Execution Rush`
- `Blood Roar`

### Warrior Magic Burst

- `Guard Stance`
- `Runic Brand`
- `Arcsteel Surge`
- `Astral Breaker`

### Mage Single Burst

- `Focus Pulse`
- `Hex Mark`
- `Seal Fracture`
- `Detonate Sigil`

### Mage AoE Burst

- `Focus Pulse`
- `Flame Burst`
- `Chain Script`
- `Ember Field`

### Mage Control

- `Arc Veil`
- `Frost Bind`
- `Gravity Knot`
- `Time Lock`

### Priest Healing Support

- `Restore`
- `Grace Field`
- `Purifying Wave`
- `Bless Armor`

### Priest Curse

- `Purge`
- `Judged Weakness`
- `Seal of Silence`
- `Judgment`

### Priest Summon

- `Restore`
- `Lantern Servitor`
- `Censer Guardian`
- `Choir Invocation`

## 11. Fit With Dungeon Tempo

Because a dungeon battle is capped at `10` rounds:

- cooldowns must stay within the `1/2/3` tier structure
- meaningful track differentiation should appear by rounds `3-5`
- AoE, sustain, control, summon presence, and survival all need to matter in short fights
- boss pacing should revolve around `2` and `3` round cooldowns

## 12. Open Follow-up Decisions

- whether skill-track labels should be visible as a separate recommendation field in bot summaries and observer UI
- whether summons occupy explicit battlefield slots or use an abstract support-entity layer
- whether some future advanced skills should be shared across two tracks of the same class
