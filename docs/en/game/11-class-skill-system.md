# Class Skill System

## 1. Overview

This document defines the class skill system used by dungeon auto-combat.

### 1.1 Core Rules

| Topic | Rule |
| --- | --- |
| Starting state | Every new character begins as `civilian` with a basic attack, the civilian skill pool, and the class-common skills of warrior, mage, and priest. |
| Class unlock | Profession change becomes available at level `10`. |
| Equip limit | Each dungeon run may equip up to `4` active skills. |
| Basic attack | Basic attack is always available and does not consume an active slot. |
| Resource model | Skills do not consume MP and are gated only by cooldowns. |
| Availability rule | Skill usability depends only on the current class. |
| Starting level | Every non-basic skill starts at level `1`. |
| Upgrade model | Skill upgrades consume gold and only increase effect strength. |
| Level cap | Every non-basic skill has a fixed cap of level `10`. |
| Build freedom | Track tags are for categorization and recommendations only, never hard equip restrictions. |

### 1.2 Skill Layers By Class

| Current class | Available non-basic skill layers |
| --- | --- |
| `civilian` | Civilian skills + warrior class-common skills + mage class-common skills + priest class-common skills |
| `warrior` | Civilian skills + warrior class-common skills + warrior class-specific skills |
| `mage` | Civilian skills + mage class-common skills + mage class-specific skills |
| `priest` | Civilian skills + priest class-common skills + priest class-specific skills |

### 1.3 Internal Track Map

| Class | Track 1 | Track 2 | Track 3 |
| --- | --- | --- | --- |
| `warrior` | `tank` | `physical_burst` | `magic_burst` |
| `mage` | `single_burst` | `aoe_burst` | `control` |
| `priest` | `healing_support` | `curse` | `summon` |

## 2. Skill Data Shape

### 2.1 Required Fields

| Field | Purpose |
| --- | --- |
| `skill_id` | Stable English identifier used by backend and data tables |
| `name` | English display name |
| `class` | Owning layer: `civilian`, `warrior`, `mage`, or `priest` |
| `route_id` | Track classification field |
| `weapon_style` | Optional build-profile tag |
| `action_type` | Main combat action category |
| `target_type` | Main targeting pattern |
| `damage_type` | Damage family when applicable |
| `skill_tier` | Cooldown tier: `basic`, `normal`, `advanced`, `ultimate` |
| `cooldown_rounds` | Cooldown in rounds |
| `priority` | AI priority hint |
| `skill_power` | Main numeric effect scale |
| `atk_ratio` | Attack coefficient |
| `def_ratio` | Defense coefficient |
| `heal_ratio` | Healing coefficient |
| `shield_ratio` | Shield coefficient |
| `status_payloads` | Applied statuses or debuffs |
| `ai_condition` | AI usage condition |

### 2.2 Classification Rules

| Topic | Rule |
| --- | --- |
| `route_id` | Describes intended playstyle and helps bot recommendations |
| Track restrictions | Tracks never hard-lock learning, upgrading, or equipping |
| Civilian layer | Civilian skills use `class = civilian` and `route_id = civilian` |
| Profession layers | Profession skills use `class = warrior | mage | priest` |
| Weapon style | `weapon_style` may exist as a profile tag, but does not gate skill access |

### 2.3 Localization Display Names

#### Track display names

| Class | `route_id` | Chinese display name |
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

#### Skill display names

| `skill_id` | Chinese display name |
| --- | --- |
| `Strike` | 重击 |
| `Arc Bolt` | 奥术飞弹 |
| `Smite` | 惩击 |
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

## 3. Loadout And Availability Rules

### 3.1 Loadout Rules

| Topic | Rule |
| --- | --- |
| Active slots | Up to `4` active skills |
| Basic attack | Does not consume a slot |
| Duplicate equip | The same skill cannot be equipped twice |
| Civilian loadout | Civilian skills + all three promoted-class common layers |
| Promoted-class loadout | Civilian skills + current-class common layer + current-class specific skills |
| Track mixing | Tracks do not restrict loadout mixing |
| Bot presets | Bots should support multiple free-mix loadout presets per class |

### 3.2 Cooldown Tier Rules

| Tier | Cooldown |
| --- | --- |
| `basic` | `0` rounds |
| `normal` | `1` round |
| `advanced` | `2` rounds |
| `ultimate` | `3` rounds |

### 3.3 Stage Access Rules

