# Combat System Document

## 1. Document Goal

This document defines the complete combat specification used by design, implementation, testing, structured logging, and observer-facing presentation.

It describes official current rules only, without phase labels, version-stage wording, or open discussion placeholders.

## 2. Combat Overview

Combat is a server-authoritative turn-based system. Player entities and monster entities are resolved under the same execution framework.

Core characteristics:

- server-authoritative resolution
- turn-based execution
- automatic action selection
- controlled randomness with explicit roll points
- full structured battle logs

Standard execution chain:

1. battle initialization
2. round start
3. in-round action loop
4. round end
5. victory or defeat resolution

## 3. Combat Scopes and Types

The unified combat engine supports these combat types:

- `field_skirmish`: short field combat
- `elite_encounter`: elite combat
- `dungeon_wave`: dungeon room combat
- `boss_phase`: boss phase combat
- `arena_duel`: arena combat

The same engine is reused for all types. Differences are configuration-driven:

- participant templates
- wave structure
- AI scripts
- failure conditions
- reward packages

## 4. Battle Entity Model

Each battle entity must include:

- `entity_id`
- `team_id`
- `entity_type` (`player` / `monster` / `summon`)
- `template_id`
- `name`
- `level`
- `rank`
- `stats`
- `current_hp`
- `shield_value`
- `status_effects`
- `cooldowns`
- `equipped_skills`
- `alive`

Unified `stats` payload:

- `max_hp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`

## 5. Battle Initialization Rules

Initialization must perform:

- instantiation of all entities from templates plus runtime modifiers
- setting initial HP (full by default unless overridden by encounter config)
- application of pre-battle passives and gear effects
- locking battle loadouts (up to 4 active skills per character)
- optional first-round order preview emission
- battle metadata emission

Battle metadata fields:

- `battle_id`
- `battle_type`
- `encounter_id`
- `region_id`
- `dungeon_id` (if present)
- `party_size`
- `wave_index`
- `started_at`

## 6. Round Flow

Each round executes in fixed order:

1. round-start effects
2. turn-order calculation
3. per-entity action loop for living entities
4. round-end effects
5. win/loss check

Round-start effects:

- resolve statuses and delayed effects that trigger at round start
- clear temporary locks that expire at round start
- emit round header log event

Round-end effects:

- periodic damage (for example `poison`, `burn`)
- periodic healing (for example `regen`)
- status duration countdown
- expired status cleanup

Dungeon battle round cap:

- one dungeon battle is capped at `10` rounds
- if enemies are still alive at cap, the battle is a defeat

## 7. Turn Order Rules

Turn order priority:

1. higher `speed`
2. lower `entity_id`

Additional rules:

- order is recalculated at the start of each round
- speed changes affect the next round order
- skip-turn crowd control does not remove an entity from order; it consumes that entity turn

## 8. Action System

Action families:

- `basic_attack`
- `skill_attack`
- `heal`
- `shield`
- `cleanse`
- `buff`
- `debuff`
- `summon`
- `wait`
- `escape`

Auto-action rules:

- both player and monster sides execute automatically
- player-side selection is from `4` equipped active skills plus basic attack
- illegal actions (cooldown, invalid target, unmet condition) cannot be executed

## 9. Targeting Rules

Target types:

- `self`
- `single_ally`
- `single_enemy`
- `all_allies`
- `all_enemies`
- `front_enemy`
- `lowest_hp_ally`
- `highest_threat_enemy`
- `random_enemy`

Principles:

- player skill targeting must be explicit
- monster targeting must follow scripted priority
- random target decisions must be logged

## 10. Threat and Target Priority

Monsters maintain a threat table against player-side entities.

Threat generation:

- dealt damage: `1.0x` final damage
- effective healing: `0.45x` effective heal
- granted shield: `0.35x` shield amount
- taunt: fixed threat injection

Default target chain:

1. forced target (script or taunt)
2. highest-threat living target
3. lowest-current-HP target for execution-style scripts
4. random living target as fallback

## 11. Hit and Randomness Rules

Randomness must remain controlled and replayable.

Rules:

- basic attack hit chance is `100%`
- skill hit chance is defined per skill
- damage variance is `+/- 5%`
- all roll points must be emitted in logs

Allowed roll points:

- hit or miss
- status apply success or failure
- random target selection
- damage variance
- post-battle loot roll

## 12. Formula and Resolution Order

Physical damage:

`final_physical = max(1, (skill_power + actor.physical_attack * atk_ratio) - target.physical_defense * def_ratio)`

Magic damage:

`final_magic = max(1, (skill_power + actor.magic_attack * atk_ratio) - target.magic_defense * def_ratio)`

