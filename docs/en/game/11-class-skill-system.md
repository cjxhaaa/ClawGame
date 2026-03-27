# Class Skill System

## 1. Design Goals

This document defines the class and weapon-style skill system used by dungeon auto-combat.

Core rules:

- each class has a shared skill pool
- each weapon style has its own exclusive skill pool
- each dungeon run may equip up to `4` active skills
- the basic attack is always available and does not consume a slot
- skills do not consume MP and are gated only by cooldowns
- dungeon combat is fully auto-acted using explainable priority rules

## 2. Skill Shape

Every skill should define:

- `skill_id`
- `name`
- `class`
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

## 3. Loadout Rules

- each dungeon entry may equip up to `4` active skills
- the basic attack is always available and does not consume a slot
- the same skill cannot be equipped twice
- shared and weapon-exclusive skills can be mixed
- a skill cannot be equipped if it does not match the current weapon style
- bots should be able to save multiple loadout presets

Cooldown tier rules:

- `normal`: `1` round cooldown
- `advanced`: `2` round cooldown
- `ultimate`: `3` round cooldown
- basic attack: `0` round cooldown

## 4. Auto-battle Selection Logic

Player-side entities choose actions in this order:

1. survival skills if self HP is low
2. healing or shielding if allies are endangered
3. AoE or battlefield-control skills when multiple enemies are active
4. boss-specific break or debuff skills when conditions are met
5. high-value damage skills if available
6. basic attack if everything is on cooldown

## 5. Warrior Skill System

### 5.1 Shared Basic Attack

- `Strike`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline physical attack
  - tier: `basic`
  - cooldown: `0`

### 5.2 Warrior Shared Active Skills

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

### 5.3 Sword and Shield Exclusive Skills

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

### 5.4 Great Axe Exclusive Skills

- `Cleave`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: medium AoE physical damage
  - tier: `normal`
  - cooldown: `1`

- `Execution Rush`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: high single-target damage with bonus damage versus low-HP enemies
  - tier: `ultimate`
  - cooldown: `3`

- `Blood Roar`
  - type: `buff`
  - target: `self`
  - effect: 2 rounds of `attack_up` and `speed_up`
  - tier: `advanced`
  - cooldown: `2`

- `Rending Arc`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: medium-low AoE damage plus `attack_down`
  - tier: `ultimate`
  - cooldown: `3`

## 6. Mage Skill System

### 6.1 Shared Basic Attack

- `Arc Bolt`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline magic attack
  - tier: `basic`
  - cooldown: `0`

### 6.2 Mage Shared Active Skills

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

### 6.3 Staff Exclusive Skills

- `Flame Burst`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: AoE magic damage with a chance to apply `burn`
  - tier: `advanced`
  - cooldown: `2`

- `Frost Bind`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium damage plus `speed_down`
  - tier: `normal`
  - cooldown: `1`

- `Meteor Shard`
  - type: `skill_attack`
  - target: `front_enemy`
  - effect: high front-target damage
  - tier: `ultimate`
  - cooldown: `3`

- `Ember Field`
  - type: `debuff`
  - target: `all_enemies`
  - effect: applies `vulnerable` for 2 rounds
  - tier: `ultimate`
  - cooldown: `3`

### 6.4 Spellbook Exclusive Skills

- `Hex Mark`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `vulnerable`
  - tier: `normal`
  - cooldown: `1`

- `Detonate Sigil`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: high burst damage against already-debuffed targets
  - tier: `ultimate`
  - cooldown: `3`

- `Chain Script`
  - type: `skill_attack`
  - target: `all_enemies`
  - effect: low AoE damage with bonus damage against debuffed targets
  - tier: `advanced`
  - cooldown: `2`

- `Seal Fracture`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `magic_defense_down`
  - tier: `advanced`
  - cooldown: `2`

## 7. Priest Skill System

### 7.1 Shared Basic Attack

- `Smite`
  - type: `basic_attack`
  - target: `single_enemy`
  - effect: baseline holy damage
  - tier: `basic`
  - cooldown: `0`

### 7.2 Priest Shared Active Skills

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

### 7.3 Scepter Exclusive Skills

- `Sanctified Blow`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-low holy damage and a small heal to the lowest-HP ally
  - tier: `normal`
  - cooldown: `1`

- `Purifying Wave`
  - type: `cleanse`
  - target: `all_allies`
  - effect: small team heal and removes one damage-over-time layer
  - tier: `ultimate`
  - cooldown: `3`

- `Grace Field`
  - type: `heal`
  - target: `all_allies`
  - effect: applies team-wide `regen`
  - tier: `advanced`
  - cooldown: `2`

- `Judged Weakness`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `attack_down`
  - tier: `advanced`
  - cooldown: `2`

### 7.4 Holy Tome Exclusive Skills

- `Prayer of Renewal`
  - type: `heal`
  - target: `all_allies`
  - effect: medium team heal
  - tier: `ultimate`
  - cooldown: `3`

- `Seal of Silence`
  - type: `debuff`
  - target: `single_enemy`
  - effect: applies `silence`
  - tier: `advanced`
  - cooldown: `2`

- `Bless Armor`
  - type: `buff`
  - target: `all_allies`
  - effect: apply 2 rounds of `defense_up` to the whole team
  - tier: `advanced`
  - cooldown: `2`

- `Judgment`
  - type: `skill_attack`
  - target: `single_enemy`
  - effect: medium-high holy damage with bonus damage versus debuffed targets
  - tier: `ultimate`
  - cooldown: `3`

## 8. Recommended Loadout Examples

### Sword and Shield Warrior

- `Guard Stance`
- `Intercept`
- `Shield Bash`
- `Bulwark Field`

### Great Axe Warrior

- `War Cry`
- `Cleave`
- `Execution Rush`
- `Blood Roar`

### Staff Mage

- `Focus Pulse`
- `Flame Burst`
- `Frost Bind`
- `Ember Field`

### Spellbook Mage

- `Hex Mark`
- `Detonate Sigil`
- `Chain Script`
- `Seal Fracture`

### Scepter Priest

- `Restore`
- `Purifying Wave`
- `Grace Field`
- `Judged Weakness`

### Holy Tome Priest

- `Restore`
- `Prayer of Renewal`
- `Seal of Silence`
- `Bless Armor`

## 9. Fit With Dungeon Tempo

Because a dungeon battle is capped at `10` rounds:

- cooldowns must stay within the `1/2/3` tier structure
- meaningful differentiation should appear by rounds `3-5`
- AoE, sustain, control, and survival all need to matter in short fights
- boss pacing should revolve around `2` and `3` round cooldowns

## 10. Open Follow-up Decisions

- whether some advanced skills should be shared across both weapon styles of a class
- whether future skills should gain dungeon-specialization tags
- whether bots should be allowed to auto-switch saved loadouts based on party composition
