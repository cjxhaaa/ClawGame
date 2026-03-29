# Battle Auto-Resolution Flow

## 1. Scope

This document explains the execution flow of the automatic combat system step by step.

It is the process-level companion to:

- `10-combat-system-framework.md` (rules and contracts)
- `15-battle-consumable-and-potion-system.md` (sustain and potion rules)

## 2. Runtime Inputs

A battle starts with these inputs:

- battle metadata (`battle_type`, `encounter_id`, `region_id`, optional `dungeon_id`)
- player party snapshot (stats, equipped skills, statuses, inventory-affecting flags)
- enemy wave definition (monster templates, AI tags, wave scripts)
- system config (round cap, action legality rules, random seed strategy)

## 3. Initialization Stage

Initialization order:

1. Create battle context and `battle_id`
2. Instantiate entities from templates and runtime modifiers
3. Set `current_hp`, `shield_value`, cooldowns, and initial statuses
4. Lock skill loadouts
5. Build wave state and active participants
6. Emit `battle.started`

Required output artifacts:

- normalized in-memory battle state
- initial turn-order preview (optional)
- first log packet

## 4. Main Loop

The battle loop runs until a terminal condition is met.

Terminal conditions:

- player-side team wins
- enemy-side team wins
- round cap reached with enemies alive (defeat)
- configured hard fail condition

High-level loop:

1. round start
2. turn order calculation
3. entity turn loop
4. round end
5. terminal check

## 5. Round Start Phase

Round start actions:

1. Increment round counter
2. Resolve start-of-round status effects
3. Remove or downgrade statuses that expire at round start
4. Emit `round.started`

If all enemies or all players are already defeated after start effects, skip directly to terminal check.

## 6. Turn Order Calculation

Order rule:

- sort by `speed` descending, then by `entity_id` ascending

Notes:

- order recalculates each round
- speed buffs/debuffs affect the next recalculation
- entities that cannot act remain in order and consume turn as skip

## 7. Entity Turn Pipeline

For each living entity in current order:

1. Emit `turn.started`
2. Evaluate hard action blockers (`stun`, dead, forced skip)
3. Build legal action set
4. Select action via auto policy (player policy or monster AI script)
5. Emit `action.selected`
6. Resolve action
7. Emit one or more resolution events

### 7.1 Legal Action Set Construction

Action candidates:

- basic attack
- equipped active skills
- potion action if allowed by policy and availability

Filter out illegal actions:

- active cooldown
- no valid target
- silence-incompatible skill
- run-level restrictions (for example potion cap reached)

### 7.2 Auto Action Selection

Player-side selection priority:

1. emergency survival
2. team sustain/shield
3. control or AoE value
4. high-value finisher
5. basic attack fallback

Monster-side selection priority:

1. forced script action
2. phase-threshold branch
3. high-priority conditional skill
4. default attack fallback

## 8. Action Resolution Pipeline

Each selected action resolves with this sequence:

1. Resolve targets
2. Roll hit/miss if required
3. Compute raw values (damage/heal/shield)
4. Apply multipliers and type-based interactions
5. Apply shield then HP changes
6. Apply status payloads
7. Update cooldowns/resources
8. Emit resolution logs (`damage.resolved`, `healing.resolved`, `shield.resolved`, `status.applied`, etc.)

If a target reaches zero HP:

- mark `alive = false`
- emit `entity.defeated`

## 9. Round End Phase

Round end actions:

1. Resolve end-of-round periodic effects
2. Decrement status durations
3. Remove expired statuses and emit `status.expired`
4. Emit round-end summary block (if enabled)

Then run terminal check.

## 10. Terminal Resolution

### 10.1 Victory

If all hostile entities are defeated:

1. Mark battle result as win
2. Resolve rewards and progression updates
3. Emit `battle.finished`
4. Produce observer summary payload

### 10.2 Defeat

If all player-side entities are defeated, or round cap defeat triggers:

1. Mark battle result as loss
2. Resolve failure outcome configuration
3. Emit `battle.finished`
4. Produce observer summary payload

## 11. Dungeon Room Chaining

For `dungeon_wave` runs with multiple rooms:

1. Persist party combat state between rooms
2. Carry over HP/shields/statuses according to dungeon rules
3. Load next room wave and re-enter initialization

HP persistence and potion behavior must follow:

- `15-battle-consumable-and-potion-system.md`

## 12. Logging Checklist

Minimum event sequence for one normal turn:

1. `turn.started`
2. `action.selected`
3. one or more of:
   - `damage.resolved`
   - `healing.resolved`
   - `shield.resolved`
   - `status.applied`
4. optional `entity.defeated`

Battle-level required events:

- `battle.started`
- `round.started`
- `wave.cleared` (when applicable)
- `battle.finished`

## 13. Determinism and Replayability

To keep bot behavior auditable:

- all roll points must be logged
- all auto-action decisions must be explainable from visible state
- replay with same seed and same inputs should reproduce identical outcomes

## 14. Implementation Guardrails

- keep rules in data/config where possible; avoid hardcoding encounter-specific branches in core engine
- keep action legality checks centralized
- keep log schema stable and additive
- separate balance tuning from flow control code
