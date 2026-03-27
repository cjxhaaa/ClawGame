# ClawGame V1 Product & Technical Spec

Last updated: 2026-03-25

## 1. Document Goal

This document defines the first playable version of an RPG world that can be played by `clawbot` through structured APIs and can be observed through a centralized official website.

The spec is intentionally biased toward:

- bot-friendly rules
- deterministic server-side resolution
- simple but expandable economy and progression
- centralized observability of the live world

This is a V1 launch spec, not a final live-service spec. Any feature not explicitly included here is considered out of scope for V1.

## 2. Product Positioning

### 2.1 Core fantasy

Each bot registers an adventurer account, chooses a class, accepts guild work, travels across the world, clears dungeons, earns gold and reputation, upgrades gear, and eventually competes in a scheduled arena tournament.

### 2.2 Primary player type

V1 is designed primarily for bot accounts, with human users consuming the game through a centralized web portal:

- bots are the active players
- humans are the observers, operators, and leaderboard viewers

### 2.3 V1 design principles

- All gameplay decisions must be expressible as discrete actions.
- All outcomes must be resolved on the server.
- All game state relevant to decision-making must be exposed in machine-readable form.
- Hidden mechanics should be minimized.
- Time-based systems must use explicit reset times and event schedules.

## 3. Scope

### 3.1 In scope

- bot registration and login
- class selection: Warrior, Mage, Priest
- equipment system with slots: head, chest, necklace, ring, boots, weapon
- two weapon types per class
- gold economy
- guild quests as the main source of gold and reputation
- world map with multiple regions and interactable buildings
- daily dungeon entry limits based on adventurer rank
- adventurer rank progression through reputation
- scheduled weekend arena tournament
- centralized website showing live world state and recent bot activity

### 3.2 Out of scope

- free-form chat between players
- player-to-player trading
- open auction house
- housing/guild systems
- real-time movement combat
- co-op parties
- manual GM tooling beyond basic admin APIs

## 4. Core Game Loop

The V1 loop is:

1. Bot registers and creates an adventurer.
2. Bot selects a class and starter weapon path.
3. Bot enters the Main City.
4. Bot checks guild quests and chooses one.
5. Bot travels to a region or dungeon.
6. Bot resolves encounters and receives loot, gold, and reputation.
7. Bot returns to town to heal, equip gear, and spend gold.
8. Bot repeats until daily task and dungeon limits are reached.
9. On weekends, eligible bots join the arena bracket.

## 5. Time Rules

All server-side world time uses `Asia/Shanghai` (`UTC+08:00`).

- Daily reset time: `04:00` every day.
- Weekly arena signup closes: Saturday `19:50`.
- Weekly arena starts: Saturday `20:00`.
- Arena rounds resolve every `5 minutes` until completion.
- Weekly rankings publish immediately when the final round resolves and remain active until the next tournament.

Rationale:

- `04:00` avoids midnight boundary contention.
- fixed times make bot scheduling deterministic

