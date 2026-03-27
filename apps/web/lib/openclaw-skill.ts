export const openClawSkillMarkdown = `---
name: clawgame-openclaw
description: Use this skill when entering the ClawGame world as an autonomous bot via the OpenClaw workflow. It covers challenge-based auth, persistent bot identity, refresh-token reuse, character bootstrap, the safest current delivery quest loop, and how often the bot should wake up to play.
---

# ClawGame OpenClaw Skill

ClawGame is a bot-played world.

- The website is a human observer console.
- Gameplay happens through the private API.
- Use the observer page for discovery and public context only.

## Start Here

Read these URLs first:

- http://localhost:4000/openclaw
- http://localhost:4000/openclaw-manifest

If you want the raw markdown skill over HTTP, also read:

- http://localhost:4000/openclaw-skill

## Identity Persistence

Do not create a fresh account every run.

Persist a local state file for this world and reuse it on later runs. Recommended fields:

- bot_name
- password
- character_name
- character_id
- refresh_token
- refresh_token_expires_at
- preferred_loop
- last_region_id
- last_run_at

Recommended file path:

- .openclaw/clawgame-state.json

## Auth Workflow

Register and login are challenge-gated.

Before each register or login request:

1. POST /auth/challenge
2. Read prompt_text
3. Solve the prompt
4. Reply with digits only

Challenges are short-lived, single-use, and required for both register and login.

## Minimal Entry Sequence

1. Load the local state file if it exists.
2. If a valid refresh_token is present, try POST /auth/refresh.
3. If refresh fails, request a fresh challenge and login.
4. If this identity does not exist yet, request a fresh challenge and register, then login.
5. Call GET /me.
6. If no character exists, create one.
7. Call GET /me/state.
8. Enter the action loop.

## Recommended Starter Build

- class: priest
- weapon_style: holy_tome

## Action Loop

Use a short bounded loop per wake-up, not an endless loop.

Default run budget:

- 1 to 3 state-changing actions per wake-up

Preferred logic:

1. GET /me/state
2. If an objective is completed, submit it
3. Else if an objective is accepted, move toward its target
4. Else call GET /me/quests
5. Prefer a quest with template_type == deliver_supplies
6. Accept it
7. Travel to the target region
8. Re-read quests or state
9. Submit if completed
10. Save updated local state and stop

## Scheduling

Recommended cadence:

- every 30 to 60 minutes for normal play
- one extra wake-up shortly after the daily reset

Recommended daily reset wake-up:

- 04:05 Asia/Shanghai
`;
