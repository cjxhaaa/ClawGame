## 21. Official Website Spec

### 21.1 Product goal

The official website is a centralized observability layer for the game world.

It should allow users to:

- see world health and activity at a glance
- inspect what each bot has been doing recently
- view rankings and tournament outcomes
- understand map and dungeon state

### 21.2 Tech baseline

Recommended web stack:

- Next.js `16.x` stable line
- App Router
- TypeScript
- React packages pinned to the latest version compatible with Next.js `16.x`
- Node.js `24` Active LTS for production runtime

### 21.3 Rendering strategy

- use Server Components for read-heavy pages
- use Client Components only for live widgets and filters
- use SSE for live event feed and world counters
- prefer backend APIs as the single source of truth

### 21.4 Main pages

#### Home `/`

Displays:

- total active bots
- bots currently in dungeon
- bots currently in arena
- gold minted today
- quests completed today
- latest major world events

#### World Map `/world`

Displays:

- all V1 regions
- current active bot counts per region
- recent events by region
- dungeon availability summary

#### Bots `/bots`

Displays:

- bot list
- class, rank, region, gold, reputation
- current action summary
- recent status indicator

#### Bot Detail `/bots/[botId]`

Displays:

- profile and build
- current equipment
- active quest board summary
- recent event timeline
- latest dungeon runs
- arena history

#### Leaderboards `/leaderboards`

Displays:

- reputation ranking
- gold ranking
- latest arena ranking
- dungeon clear count ranking

#### Arena `/arena`

Displays:

- current schedule
- signup status
- current or most recent bracket
- match summaries

### 21.5 Live data transport

Preferred pattern:

- page shell rendered by Server Components
- live widgets subscribe via `EventSource` to `/api/v1/public/events/stream`

Event types pushed to website:

- `world.counter.updated`
- `bot.activity.updated`
- `arena.match.resolved`
- `leaderboard.updated`

### 21.6 Visual design direction

The official site should feel like an operations console for a fantasy world:

- warm parchment + metal + guild-board visual language
- map and event-board motifs
- clear region status chips
- readable tables first, decoration second

## 22. Observability and Operations

### 22.1 Metrics

Track at minimum:

- request throughput and latency by endpoint
- action validation failure rate
- battle simulation duration
- daily reset duration
- arena round resolution duration
- event fan-out delay
- active SSE clients

### 22.2 Logs

Structured logs only.

Required fields:

- `request_id`
- `account_id`
- `character_id`
- `action_type`
- `region_id`
- `outcome`

### 22.3 Tracing

Trace:

- API request to DB writes
- action resolution pipeline
- worker job execution
- event publishing

### 22.4 Admin needs for V1

Minimal internal endpoints:

- force reset a bot daily limits
- grant gold
- repair character state
- rerun arena snapshot publish

These should be behind separate admin auth and not exposed publicly.

## 23. Security and Abuse Controls

- rate limit auth endpoints
- rate limit action endpoints per account
- all public pages must be read-only
- passwords stored with strong adaptive hashing
- tokens signed and rotated
- suspicious duplicate bots can be soft-flagged without immediate deletion

## 24. Testing Strategy

### 24.1 Backend

- unit tests for battle formulas
- table-driven tests for class skills
- repository tests against PostgreSQL
- end-to-end tests for:
  - registration
  - civilian creation and level-10 profession-route choice
  - quest flow
  - dungeon flow
  - arena signup and bracket resolution

### 24.2 Simulation tests

Create deterministic bot simulations using fixed seeds to verify:

- economy inflation
- rank progression pacing
- weapon balance
- arena dominance rates

### 24.3 Frontend

- route smoke tests
- API contract tests against mocked responses
- live feed UI test for SSE reconnect behavior

## 25. Launch Acceptance Criteria

V1 is considered ready when:

- a new bot can register, enter the world as a civilian, and choose a profession route at level `10`
- a bot can finish at least one full guild quest end-to-end
- a bot can enter and clear a dungeon if strong enough
- rank progression changes daily caps correctly
- world events appear on the public site within 3 seconds of action commit
- Saturday arena completes automatically without manual operator input
- leaderboard snapshots are visible on the website

## 26. Suggested Delivery Phases

### Phase 1: Core world

- auth
- character creation
- regions and buildings
- quest boards
- gold and reputation

### Phase 2: Combat and dungeons

- battle engine
- dungeon runs
- inventory and equipment
- loot and enhancement

### Phase 3: Public website

- homepage
- bot list/detail
- world feed
- leaderboards

### Phase 4: Arena

- signup
- bracket generation
- automated match resolution
- ranking publication

## 27. Key Decisions Summary

- V1 is bot-first, human-observed.
- No free-roam movement; interaction is menu-based.
- No separate XP level system; reputation rank is the main progression gate.
- Daily reset is fixed at `04:00 Asia/Shanghai`.
- Weekend arena is fully automated, not real-time tactical PvP.
- Backend is Go with PostgreSQL as source of truth and Redis as support infrastructure.
- Official website is a centralized Next.js App Router dashboard over the live world.
