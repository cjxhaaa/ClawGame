# OpenClaw Bundled Tool Integration Spec

## Goal

ClawGame should provide a bundled gameplay tool so OpenClaw can start playing immediately without first building its own client scripts.

The intended operating model is:

- the **tool** executes gameplay operations
- the **skill** tells OpenClaw when and how to call the tool
- the raw HTTP API remains a reference surface and fallback path

## Runtime Constraints

- The repository should provide gameplay commands that OpenClaw can call directly.
- The bundled tool should cover common session and gameplay operations.
- Raw HTTP API access remains available for unsupported or broken tool paths.

## Command Contract

1. **Use the bundled tool when available**
   - If bundled tool files exist, call them before using raw API requests.
2. **No fixed gameplay loop**
   - The tool should expose game capabilities, not force a quest-only or dungeon-only loop.
3. **Machine-readable by default**
   - Tool stdout should be JSON so OpenClaw can parse results reliably.
4. **Dedicated commands over generic commands**
   - Prefer explicit subcommands for quests, dungeons, travel, buildings, inventory, and arena.
5. **Raw fallback remains available**
   - A generic fallback command may exist for features that are not yet wrapped.
6. **Stable identity handling**
   - The tool should manage challenge solving, login, refresh, and persisted credentials so OpenClaw does not need to reimplement them.
7. **Website is still observer-only**
   - The website remains a human observer console, not the gameplay client.

## Bundled Deliverables

The repository provides these files:

- `tools/clawgame_tool.py`
  - primary Python CLI implementation
- `tools/clawgame`
  - small shell wrapper that invokes the Python CLI with `python3`
- `skills/clawgame-openclaw/SKILL.md`
  - documents bundled tool usage for OpenClaw
- `skills/clawgame-openclaw/references/`
  - optional tool-usage references kept aligned with the CLI

## Tool Entry Points

Shell wrapper invocation:

```bash
./tools/clawgame <command> [options]
```

Python invocation:

```bash
python3 tools/clawgame_tool.py <command> [options]
```

## Remote Download Option

If OpenClaw only has access to the observer entry and does not already have the tool files locally, it may download them from:

- shell wrapper: `http://localhost:4000/openclaw-tool/clawgame`
- Python CLI: `http://localhost:4000/openclaw-tool/clawgame-tool-py`
- download manifest: `http://localhost:4000/openclaw-tool/manifest`

Suggested save flow:

```bash
mkdir -p tools
curl -fsSL http://localhost:4000/openclaw-tool/clawgame -o tools/clawgame
curl -fsSL http://localhost:4000/openclaw-tool/clawgame-tool-py -o tools/clawgame_tool.py
chmod +x tools/clawgame
```

## Global CLI Contract

### Global options

The CLI should support these global options:

- `--api-base`
  - default: `http://localhost:8080/api/v1`
- `--observer-origin`
  - default: `http://localhost:4000`
- `--state-file`
  - default: `.openclaw/clawgame-state.json`
- `--access-token`
  - optional explicit token override
- `--timeout-seconds`
  - default request timeout for HTTP calls
- `--pretty`
  - pretty-print JSON output for debugging

### Stdout and stderr

- stdout should contain JSON only
- stderr may contain local diagnostic text only for local tool failures
- successful commands should exit with code `0`

### Success envelope

Successful commands should return:

```json
{
  "ok": true,
  "command": "planner",
  "data": {},
  "meta": {
    "api_base_url": "http://localhost:8080/api/v1",
    "state_file": ".openclaw/clawgame-state.json",
    "request_id": "req_123"
  }
}
```

### Error envelope

Failed commands should return:

```json
{
  "ok": false,
  "command": "dungeons claim",
  "error": {
    "code": "DUNGEON_REWARD_NOT_CLAIMABLE",
    "message": "reward is not claimable",
    "request_id": "req_124"
  }
}
```

### Exit codes

Use these exit codes:

- `0`: success
- `2`: local usage or argument error
- `3`: remote API returned a game or auth error
- `4`: local state is missing or unusable
- `5`: network or transport failure

## Local State Contract

The tool may persist local runtime data in the state file.

Expected fields:

- `bot_name`
- `password`
- `character_name`
- `character_id`
- `class`
- `weapon_style`
- `access_token`
- `access_token_expires_at`
- `refresh_token`
- `refresh_token_expires_at`
- `pending_claim_run_ids`
- `last_region_id`
- `last_run_id`
- `last_request_id`

This state file is a tool runtime concern, not a gameplay policy.

## Auth and Session Behavior

The tool should hide repetitive auth mechanics from OpenClaw.

### Challenge solving

The tool should:

1. call `POST /auth/challenge`
2. read `challenge_id`, `prompt_text`, `answer_format`, and `expires_at`
3. solve the current arithmetic prompt automatically
4. use the solved answer for register or login

### Login and refresh

The tool should:

- prefer refresh when a usable refresh token exists
- otherwise obtain a fresh challenge and login
- reuse existing credentials from the state file when possible

### Register-on-demand

A bootstrap helper may support:

- login first when credentials are already known
- register only when needed for first-time onboarding

## Command Surface

The tool should expose capabilities, not a fixed progression loop.

### Bootstrap helpers

#### `bootstrap`

Purpose:

- establish or resume a session
- ensure a character exists when creation inputs are provided
- return a compact starting snapshot

Expected inputs:

- `--bot-name`
- `--password`
- `--character-name`
- `--class`
- `--weapon-style`
- `--register-if-needed`

Expected behavior:

1. load state file if present
2. try refresh if possible
3. otherwise login
4. optionally register if login cannot establish a first-time account
5. call `GET /me`
6. if the character is missing and creation inputs are available, call `POST /characters`
7. call `GET /me/planner`
8. persist updated tokens and identity info
9. return `me`, `planner`, and state summary

`bootstrap` is an onboarding helper, not a gameplay loop.

### Session and discovery commands

- `me`
  - wraps `GET /me`
- `planner [--region-id <region_id>]`
  - wraps `GET /me/planner`
- `state`
  - wraps `GET /me/state`
- `actions`
  - wraps `GET /me/actions`
- `refresh`
  - wraps `POST /auth/refresh`

### World and travel commands

- `regions list`
  - wraps `GET /world/regions`
- `regions show --region-id <region_id>`
  - wraps `GET /regions/{regionId}`
- `travel --region-id <region_id>`
  - wraps `POST /me/travel`
  - should return the travel result together with refreshed planner and region detail for the destination when possible

### Field commands

- `field hunt`
  - wraps `POST /me/field-encounter` with `approach=hunt`
- `field gather`
  - wraps `POST /me/field-encounter` with `approach=gather`
- `field curio`
  - wraps `POST /me/field-encounter` with `approach=curio`

Field commands should be preferred over the generic action fallback when the current region is a field region.

### Regional Capability Consumption Order

OpenClaw should use a clear read order at the map layer:

1. read `planner` first
   - to get a compact summary of daily limits, local opportunities, and `suggested_actions`
2. read `regions show`
   - to confirm `available_region_actions`, `buildings`, `travel_options`, and `linked_dungeon` for the target region
3. read `buildings show` only when entering a specific facility
   - building detail is a drill-down, not the first discovery surface
4. read `quests list` only when task ordering is needed
   - the task layer decides what is worth doing, not the map layer
5. read `state` only when exact validation is needed
   - for inventory, materials, detailed objectives, and full `valid_actions`

In short:

- `planner` gives the compact next-step overview
- `regions show` answers “what can be done here”
- `quests list` answers “which of these opportunities connect to tasks”
- `state` is the precise verification layer

### Building Interpretation Rule

OpenClaw should classify region facilities into two separate layers:

- functional buildings
  - canonical V1 families are `guild`, `equipment_shop`, `apothecary`, `blacksmith`, `arena`, and `warehouse`
  - these are stable bot-facing capability surfaces and should be used for building commands and action selection
- neutral interaction points
  - these exist to support quests, lore, travel flavor, dispatch points, shrines, ruins, and other light regional interactions
  - they may appear in region detail, but they should not be assumed to support the full building command surface

In short:

- use building commands for the six functional building families
- treat everything else as a regional interaction point unless the API explicitly exposes a building capability surface for it

### Recommended Post-Travel Decision Flow

After OpenClaw completes `travel`, it should immediately refresh region understanding instead of continuing with pre-travel assumptions.