Healing:

`final_heal = max(1, skill_power + actor.healing_power * heal_ratio)`

Shield:

`final_shield = max(1, skill_power + actor.healing_power * shield_ratio)`

Resolution order:

1. calculate raw effect value
2. apply multipliers
3. absorb damage with shield first
4. reduce HP by remaining damage
5. resolve follow-up statuses

## 13. Damage Types

Damage types:

- `physical`
- `magic`
- `holy`
- `poison`
- `burn`
- `true`

Defense interaction:

- `physical` uses `physical_defense`
- `magic` and `holy` use `magic_defense`
- `poison` and `burn` defense bypass is skill-defined
- `true` ignores both defenses

## 14. Status System

Status payload fields:

- `status_id`
- `status_type`
- `source_entity_id`
- `duration_rounds`
- `stacks`
- `potency`

Core statuses:

- `poison`
- `burn`
- `stun`
- `silence`
- `regen`
- `shielded`
- `defense_up`
- `defense_down`
- `attack_up`
- `attack_down`
- `speed_up`
- `speed_down`
- `vulnerable`

Status rules:

- non-stackable statuses refresh duration on reapply
- each status must declare its resolution timing

## 15. Cooldowns and Action Legality

Skill tiers:

- `normal`
- `advanced`
- `ultimate`

Cooldown rules:

- `normal`: `1` round
- `advanced`: `2` rounds
- `ultimate`: `3` rounds
- basic attack: `0` rounds

Action legality:

- active cooldown: illegal
- no valid target: illegal
- under `stun`: turn is skipped
- under `silence`: magic-tagged skills are illegal
- non-equipped skill: illegal

## 16. Skill Schema and Auto Priority

Unified skill schema fields:

- `skill_id`
- `name`
- `class`
- `weapon_style` (if restricted)
- `action_type`
- `target_type`
- `damage_type`
- `cooldown_rounds`
- `hit_chance`
- `skill_power`
- `atk_ratio`
- `def_ratio`
- `heal_ratio`
- `shield_ratio`
- `status_payloads`
- `ai_tags`

Loadout rules:

- each character equips up to `4` active skills before battle
- basic attack does not consume a slot
- equipped skills must match current weapon style restrictions

Recommended auto priority:

1. emergency survival
2. team sustain and shielding
3. control or AoE
4. finisher or high-multiplier damage
5. basic attack fallback

## 17. Monster AI Rules

Monster AI uses an explainable scripted chain:

1. forced-skill instruction
2. phase-threshold branch
3. high-priority skill condition check
4. default attack fallback

Common condition dimensions:

- self HP threshold
- enemy alive count threshold
- phase switch marker
- target status (for example shield, silence)

## 18. Wave and Encounter Structure

Encounter fields:

- `encounter_id`
- `battle_type`
- `region_id`
- `monster_pack`
- `waves`
- `victory_rewards`
- `failure_outcome`

Wave constraints:

- field encounters are usually `1` wave
- elite encounters are usually `1-2` waves
- dungeons usually use regular waves plus a boss wave
- a boss may have multiple internal phases inside one encounter

## 19. Win and Loss Resolution

Victory:

- all hostile entities are defeated
- rewards are granted
- quest and progression updates are applied
- battle summary is emitted

Defeat:

- player-side wipe or configured failure condition
- configured failure outcomes are applied

Retreat:

- retreat availability is controlled by combat-type configuration

## 20. Battle Log Contract

Required event families:

- `battle.started`
- `round.started`
- `turn.started`
- `action.selected`
- `action.resolved`
- `status.applied`
- `status.expired`
- `damage.resolved`
- `healing.resolved`
- `shield.resolved`
- `entity.defeated`
- `wave.cleared`
- `battle.finished`

Required fields per log entry:

- `battle_id`
- `round`
- `turn_index`
- `event_type`
- `actor_entity_id`
- `target_entity_ids`
- `skill_id`
- `damage_type`
- `rolled_values`
- `final_values`
- `status_changes`
- `timestamp`

## 21. Observer Summary Contract

Each battle also emits a compact summary:

- winner
- total rounds
- total damage
- total healing
- key statuses
- finishing action
- key rewards

## 22. Integration With Progression

Combat power aggregation:

`base_template + season_level_growth + equipment_bonus + temporary_modifiers`

Combat result outputs:

- experience
- reputation
- materials
- equipment
- dungeon metrics

## 23. Integration With Potion System

Battle sustain and potion behavior is defined by:

- `15-battle-consumable-and-potion-system.md`

Execution flow and runtime sequencing are defined by:

- `16-battle-auto-resolution-flow.md`
