# PVE E2E

This app contains end-to-end gameplay tests that drive the API through real HTTP requests.

Current coverage:

- single-account first-day civilian loop
  - register and log in
  - create a civilian character
  - auto-generate a four-slot first-day board that stays in civilian-safe contract types
  - complete and submit all four contracts
  - buy starter combat prep from the real shops
  - clear the novice equipment dungeon and claim rewards
  - equip the best directly usable loot upgrades and verify combat power growth
- six-account world boss raid
  - create six combat-ready characters
  - join the world boss queue
  - resolve a six-player raid on the sixth join
  - verify each account receives gold and `reforge_stone`
  - verify queue status and raid detail are readable for every member
- six-account first-day boss progression
  - run the full first-day civilian solo loop on six separate accounts
  - clear the novice equipment dungeon before group play
  - enter the world boss with those six first-day accounts
  - verify boss rewards arrive and each account can afford the `800` gold profession change afterward

The helper flows are intentionally modular so later scenarios can stitch together first-day, dungeon, and group-play stages.

Run:

```bash
go test ./apps/e2e -v
```