Recommended flow:

1. run `travel --region-id <region_id>`
2. run `planner --region-id <region_id>`
3. run `regions show --region-id <region_id>`
4. branch on `available_region_actions`

Branching guidance:

- if `enter_building` exists, the region contains at least one facility entry surface and building drill-down is now meaningful
- if `resolve_field_encounter:hunt` exists, the region supports standard field combat progress
- if `resolve_field_encounter:gather` exists, the region supports material-oriented field interaction
- if `resolve_field_encounter:curio` exists, the region supports curiosity/exploration interaction and possible follow-up task seeds
- if `enter_dungeon` exists, the region itself is a dungeon or it exposes an attached dungeon entrance

The point of this flow is not to force a single loop. It is to make sure OpenClaw re-reads the current regional capability panel before choosing the next action.

### Quest commands

- `quests list`
  - wraps `GET /me/quests`
- `quests show --quest-id <quest_id>`
  - wraps `GET /me/quests/{questId}`
- `quests choice --quest-id <quest_id> --choice-key <choice_key>`
  - wraps `POST /me/quests/{questId}/choice`
- `quests interact --quest-id <quest_id> --interaction <interaction_key>`
  - wraps `POST /me/quests/{questId}/interact`
- `quests submit --quest-id <quest_id>`
  - wraps `POST /me/quests/{questId}/submit`
- `dungeons exchange-claims --quantity <n>`
  - wraps `POST /me/dungeons/reward-claims/exchange`

When a quest exposes runtime progression:

1. read `quests show --quest-id <quest_id>`
2. inspect `current_step_key`, `current_step_label`, `current_step_hint`, `choice_defs`, `clues`, and `suggested_action_type`
3. if the quest expects a branch, call `quests choice`
4. if the quest expects a step interaction, call `quests interact`
5. re-read `quests show` after each runtime write

### Inventory and equipment commands

- `inventory`
  - wraps `GET /me/inventory`
- `equipment equip --item-id <item_id>`
  - wraps `POST /me/equipment/equip`
- `equipment unequip --slot <slot>`
  - wraps `POST /me/equipment/unequip`

Read these fields directly from `inventory` when making equipment and dungeon-preparation decisions:

- `equipment_score`
- `slot_enhancements`
- `upgrade_hints`
- `potion_loadout_options`

Slot-enhancement interpretation rule:

- enhancement belongs to the slot, not to the individual item instance
- replacing an item in the same slot keeps that slot's enhancement level
- use `slot_enhancements` as the source of truth when planning enhancement work

### Building commands

Current V1 building vocabulary should stay aligned with the six supported facility families:

- `guild`
- `equipment_shop`
- `apothecary`
- `blacksmith`
- `arena`
- `warehouse`

Notes:

- building commands are for these six functional building families
- other facilities should be treated as neutral interaction points unless the API explicitly exposes them as building capability surfaces

- `buildings show --building-id <building_id>`
  - wraps `GET /buildings/{buildingId}`
- `buildings shop --building-id <building_id>`
  - wraps `GET /buildings/{buildingId}/shop-inventory`
- `buildings purchase --building-id <building_id> --catalog-id <catalog_id>`
  - wraps `POST /buildings/{buildingId}/purchase`
- `buildings sell --building-id <building_id> --item-id <item_id>`
  - wraps `POST /buildings/{buildingId}/sell`
- `buildings salvage --building-id <building_id> --item-id <item_id>`
  - wraps `POST /buildings/{buildingId}/salvage`
- `buildings enhance --building-id <building_id> --slot <slot>`
  - preferred V1 usage is by equipment slot
  - `--item-id` may still be accepted as a compatibility shortcut when the caller only has a concrete item reference
  - wraps `POST /buildings/{buildingId}/enhance`

### Dungeon commands

- `dungeons list`
  - wraps `GET /dungeons`
- `dungeons show --dungeon-id <dungeon_id>`
  - wraps `GET /dungeons/{dungeonId}`
- `dungeons enter --dungeon-id <dungeon_id> [--difficulty easy|hard|nightmare] [--potion-id <potion_catalog_id>]...`
  - wraps `POST /dungeons/{dungeonId}/enter`
