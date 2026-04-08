# Game Spec Modules

This folder contains the modular split of `game-spec-v1.md`.

Files:

- `01-overview.md`
  Product positioning, scope, loop, and time rules
- `02-progression-and-combat.md`
  Classes, ranks, stats, combat, and skill kits
- `03-world-systems.md`
  Equipment, economy, world map, buildings, quests, dungeons, and arena
- `04-bot-platform.md`
  Bot integration, event model, backend architecture, data model, and API quality
- `05-website-ops-and-delivery.md`
  Website spec, observability, security, testing, launch criteria, and delivery phases
- `06-world-map-definition.md`
  Dedicated map definition for geography, routes, region identity, and website map presentation
- `07-location-catalog-and-resource-definition.md`
  Observer-first place catalog covering naming, lore, NPCs, facilities, dungeon links, and material outputs
- `08-seasonal-leveling-and-stat-framework.md`
  Season cadence, level XP curve, level stat growth, and the numeric foundation for the future equipment system
- `09-equipment-dungeon-and-loot-framework.md`
  Equipment slots, shared seasonal loot pools, four parallel dungeon set families, class-agnostic weapon drops, and rating-based dungeon gear rewards using `S/A/B/C/D/E`
- `10-combat-system-framework.md`
  Turn flow, action resolution, targeting, status logic, AI behavior rules, and battle log contracts for future dungeon monsters
- `11-class-skill-system.md`
  Class skill pools, weapon-style stat leaning, four-skill loadout rules, and auto-battle selection logic
- `12-dungeon-monster-and-difficulty-system.md`
  Multi-enemy room compositions, normal/elite/boss monster tiers, boss-only-final-room rule, three-tier difficulty (easy/hard/nightmare) with stat multipliers and mechanic density, and current dungeon coverage
- `13-dungeon-data-tables-and-template-spec.md`
  Data schema for the multi-enemy dungeon system including difficulty profiles, room wave slots, boss placement constraints, and enforced validation rules
- `14-first-batch-dungeon-balance-sheets.md`
  Complete 6-room easy/hard/nightmare composition tables for Ancient Catacomb and Sandworm Den, with full monster stats, boss phase specs, nightmare threshold guidance, and validation checklist
- `15-battle-consumable-and-potion-system.md`
  Removes full HP refill between dungeon rooms and defines the current ungated HP/attack/defense/speed potion economy
- `16-battle-auto-resolution-flow.md`
  Step-by-step runtime flow for automatic battle simulation, from initialization to round loop, action resolution, logging, and terminal settlement
- `17-combat-power-evaluation-and-preview-system.md`
  Defines panel combat power, per-item equipment scoring, total score composition, and strength preview models for dungeons and arena

Notes:

- The full combined version remains at `docs/en/game-spec-v1.md`.
