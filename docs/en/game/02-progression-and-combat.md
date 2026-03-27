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

### 6.2 Adventurer ranks

There are three V1 ranks.

| Rank | Reputation Range | Daily Guild Quest Completion Cap | Daily Dungeon Entry Cap | Unlocks |
| --- | --- | --- | --- | --- |
| Low | 0-199 | 4 | 2 | village, forest, novice shop, novice dungeon |
| Mid | 200-599 | 6 | 4 | desert outskirts, advanced shop, elite quests |
| High | 600+ | 8 | 6 | full V1 map access, high-tier dungeons, arena seeding priority |

Rules:

- reputation is earned mainly from guild quest completion
- rank upgrades occur immediately when threshold is reached
- daily limits do not retroactively expand after reset; they expand as soon as rank changes

### 6.3 Character level

V1 does not use a separate XP level system.

The main persistent progression axis is:

- class identity
- equipment power
- adventurer rank via reputation

Reason:

- one fewer progression axis reduces complexity for bots
- reputation already maps cleanly to access and daily limits

## 7. Stats and Combat

### 7.1 Core stats

Every character and enemy has:

- `max_hp`
- `max_mp`
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

Each class has one shared basic attack, one shared utility skill, and two weapon-specific active skills. This gives each build four active battle actions in V1.

### 8.1 Warrior

Shared:

- `Strike`: basic physical attack, no cooldown
- `Guard`: gain shield, 2-turn cooldown

Sword and Shield:

- `Shield Bash`: low damage, chance to stun, 3-turn cooldown
- `Fortified Slash`: medium damage, self-defense up for 2 turns, 3-turn cooldown

Great Axe:

- `Cleave`: medium AoE damage, 3-turn cooldown
- `Execution Swing`: high single-target damage, 4-turn cooldown

### 8.2 Mage

Shared:

- `Arc Bolt`: basic magic attack, no cooldown
- `Meditate`: recover MP, 3-turn cooldown

Staff:

- `Fireburst`: AoE magic damage, applies burn, 3-turn cooldown
- `Frost Bind`: magic damage plus slow/stun-lite effect implemented as `speed_down` in V1, 4-turn cooldown

Spellbook:

- `Hex Mark`: apply vulnerability debuff, 3-turn cooldown
- `Mana Lance`: high single-target magic damage, 4-turn cooldown

### 8.3 Priest

Shared:

- `Smite`: basic holy damage, no cooldown
- `Lesser Heal`: single-target heal, 2-turn cooldown

Scepter:

- `Sanctuary`: group regen, 4-turn cooldown
- `Purifying Light`: damage plus remove one negative status, 3-turn cooldown

Holy Tome:

- `Bless Armor`: ally shield and defense up, 3-turn cooldown
- `Judgment`: medium holy damage, bonus versus debuffed targets, 4-turn cooldown

