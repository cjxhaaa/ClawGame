# Combat System Framework

## 1. Design Goals

This document defines the full battle foundation that dungeon monsters and encounter scripts will use later.

Goals:

- make combat deterministic enough for bots to plan around
- keep enough variation to make enemy identity matter
- expose all important decisions through structured logs
- support dungeon parties of up to `4` bots from the first playable version
- make combat fully auto-executed but still explainable
- align with the seasonal stat system and dungeon gear system already defined

## 2. Combat Model

Combat in V1 is:

- fully server-authoritative
- fully turn-based
- menu-driven, not positional action combat
- deterministic except for explicitly declared roll points
- logged step by step for debugging, spectator viewing, and bot learning

Every battle is resolved as a sequence of:

1. battle initialization
2. round start
3. entity turn loop
4. round end effects
5. victory or defeat resolution

## 3. Battle Scope

V1 should support these battle scopes:

- `field_skirmish`
  - short encounters used for quest, material, and XP loops
- `elite_encounter`
  - tougher single-wave or double-wave fights
- `dungeon_wave`
  - standard dungeon room combat
- `boss_phase`
  - named boss fight or a phase of one
- `arena_duel`
  - bot versus bot future mode

The battle engine should be the same across all scopes. What changes is:

- participant template
- wave count
- reward package
- AI script
- fail conditions

## 4. Entity Model

Every battle entity should expose:

- `entity_id`
- `team_id`
- `entity_type`
  - `player`
  - `monster`
  - `summon`
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

Recommended stat payload:

- `max_hp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`
- optional future stats:
  - `accuracy`
  - `status_mastery`
  - `status_resistance`

## 5. Battle Start Rules

On battle start:

- all entities are instantiated from templates plus runtime modifiers
- current HP is set to max unless the encounter says otherwise
- pre-battle passives and set effects are applied
- the party roster and the `4` equipped active skills per bot are locked for the run
- a first-round turn order preview can be calculated and stored
- battle metadata is emitted for observer and replay systems

Required battle metadata:

- `battle_id`
- `battle_type`
- `encounter_id`
- `region_id`
- `dungeon_id` if present
- `party_size`
- `wave_index`
- `started_at`

## 6. Round Structure

Each round follows this sequence:

1. round start effects
2. turn order calculation
3. one turn per living entity
4. round end effects
5. win/loss check

Dungeon round-limit rule:

- a dungeon battle may last at most `10` rounds
- if enemies are still alive after round `10`, the battle is a failure
- this timeout failure does not consume a dungeon entry
- other failure modes may still consume the entry depending on mode rules

Round start effects include:

- decrementing temporary locks that expire at round start
- resolving start-of-round regen or delayed triggers if defined by skill
- emitting a round header log entry

Round end effects include:

- poison and burn damage
- regen healing
- duration countdowns
- expiry cleanup

## 7. Turn Order Rules

Turn order is determined by:

1. higher `speed`
2. lower `entity_id`

Rules:

- turn order is recalculated at the start of every round
- effects that change `speed` can shift the next round order
- stun, freeze-like skip, or hard crowd control do not remove an entity from order; they consume its turn

## 8. Action Types

V1 supports these action families:

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

For the first dungeon release, `summon` and `escape` can remain engine-supported but content-unused.

Action selection rule:

- player-side entities are never manually piloted during battle
- each entity acts from its `4` equipped skills plus the always-available basic attack
- the action chosen each turn must be explainable by the auto-battle priority rules

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

V1 targeting principles:

- player skills should avoid ambiguous targeting
- monster AI can use simple scripted priorities
- any random target selection must be explicitly logged

## 10. Threat and Target Priority

To support dungeon monsters cleanly, the combat system should include a lightweight threat model.

Every monster tracks a threat table against player-side entities.

Threat generated:

- direct damage dealt: `1.0x` final damage
- direct healing performed: `0.45x` effective healing, split to visible enemies if needed
- shielding granted: `0.35x` shield amount
- taunt skill: fixed threat injection

Monster default target logic:

1. forced target from taunt or script
2. highest threat living target
3. lowest current HP target if the script prefers execution behavior
4. random living target if no other condition applies

Reason:

- this keeps tank-like roles meaningful later
- it also keeps monster behavior explainable rather than arbitrary

## 11. Hit, Randomness, and Variance

Bot-friendly combat should keep randomness controlled.

Rules:

- basic attacks hit at `100%`
- each skill defines its hit chance if not guaranteed
- no hidden dodge stat in the first combat release
- no random crits on basic attacks
- damage variance is `+/- 5%`
- all roll points must be logged

Allowed roll points:

- hit or miss
- status application success
- random target selection
- damage variance
- loot rolls after battle

## 12. Damage and Healing Formula Family

Physical damage:

`final_physical = max(1, (skill_power + actor.physical_attack * atk_ratio) - target.physical_defense * def_ratio)`

Magic damage:

`final_magic = max(1, (skill_power + actor.magic_attack * atk_ratio) - target.magic_defense * def_ratio)`

Healing:

`final_heal = max(1, skill_power + actor.healing_power * heal_ratio)`

Shield:

`final_shield = max(1, skill_power + actor.healing_power * shield_ratio)`

Damage processing order:

1. raw effect calculated
2. multiplicative bonuses applied
3. shield absorbs damage first
4. remaining damage reduces HP
5. on-hit status checks resolve

## 13. Damage Types

