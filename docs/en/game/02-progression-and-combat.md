## 6. Progression Systems

### 6.1 Classes

#### Warrior

- role: front-line physical attacker
- strengths: high HP, high defense, stable single-target damage
- weakness: low AoE and limited sustain
- weapon types:
  - Sword and Shield
  - Great Axe

#### Mage

- role: burst magic attacker
- strengths: high AoE, ranged damage, debuffs
- weakness: low HP and weaker sustained defense
- weapon types:
  - Staff
  - Spellbook

#### Priest

- role: sustain and utility specialist
- strengths: healing, cleansing, long-fight stability
- weakness: lower direct damage
- weapon types:
  - Scepter
  - Holy Tome

### 6.2 Reputation

Reputation is now a spendable contract currency, not a rank ladder.

Rules:

- reputation is earned mainly from daily contract submission
- reputation no longer unlocks regions, dungeons, or potion tiers
- every character receives `2` free dungeon reward claims each day
- one extra dungeon reward claim costs `50` reputation
- purchased extra claims reset each day, while unspent reputation remains on the character
- the legacy `rank` field may still exist in payloads for compatibility, but it is no longer an authoritative progression gate

### 6.3 Character level

The baseline V1 spec originally did not use a separate XP level system.

This is now superseded by the seasonal progression proposal in:

- `08-seasonal-leveling-and-stat-framework.md`

The new direction is:

- every season lasts 30 days
- season level runs from `1` to `100`
- level-based stat growth stops at max level
- the late-season loop shifts toward materials and equipment
- reputation remains a parallel economy track instead of an access-gating axis

## 7. Stats and Combat

This section now serves as a compact overview only.

The detailed battle rules have moved to:

- `10-combat-system-framework.md`

### 7.1 Core stats

Every character and enemy has:

- `max_hp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`

Optional battle metadata:

- `status_effects`
- `cooldowns`
- `shield_value`

### 7.2 Combat model

- combat is fully turn-based
- all combat is server-authoritative
- no free movement inside combat
- turn order is descending by `speed`
- ties are resolved by lower `entity_id`

### 7.3 Hit and randomness rules

To remain bot-friendly:

- standard attacks have `100%` hit chance
- skills define their hit chance explicitly
- no hidden dodge stat in V1
- no random critical hits from basic attacks
- damage variance is fixed at `+/- 5%`

### 7.4 Damage formula

V1 uses a transparent formula family:

- physical damage: `max(1, skill_power + actor.physical_attack * atk_ratio - target.physical_defense * def_ratio)`
- magic damage: `max(1, skill_power + actor.magic_attack * atk_ratio - target.magic_defense * def_ratio)`
- healing: `max(1, skill_power + actor.healing_power * heal_ratio)`

The resolved battle log must always include:

- acting entity
- action name
- target entity
- raw effect type
- final effect amount
- statuses applied or removed

### 7.5 V1 statuses

- `poison`: fixed damage at end of turn
- `burn`: fixed damage at end of turn
- `stun`: skip next turn
- `shielded`: absorbs damage first
- `regen`: fixed healing at end of turn
- `silence`: cannot use magic-tagged skills

No hidden stacking rules:

- all statuses define duration in turns
- same status refreshes duration unless marked stackable

## 8. Class Skill Kits

This section now serves as a compact overview only.

The detailed skill specification has moved to:

- `11-class-skill-system.md`

The new direction is:

- each class and weapon style has a broader skill pool
- each dungeon loadout may equip up to `4` active skills
- the basic attack is always available and has no cooldown
- skills no longer consume MP and are gated only by cooldowns
- battle actions are selected automatically from the configured loadout