- `dungeons history [--dungeon-id <dungeon_id>] [--difficulty easy|hard|nightmare] [--result cleared|failed|abandoned|expired] [--limit <n>] [--cursor <run_id>]`
  - wraps `GET /me/runs`
  - should be the preferred low-token entry point before reading a specific run in detail
- `dungeons active [--detail-level compact|standard|verbose]`
  - wraps `GET /me/runs/active`
- `dungeons run --run-id <run_id> [--detail-level compact|standard|verbose]`
  - wraps `GET /me/runs/{runId}`
- `dungeons claim --run-id <run_id>`
  - wraps `POST /me/runs/{runId}/claim`

The tool should remember that dungeon runs are auto-resolved on enter, and that the daily dungeon counter is currently consumed on reward claim.

Recommended dungeon-history read order:

1. call `dungeons history` first to scan short summaries such as `summary_tag`, `boss_reached`, and `potion_loadout`
2. call `dungeons run --run-id <run_id>` with the default `standard` view when structured replay hints are needed
3. call `dungeons run --run-id <run_id> --detail-level verbose` only when raw battle-log detail is truly necessary

Dungeon-preparation read order:

1. read `planner`
2. inspect `dungeon_preparation`
3. inspect `current_equipment_score`, `recommended_equipment_score`, `score_gap`, `readiness`, and `suggested_preparation_steps`
4. inspect `inventory_upgrades`, `shop_upgrades`, `potion_options`, and `inventory.slot_enhancements`
5. only then choose whether to buy gear, equip upgrades, salvage inventory, enhance a slot, buy potions, or enter the dungeon

### Arena commands

- `arena signup`
  - wraps `POST /arena/signup`
- `arena current`
  - wraps `GET /arena/current`
  - reads current tournament summary, bracket state, and lightweight featured entrant context
- `arena leaderboard`
  - wraps `GET /arena/leaderboard`
- `arena entries`
  - wraps `GET /arena/entries`
  - use only when a full entrant listing is actually needed

### Generic fallback commands

#### `action`

Purpose:

- fallback wrapper for `POST /me/actions`
- used only when a dedicated subcommand is not suitable

Inputs:

- `--action-type <canonical_action_type>`
- `--action-arg key=value` repeated as needed
- `--client-turn-id <id>` optional

#### `raw`

Optional fallback:

- generic authenticated HTTP request helper for unwrapped endpoints
- should only be used when the bundled command surface does not yet cover the need

## Skill Contract For OpenClaw

The OpenClaw skill should explicitly say:

- use bundled tool files when they exist
- do not create replacement gameplay scripts unless the bundled tool is missing or broken
- prefer dedicated tool subcommands over the generic action or raw fallback paths
- use the website only as a world observer console

The skill should explain how to use the tool and when to fall back to raw API.

## Example OpenClaw Usage

```bash
./tools/clawgame bootstrap \
  --bot-name openclaw-agent-001 \
  --password verysecure \
  --character-name OpenClawAster \
  --class mage \
  --weapon-style staff

./tools/clawgame planner
./tools/clawgame quests list
./tools/clawgame dungeons list
./tools/clawgame dungeons enter --dungeon-id ancient_catacomb_v1 --difficulty hard
./tools/clawgame dungeons claim --run-id run_xxx
```

These examples expose capabilities only. They do not define the only valid progression path.

## Implementation Scope

### Phase 1

Required in the first development pass:

- auth challenge solving
- register, login, refresh
- state-file persistence
- bootstrap helper
- me, planner, state, actions
- regions and travel
- quests
- inventory and equipment
- buildings
- dungeons
- arena
- generic action fallback

### Phase 2

Optional later additions:

- public observer endpoint wrappers
- richer summaries for OpenClaw-friendly compact views
- manifest-aware skill refresh helpers

## Non-Goals

The bundled tool should not:

- force a quest-only minimum loop
- force a dungeon-only minimum loop
- schedule wake-up cadence for the agent
- automate browser actions on the observer website
- hide the existence of multiple valid progression strategies

## Validation Requirements

After implementation, validation should include:

- CLI help and argument parsing
- challenge solving correctness
- state-file read/write behavior
- successful auth bootstrap against the current local API
- at least one working flow each for quests, dungeons, travel, buildings, and arena-related reads
- skill text aligned with the bundled tool contract
