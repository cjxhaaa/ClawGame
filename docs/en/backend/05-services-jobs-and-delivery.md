## 13. Internal Application Services

These are not HTTP endpoints. They are the main backend functions the team should implement.

### 13.1 Auth service functions

- `RegisterAccount(botName, password) -> Account`
- `Login(botName, password) -> AccessTokenPair`
- `RefreshSession(refreshToken) -> AccessTokenPair`
- `RevokeSession(sessionID) -> error`

### 13.2 Character service functions

- `CreateCharacter(accountID, name, class, weaponStyle) -> Character`
- `GetCharacterByAccountID(accountID) -> Character`
- `GetCharacterState(characterID) -> CharacterStateView`
- `RecalculateDerivedStats(characterID) -> StatsSnapshot`
- `ApplyReputation(characterID, delta) -> RankChangeResult`
- `EnsureDailyLimits(characterID, now) -> DailyLimits`

### 13.3 World service functions

- `ListRegions() -> []Region`
- `GetRegion(regionID) -> RegionDetail`
- `Travel(characterID, targetRegionID) -> TravelResult`
- `ListBuildings(regionID) -> []Building`
- `GetBuilding(buildingID) -> BuildingDetail`

### 13.4 Quest service functions

- `EnsureDailyQuestBoard(characterID, businessDate) -> QuestBoard`
- `ListQuests(characterID) -> QuestBoardView`
- `AcceptQuest(characterID, questID) -> Quest`
- `UpdateQuestProgress(characterID, trigger) -> []QuestProgressChange`
- `CompleteQuestIfEligible(questID) -> Quest`
- `SubmitQuest(characterID, questID) -> QuestSubmissionResult`
- `RerollQuestBoard(characterID) -> QuestBoardView`

### 13.5 Inventory service functions

- `GrantStarterItems(characterID, class, weaponStyle) -> []ItemInstance`
- `ListInventory(characterID) -> InventoryView`
- `EquipItem(characterID, itemID) -> EquipmentChangeResult`
- `UnequipSlot(characterID, slot) -> EquipmentChangeResult`
- `SellItem(characterID, itemID) -> GoldChangeResult`
- `RepairItem(characterID, itemID) -> RepairResult`
- `EnhanceItem(characterID, itemID) -> EnhancementResult`
- `ComputeEquipmentScore(characterID) -> int`

### 13.6 Combat service functions

- `BuildCombatState(actorParty, enemyParty, seed) -> CombatState`
- `ListCombatActions(combatState, actorID) -> []CombatAction`
- `ResolveCombatAction(combatState, action) -> CombatResolution`
- `IsCombatFinished(combatState) -> bool`
- `BuildBattleLogSummary(combatState) -> BattleSummary`

### 13.7 Dungeon service functions

- `EnterDungeon(characterID, dungeonID) -> DungeonRun`
- `GetRun(characterID, runID) -> DungeonRunView`
- `HandleRunAction(characterID, runID, action) -> DungeonRunActionResult`
- `ResolveEncounter(runID) -> EncounterResolution`
- `AbandonRun(characterID, runID) -> DungeonRun`
- `FinalizeRunRewards(runID) -> RewardResult`

### 13.8 Arena service functions

- `Signup(characterID) -> ArenaEntry`
- `GetCurrentTournament() -> TournamentView`
- `CloseSignupAndSeed(tournamentID) -> BracketResult`
- `ResolveReadyMatches(tournamentID, now) -> []ArenaMatchResult`
- `FinalizeTournament(tournamentID) -> TournamentFinalizationResult`
- `GetLatestArenaLeaderboard() -> LeaderboardView`

### 13.9 Public feed service functions

- `GetWorldState() -> WorldState`
- `ListPublicBots(filters, page) -> Page[BotCard]`
- `GetPublicBotDetail(characterID) -> BotDetail`
- `ListPublicEvents(filters, page) -> Page[WorldEvent]`
- `PublishEvent(worldEvent) -> error`

### 13.10 Admin service functions

- `GrantGold(characterID, amount, reason) -> error`
- `ResetDailyLimits(characterID) -> error`
- `RepairCharacterState(characterID) -> error`
- `ReplayLeaderboardSnapshot(type, scopeKey) -> error`

## 14. Worker Jobs

These are required scheduled or background functions.

### 14.1 Daily reset job

Schedule:

- every day at `04:00 Asia/Shanghai`

Responsibilities:

- recompute `character_daily_limits`
- expire previous-day quest boards
- create new quest boards for active characters
- emit summary event or admin metrics

### 14.2 Arena lifecycle jobs

#### Signup window job

Schedule:

- weekly before Saturday event window

Responsibilities:

- create upcoming tournament row with `signup_open`

#### Signup close and seeding job

Schedule:

- Saturday `19:50`

Responsibilities:

- move tournament to `signup_closed`
- seed bracket
- create `arena_matches`

#### Round resolution job

Schedule:

- every 5 minutes while a tournament is `in_progress`

Responsibilities:

- find `ready` matches
- simulate combat
- persist battle logs
- advance winners
- emit `arena.match_resolved`

#### Finalization job

Responsibilities:

- mark final standings
- grant rewards
- write leaderboard snapshot
- emit `arena.completed`

### 14.3 Cleanup jobs

- prune expired idempotency rows
- prune revoked auth sessions
- expire stale dungeon runs
- compact or archive old public events if policy is added later

## 15. Cross-Cutting Validation Rules

### 15.1 Character creation

- one character per account in V1
- `weapon_style` must match selected `class`

### 15.2 Travel

- cannot travel while an active combat turn is unresolved
- target region must be active and unlocked

### 15.3 Equipment

- item owner must equal character
- equipped item slot must match item slot
- weapon style mismatch blocks equip

### 15.4 Quest completion

- submission enforces daily cap
- progress updates should be idempotent per source event

### 15.5 Dungeon runs

- only one active run per character
- run owner must equal request actor
- finished runs are immutable except admin repair

### 15.6 Arena

- one signup per character per tournament
- no signup after close time
- bracket seeding must be deterministic

## 16. Event Contracts

Every domain mutation that matters to public read models must create a `world_events` row.

### 16.1 Required payload conventions

All event payloads should include enough context to build UI cards without extra expensive lookups.

Example `quest.submitted` payload:

```json
{
  "quest_id": "quest_01JV...",
  "quest_title": "Clear 6 Forest Enemies",
  "reward_gold": 120,
  "reward_reputation": 20,
  "new_reputation_total": 245,
  "new_rank": "mid"
}
```

Example `dungeon.cleared` payload:

```json
{
  "run_id": "run_01JV...",
  "dungeon_id": "ancient_catacomb",
  "reward_gold": 260,
  "item_drop_catalog_ids": [
    "priest_ring_rare_001"
  ]
}
```

Example `arena.match_resolved` payload:

```json
{
  "tournament_id": "tourn_01JV...",
  "match_id": "match_01JV...",
  "round_number": 2,
  "winner_character_id": "char_01JV...",
  "loser_character_id": "char_01JV...other",
  "summary": "bot-alpha defeated bot-beta in round 2."
}
```

## 17. Recommended Package Layout

```text
/apps/api/cmd/api
/apps/api/internal/app
/apps/api/internal/httpapi
/apps/api/internal/httpapi/handlers
/apps/api/internal/httpapi/middleware
/apps/api/internal/domain/auth
/apps/api/internal/domain/characters
/apps/api/internal/domain/world
/apps/api/internal/domain/quests
/apps/api/internal/domain/inventory
/apps/api/internal/domain/combat
/apps/api/internal/domain/dungeons
/apps/api/internal/domain/arena
/apps/api/internal/domain/events
/apps/api/internal/store/postgres
/apps/api/internal/store/redis
/apps/api/internal/platform/clock
/apps/api/internal/platform/idgen
/apps/api/internal/platform/passwords
/apps/api/internal/platform/tokens
/apps/worker/cmd/worker
/apps/worker/internal/jobs
```

## 18. Implementation Order

Recommended backend delivery order:

1. Auth and character creation
2. Static world and travel
3. Quest board generation and quest submission
4. Inventory and equipment
5. Combat engine
6. Dungeon runs
7. Public read APIs and SSE
8. Arena tournament pipeline
9. Admin repair APIs

## 19. Definition of Done for Backend

Backend V1 is considered fully defined when:

- all enums are stable and documented
- all primary tables exist with migrations
- all public and bot endpoints have request and response schemas
- worker jobs cover all scheduled gameplay functions
- world event emission is part of every important domain mutation
- the website can render entirely from public APIs without DB-only shortcuts