V1 damage types:

- `physical`
- `magic`
- `holy`
- `poison`
- `burn`
- `true`

Defense interaction:

- `physical` uses `physical_defense`
- `magic` and `holy` use `magic_defense`
- `poison` and `burn` ignore a portion of defense only if the skill explicitly says so
- `true` damage ignores both defenses

## 14. Status System

Statuses should be explicit, short-lived, and machine-readable.

V1 primary statuses:

- `poison`
  - end-of-round damage
- `burn`
  - end-of-round damage
- `stun`
  - skip next turn
- `silence`
  - cannot cast magic-tagged skills
- `regen`
  - end-of-round healing
- `shielded`
  - absorbs incoming damage first
- `defense_up`
  - percentage bonus to physical and magic defense
- `defense_down`
  - percentage reduction to physical and magic defense
- `attack_up`
  - percentage bonus to main attack stats
- `attack_down`
  - percentage reduction to main attack stats
- `speed_up`
  - percentage bonus to speed
- `speed_down`
  - percentage reduction to speed
- `vulnerable`
  - target takes increased incoming damage

Status rules:

- every status has:
  - `status_id`
  - `status_type`
  - `source_entity_id`
  - `duration_rounds`
  - `stacks`
  - `potency`
- unless marked stackable, reapplying the same status refreshes duration only
- status resolution timing must be part of the status definition

## 15. Cooldowns and Action Gating

Every skill should define:

- skill tier
  - `normal`
  - `advanced`
  - `ultimate`
- cooldown
- target type
- damage or healing family
- status effects if any

Cooldown tier rules:

- normal skill: `1` round cooldown
- advanced skill: `2` round cooldown
- ultimate skill: `3` round cooldown
- basic attack: `0` round cooldown

Action gating rules:

- if cooldown is active, the action is illegal
- if target requirement is not met, the action is illegal
- if the actor is stunned, only skip-turn resolution happens
- if the actor is silenced, only non-magic actions remain valid
- if the skill is not part of the active `4`-skill loadout, it is illegal for that battle

## 16. Player-side Skill Structure

A skill definition should include:

- `skill_id`
- `name`
- `class`
- `weapon_style` if restricted
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

This shape should be shared by player and monster active skills where possible.

Dungeon loadout rules:

- each character can equip up to `4` active skills before entering a dungeon
- the basic attack does not consume a skill slot
- only skills valid for the currently equipped weapon style may be loaded
- bots should be able to save loadout presets per dungeon or role

Recommended auto-battle priority:

1. use emergency survival skills if needed
2. use healing or shielding if allies are in danger
3. use control or AoE when the battlefield state favors it
4. use an available ultimate or the strongest single-target finisher
5. fall back to the basic attack if all skills are on cooldown

## 17. Monster AI Rules

Monster AI in V1 should be scripted and explainable, not emergent.

Recommended decision ladder:

1. if a phase script says "use forced skill", do that
2. if HP threshold behavior is active, follow that branch
3. if a high-priority skill is off cooldown and condition is met, use it
4. else use the strongest legal default attack

Condition examples:

- target is silenced
- self HP below `50%`
- at least `2` enemies alive
- boss phase changed this round
- highest threat target has shield active

This allows monsters to feel intelligent without making behavior opaque.

## 18. Encounter and Wave Structure

Encounter container should support:

- `encounter_id`
- `battle_type`
- `region_id`
- `monster_pack`
- `waves`
- `victory_rewards`
- `failure_outcome`

Wave rules:

- field encounters usually use `1` wave
- elites can use `1-2` waves
- dungeons can use `1-3` regular waves plus a boss wave
- boss fights can have `2-3` internal phases while still remaining one encounter

## 19. Victory and Defeat Resolution

Victory:

- all hostile entities are defeated
- rewards are granted
- quest and seasonal XP progress is updated
- battle summary is emitted

Defeat:

- all player-side entities are defeated
- encounter marked failed
- dungeon failure logic or retreat logic resolves
- partial reward is only granted if the encounter definition allows it

Retreat:

- not allowed in arena
- allowed in some field encounters
- in dungeons, retreat rules should be defined per mode

## 20. Battle Log Contract

Every battle must emit a structured log sequence.

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

Each log entry should expose:

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

## 21. Observer-facing Summary

For the website, every battle should also produce a compact summary:

- winner
- rounds taken
- total damage dealt
- total healing done
- key statuses applied
- finishing action
- notable drops or rewards

This allows the observer website to show bot activity without dumping raw logs by default.

## 22. Integration With Progression and Equipment

Combat power should resolve from:

`base_template + season_level_growth + equipment_bonus + temporary_modifiers`

Battle rewards should later feed:

- seasonal XP
- adventurer reputation if tied to quests or contracts
- material drops
- equipment drops
- dungeon clear metrics

## 23. Rules for the First Playable Release

To keep the first combat implementation shippable, the initial scope should be:

- `1-4` bots per dungeon party
- one to four monsters per wave
- no summons in shipping content
- no escape in boss battles
- no hidden stats
- no proc chains
- no reactive counterattack system yet

## 24. Open Tuning Decisions

- whether `holy` should remain separate from general magic in later formulas
- whether threat should be exposed to bots numerically or only indirectly
- whether `accuracy`, `status_mastery`, and `status_resistance` should launch with combat or with the gear expansion
- whether elite encounters should always include one script-driven mechanic for observer interest
