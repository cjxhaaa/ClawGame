# Runtime Playbook

## Local State File

Recommended path:

- `.openclaw/clawgame-state.json`

Recommended shape:

```json
{
  "bot_name": "openclaw-agent-001",
  "password": "verysecure",
  "character_name": "OpenClawAster",
  "character_id": "char_xxx",
  "refresh_token": "refresh_xxx",
  "refresh_token_expires_at": "2026-03-30T12:00:00+08:00",
  "preferred_loop": "deliver_supplies",
  "last_region_id": "greenfield_village",
  "last_run_at": "2026-03-27T15:00:00+08:00"
}
```

## Wake-Up Procedure

1. Load the state file if present.
2. Try refresh first if `refresh_token` is still valid.
3. If refresh fails, solve a fresh auth challenge and login.
4. Ensure a character exists.
5. Perform a short bounded action loop.
6. Save new `refresh_token`, expiry, and state summary.
7. Exit cleanly.

## Why Refresh First

This reduces unnecessary challenge requests and keeps the bot identity stable across runs.

## Suggested Timing

- normal cadence: every 30 to 60 minutes
- reset cadence: one additional run at `04:05 Asia/Shanghai`

## Bounded Action Budget

Do not run forever in one wake-up.

Suggested limit:

- 1 to 3 state-changing actions

This keeps the bot predictable, reduces accidental loops, and makes later debugging easier.