| Stage | Available skill content |
| --- | --- |
| Character creation | Basic attack + all civilian skills at level `1` + all class-common skills at level `1` |
| Civilian level `1-9` | Continue using civilian skills + all class-common skills |
| Level `10+` after profession choice | Civilian skills + chosen-class common skills + chosen-class specific skills |
| Level `10+` while staying civilian | Continue using civilian skills + all class-common skills |

## 4. Progression And Upgrade Rules

### 4.1 Skill Progression Rules

| Topic | Rule |
| --- | --- |
| Initial level | Every non-basic skill starts at level `1` |
| Access persistence | Learned skill levels are preserved across class changes |
| Access check | A skill may be used when it belongs to a currently valid layer |
| Class rollback | Switching back to a previous class restores access to its learned skills |
| Guild role | Skill upgrades are handled at the `Adventurers Guild` |

### 4.2 Profession Change Rules

| Topic | Rule |
| --- | --- |
| Unlock point | Level `10` |
| Allowed classes | `civilian`, `warrior`, `mage`, `priest` |
| Gold cost | `800` gold per class change |
| Skill levels | Preserved and never reset |
| Active loadout | Cleared on class change |
| Bot behavior | Bot should choose a fresh loadout after class change |
| Starter weapon | Entering a promoted class from `civilian` grants a recommended starter weapon family |

### 4.3 Starter Weapon Mapping

| Profession | Recommended starter weapon |
| --- | --- |
| `warrior` | `sword_shield` |
| `mage` | `spellbook` |
| `priest` | `holy_tome` |

### 4.4 Upgrade Rules

| Topic | Rule |
| --- | --- |
| Upgrade currency | `gold` |
| Level cap | `10` |
| Scaling rule | Each level adds `+2%` to the skill's core effect relative to level `1` |
| Suggested formula | `effective_value = base_value * (1 + 0.02 * (skill_level - 1))` |
| Level 1 example | `50% * 1.00` |
| Level 10 example | `50% * 1.18` |
| Non-scaling caution | Hard crowd control and duration should not automatically scale linearly unless explicitly tagged |

### 4.5 Upgrade Cost Table

| Upgrade target | Gold cost |
| --- | --- |
| `1 -> 2` | `180` |
| `2 -> 3` | `300` |
| `3 -> 4` | `480` |
| `4 -> 5` | `750` |
| `5 -> 6` | `1140` |
| `6 -> 7` | `1680` |
| `7 -> 8` | `2400` |
| `8 -> 9` | `3360` |
| `9 -> 10` | `4620` |

## 5. Civilian Layer

### 5.1 Civilian Layer Summary

| Topic | Rule |
| --- | --- |
| Layer owner | `civilian` |
| Availability | Shared by every class |
| Design role | Early onboarding, tempo setup, lightweight utility |
| Post-promotion role | Remains part of active build space inside promoted-class loadouts |

### 5.2 Civilian Skill Table

| Skill | Type | Target | Effect summary | Tier | Cooldown |
| --- | --- | --- | --- | --- | --- |
| `Quickstep` | `self_buff` | `self` | Gain `speed_up` and `accuracy_up` for 2 rounds | `normal` | `1` |
| `Pocket Sand` | `utility_attack` | `single_enemy` | Light physical damage and `accuracy_down` for 2 rounds | `normal` | `1` |
| `Emergency Roll` | `defensive_mobility` | `self` | Gain `evasion_up` and heavy reduction against the next single-target hit this round | `advanced` | `2` |
| `Signal Flare` | `battlefield_setup` | `all_enemies` | Apply `revealed` for 2 rounds so allied attacks gain a hit bonus | `advanced` | `2` |
| `Field Tonic` | `tempo_support` | `self` | Reduce one equipped skill cooldown by `1` and gain short `speed_up` | `advanced` | `2` |
| `Tripwire Kit` | `trap` | `single_enemy` | Place a trap that deals light damage and applies `speed_down` on next enemy action | `ultimate` | `3` |

## 6. Class Skill Tables

### 6.1 Warrior

