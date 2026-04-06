# Quest YAML Catalog

This directory is the current source of truth for backend quest templates.

## Layout

```text
configs/
  /daily
  /supplemental
```

- `daily/` holds templates that can appear on the default daily board.
- `supplemental/` holds templates injected by runtime systems such as curio follow-ups.

## File Naming

Use the pattern:

```text
NNN-template-name.yaml
```

Examples:

- `010-kill-region-enemies.yaml`
- `050-investigate-anomaly.yaml`

`NNN` controls human-readable ordering in the catalog.

## Required Top-Level Fields

- `pool`
- `order`
- `template_type`
- `difficulty`
- `flow_kind`
- `rarity`
- `title`
- `description`
- `progress_target`
- `runtime`

## Runtime Shape

`runtime` should define:

- `initial_step_key`
- `completion_step_key`
- `steps`

Optional runtime fields:

- `inspect_step_key`
- `choice_step_key`
- `progress_trigger_type`
- `progress_source`
- `requires_choice`
- `requires_inspection`
- `requires_route_confirm`
- `base_clues`
- `interaction_specs`
- `choice_specs`

## Step Rules

Every step key referenced by:

- `initial_step_key`
- `completion_step_key`
- `inspect_step_key`
- `choice_step_key`
- `interaction_specs[].step_key`
- `interaction_specs[].next_step_key`
- `choice_specs[].next_step_key`

must also exist in `runtime.steps`.

## Authoring Guidance

- Prefer adding a new YAML file over modifying Go code.
- Only change Go logic when the new quest introduces a genuinely new mechanic.
- Keep step labels and hints explicit so OpenClaw can consume them directly.
- Keep `choice_specs` and `interaction_specs` self-contained; they should describe both state changes and clue additions.
