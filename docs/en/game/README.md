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
  Equipment slots, rarity rules, affix counts, dungeon-tier gear tables, set effects, and rating-based dungeon gear rewards using `S/A/B/C/D/E`
- `10-combat-system-framework.md`
  Turn flow, action resolution, targeting, status logic, AI behavior rules, and battle log contracts for future dungeon monsters
- `11-class-skill-system.md`
  Class and weapon-style skill pools, four-skill loadout rules, and auto-battle selection logic
- `12-dungeon-monster-and-difficulty-system.md`
  Monster templates, up to `6` escalating rooms, boss phases, rating rules, and the split of rating-based gear rewards versus kill-based material drops
- `13-dungeon-data-tables-and-template-spec.md`
  Concrete table-oriented specs for dungeon definitions, rooms, monster templates, boss scripts, rating reward tables, and monster material drop tables
- `14-first-batch-dungeon-balance-sheets.md`
  First-pass monster stats, skills, AI, wave plans, and loot weights for Ancient Catacomb and Thorned Hollow

Notes:

- The full combined version remains at `docs/en/game-spec-v1.md`.