| Layer | Skill | Type | Target | Effect summary | Tier | Cooldown |
| --- | --- | --- | --- | --- | --- | --- |
| Basic | `Strike` | `basic_attack` | `single_enemy` | Baseline physical attack | `basic` | `0` |
| Common | `Guard Stance` | `shield` | `self` | Gain a max-HP-based shield and 2 rounds of `defense_up` | `advanced` | `2` |
| Common | `War Cry` | `buff` | `all_allies` | Team gains 2 rounds of `attack_up` | `ultimate` | `3` |
| Common | `Intercept` | `debuff` | `single_enemy` | Light damage plus strong taunt threat | `normal` | `1` |
| `tank` | `Shield Bash` | `skill_attack` | `single_enemy` | Medium-low physical damage and 30% `stun` chance | `normal` | `1` |
| `tank` | `Fortified Slash` | `skill_attack` | `single_enemy` | Medium physical damage plus 2 rounds of `defense_up` to self | `advanced` | `2` |
| `tank` | `Bulwark Field` | `shield` | `all_allies` | Small team-wide shield | `ultimate` | `3` |
| `tank` | `Linebreaker` | `skill_attack` | `front_enemy` | Strong front-target hit plus `defense_down` | `advanced` | `2` |
| `physical_burst` | `Cleave` | `skill_attack` | `all_enemies` | Medium AoE physical damage | `normal` | `1` |
| `physical_burst` | `Blood Roar` | `buff` | `self` | 2 rounds of `attack_up` and `speed_up` | `advanced` | `2` |
| `physical_burst` | `Execution Rush` | `skill_attack` | `single_enemy` | High single-target damage with bonus damage vs low-HP enemies | `ultimate` | `3` |
| `physical_burst` | `Rending Arc` | `skill_attack` | `all_enemies` | Medium-low AoE damage plus `attack_down` | `ultimate` | `3` |
| `magic_burst` | `Runic Brand` | `debuff` | `single_enemy` | Apply `magic_vulnerable` | `normal` | `1` |
| `magic_burst` | `Arcsteel Surge` | `skill_attack` | `single_enemy` | Medium-high magic damage based on weapon force | `advanced` | `2` |
| `magic_burst` | `Spellrend Wave` | `skill_attack` | `front_enemy` | Front-row magic slash plus `magic_defense_down` | `advanced` | `2` |
| `magic_burst` | `Astral Breaker` | `skill_attack` | `single_enemy` | Very high magic burst against branded or debuffed targets | `ultimate` | `3` |

### 6.2 Mage

| Layer | Skill | Type | Target | Effect summary | Tier | Cooldown |
| --- | --- | --- | --- | --- | --- | --- |
| Basic | `Arc Bolt` | `basic_attack` | `single_enemy` | Baseline magic attack | `basic` | `0` |
| Common | `Arc Veil` | `shield` | `self` | Gain a spell shield and 2 rounds of `magic_defense_up` | `advanced` | `2` |
| Common | `Focus Pulse` | `buff` | `self` | 2 rounds of `magic_attack_up` | `normal` | `1` |
| Common | `Disrupt Ray` | `debuff` | `single_enemy` | Medium damage plus `speed_down` | `advanced` | `2` |
| `single_burst` | `Hex Mark` | `debuff` | `single_enemy` | Apply `vulnerable` | `normal` | `1` |
| `single_burst` | `Seal Fracture` | `debuff` | `single_enemy` | Apply `magic_defense_down` | `advanced` | `2` |
| `single_burst` | `Detonate Sigil` | `skill_attack` | `single_enemy` | High burst damage against already-debuffed targets | `ultimate` | `3` |
| `single_burst` | `Star Pierce` | `skill_attack` | `single_enemy` | Heavy single-target magic damage with bonus crit rate | `advanced` | `2` |
| `aoe_burst` | `Flame Burst` | `skill_attack` | `all_enemies` | AoE magic damage with a chance to apply `burn` | `advanced` | `2` |
| `aoe_burst` | `Meteor Shard` | `skill_attack` | `front_enemy` | High front-target damage | `ultimate` | `3` |
| `aoe_burst` | `Chain Script` | `skill_attack` | `all_enemies` | Low AoE damage with bonus damage against debuffed targets | `advanced` | `2` |
| `aoe_burst` | `Ember Field` | `debuff` | `all_enemies` | Apply `vulnerable` for 2 rounds | `ultimate` | `3` |
| `control` | `Frost Bind` | `skill_attack` | `single_enemy` | Medium damage plus `speed_down` | `normal` | `1` |
| `control` | `Gravity Knot` | `debuff` | `all_enemies` | Apply short `slow` and mild `accuracy_down` | `advanced` | `2` |
| `control` | `Silencing Prism` | `debuff` | `single_enemy` | Apply `silence` or cast-delay pressure | `advanced` | `2` |
| `control` | `Time Lock` | `debuff` | `front_enemy` | Heavy control attempt with large `speed_down` and break utility | `ultimate` | `3` |

### 6.3 Priest

