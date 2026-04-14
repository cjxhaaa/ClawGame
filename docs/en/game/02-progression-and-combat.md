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

Reputation is a spendable contract currency used for daily progression pacing.

Rules:

- reputation is earned mainly from daily contract submission
- reputation is used for contract rewards and extra dungeon reward-claim purchases
- every character receives `2` free dungeon reward claims each day
- one extra dungeon reward claim costs `50` reputation
- purchased extra claims reset each day, while unspent reputation remains on the character

### 6.3 Character level

Character level progression is defined in:

- `08-seasonal-leveling-and-stat-framework.md`

Key points:

- every season lasts 30 days
- season level runs from `1` to `100`
- level-based stat growth stops at max level
- the late-season loop shifts toward materials and equipment
- reputation works as a parallel economy track alongside gold and season-level progression

## 7. Stats and Combat

This section serves as a compact overview.

The detailed battle rules have moved to:

- `10-combat-system-framework.md`

### 7.1 Core stats

Current core combat stats are:

| Field | Name | Meaning |
| --- | --- | --- |
| `max_hp` | Max HP | Maximum health pool. |
| `physical_attack` | Physical Attack | Main offensive stat for physical damage skills and attacks. |
| `magic_attack` | Magic Attack | Main offensive stat for magic damage skills and attacks. |
| `physical_defense` | Physical Defense | Main mitigation stat against physical damage. |
| `magic_defense` | Magic Defense | Main mitigation stat against magic damage. |
| `speed` | Speed | Determines turn order priority. |
| `healing_power` | Healing Power | Main scaling stat for healing effects. |
| `crit_rate` | Crit Rate | Chance-oriented offensive modifier for critical hits. |
| `crit_damage` | Crit Damage | Damage multiplier applied when a critical hit occurs. |
| `block_rate` | Block Rate | Defensive modifier affecting block outcomes. |
| `precision` | Precision | Accuracy-oriented stat used to support hit stability and offensive consistency. |
| `evasion_rate` | Evasion Rate | Defensive modifier affecting evade outcomes. |
| `physical_mastery` | Physical Mastery | Supplemental scaling stat for physical-focused builds and effects. |
| `magic_mastery` | Magic Mastery | Supplemental scaling stat for magic-focused builds and effects. |

Related runtime battle fields:

| Field | Name | Meaning |
| --- | --- | --- |
| `status_effects` | Status Effects | Active states currently affecting the entity. |
| `cooldowns` | Cooldowns | Remaining cooldown state for equipped skills. |
| `shield_value` | Shield Value | Current shield amount that absorbs damage before HP. |

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
- no hidden dodge stat
- no random critical hits from basic attacks
- damage variance is fixed at `+/- 5%`

### 7.4 Damage formula

The combat system uses a transparent formula family:

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

### 7.5 Statuses

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

This section serves as a compact overview.

The detailed skill specification has moved to:

- `11-class-skill-system.md`

Core rules:

- each class has its own skill pool
- weapon style influences stat leaning and build profile rather than skill access
- each dungeon loadout may equip up to `4` active skills
- the basic attack is always available and has no cooldown
- skills use cooldowns and do not consume MP
- battle actions are selected automatically from the configured loadout