| Layer | Skill | Type | Target | Effect summary | Tier | Cooldown |
| --- | --- | --- | --- | --- | --- | --- |
| Basic | `Smite` | `basic_attack` | `single_enemy` | Baseline holy damage | `basic` | `0` |
| Common | `Restore` | `heal` | `single_ally` | Single-target heal | `normal` | `1` |
| Common | `Sanctuary Mark` | `shield` | `single_ally` | Shield one ally and apply 2 rounds of `defense_up` | `advanced` | `2` |
| Common | `Purge` | `cleanse` | `single_ally` | Remove one negative status and heal slightly | `normal` | `1` |
| `healing_support` | `Grace Field` | `heal` | `all_allies` | Apply team-wide `regen` | `advanced` | `2` |
| `healing_support` | `Purifying Wave` | `cleanse` | `all_allies` | Small team heal and remove one DoT layer | `ultimate` | `3` |
| `healing_support` | `Prayer of Renewal` | `heal` | `all_allies` | Medium team heal | `ultimate` | `3` |
| `healing_support` | `Bless Armor` | `buff` | `all_allies` | Apply 2 rounds of `defense_up` to the whole team | `advanced` | `2` |
| `curse` | `Judged Weakness` | `debuff` | `single_enemy` | Apply `attack_down` | `advanced` | `2` |
| `curse` | `Seal of Silence` | `debuff` | `single_enemy` | Apply `silence` | `advanced` | `2` |
| `curse` | `Wither Prayer` | `debuff` | `all_enemies` | Apply short `healing_down` and mild `defense_down` | `ultimate` | `3` |
| `curse` | `Judgment` | `skill_attack` | `single_enemy` | Medium-high holy damage with bonus damage vs cursed or debuffed targets | `ultimate` | `3` |
| `summon` | `Sanctified Blow` | `skill_attack` | `single_enemy` | Medium-low holy damage and a small heal to the lowest-HP ally | `normal` | `1` |
| `summon` | `Lantern Servitor` | `summon` | `self_side` | Summon a support spirit that provides minor healing each round | `advanced` | `2` |
| `summon` | `Censer Guardian` | `summon` | `self_side` | Summon a guardian construct that grants small shields or intercept value | `advanced` | `2` |
| `summon` | `Choir Invocation` | `summon` | `self_side` | Summon a major holy entity that pulses heal or holy damage for 2 rounds | `ultimate` | `3` |

## 7. Recommended Loadout Table

| Build | Suggested loadout |
| --- | --- |
| Warrior Tank | `Guard Stance`, `Intercept`, `Shield Bash`, `Bulwark Field` |
| Warrior Physical Burst | `War Cry`, `Cleave`, `Execution Rush`, `Blood Roar` |
| Warrior Magic Burst | `Guard Stance`, `Runic Brand`, `Arcsteel Surge`, `Astral Breaker` |
| Mage Single Burst | `Focus Pulse`, `Hex Mark`, `Seal Fracture`, `Detonate Sigil` |
| Mage AoE Burst | `Focus Pulse`, `Flame Burst`, `Chain Script`, `Ember Field` |
| Mage Control | `Arc Veil`, `Frost Bind`, `Gravity Knot`, `Time Lock` |
| Priest Healing Support | `Restore`, `Grace Field`, `Purifying Wave`, `Bless Armor` |
| Priest Curse | `Purge`, `Judged Weakness`, `Seal of Silence`, `Judgment` |
| Priest Summon | `Restore`, `Lantern Servitor`, `Censer Guardian`, `Choir Invocation` |

## 8. Auto-battle Priority

| Priority | Rule |
| --- | --- |
| `1` | Use survival skills if self HP is low |
| `2` | Use healing, shielding, or summon-refresh if allies are endangered |
| `3` | Use AoE or battlefield control when multiple enemies are active |
| `4` | Use boss-specific break, curse, or class payoff skills when conditions are met |
| `5` | Use the highest-value damage or summon-empower skill available |
| `6` | Fallback to the basic attack if everything is on cooldown |

## 9. Fit With Dungeon Tempo

| Topic | Rule |
| --- | --- |
| Battle length | Dungeon combat is capped at `10` rounds |
| Cooldown structure | Cooldowns should stay inside the `1 / 2 / 3` structure |
| Track clarity | Track differentiation should appear by rounds `3-5` |
| Short-fight value | AoE, sustain, control, summon presence, and survival all need to matter in short fights |
| Boss pacing | Boss timing should revolve around `2` and `3` round cooldowns |

## 10. Open Follow-up Decisions

| Topic | Pending decision |
| --- | --- |
| Track visibility | Whether track labels should be visible as a separate recommendation field in bot summaries and observer UI |
| Summon model | Whether summons occupy explicit battlefield slots or use an abstract support-entity layer |
| Bot weighting | Whether track labels should directly influence bot build recommendations instead of staying presentation-only |
