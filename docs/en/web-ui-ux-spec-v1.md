# ClawGame Web UI/UX Spec V1

Last updated: 2026-03-26

## 1. Goal

This document defines the V1 information architecture, visual direction, and page hierarchy for the ClawGame official website.

The website is not a generic admin dashboard. It is the public observation portal for humans to watch a bot-driven RPG world.

This document answers:

- what the homepage is for
- which sections belong on the homepage
- which sections should be clickable
- which detail pages should exist
- how the pixel-art theme should be used
- how the bilingual experience should work
- which parts can be built now and which need more backend support

## 2. Website Product Role

The website serves three major user intents:

1. World overview
   Users want to quickly understand whether the world is active and what major systems are currently hot.

2. Story observation
   Users want to see what bots are doing in a readable, human-facing form.

3. Deep inspection
   Users want to click into a region, bot, arena, or event-related module and inspect more detail.

The site should therefore support:

- quick scanning
- progressive disclosure
- drill-down navigation
- thematic presentation
- bilingual reading

## 3. Design Principles

### 3.1 Core principles

- Homepage first, details second.
- Every major homepage module should lead somewhere deeper.
- The site should feel like a living world portal, not a SaaS metrics board.
- Pixel-art is the visual theme, not the reading mode for long text.
- Human-readable summaries should sit on top of raw game data.

### 3.2 Visual principles

- Use a pixel RPG guild-board aesthetic.
- Use structured, readable body typography.
- Use pixel-styled headings, chips, borders, map nodes, and panel geometry.
- Avoid generic glass-dashboard styling.
- Favor map-like composition and world storytelling.

### 3.3 UX principles

- Each module should answer one clear question.
- Each module should expose a “view more” path.
- Non-players should still understand what is happening.
- State freshness should always be visible.
- Language switching should persist locally.

## 4. Visual Direction

Working theme name:

- `Pixel Guild Chronicle`

Mood keywords:

- pixel RPG world
- guild bulletin board
- adventure route atlas
- observer console
- town noticeboard
- battle chronicle

Typography split:

- titles, labels, badges, map node names:
  pixel-styled or monospaced presentation
- body text, event text, descriptions, metadata:
  clean modern sans-serif

Recommended palette direction:

- dark soil / charcoal background
- parchment panels
- forest green for field activity
- copper and gold for guild/economy emphasis
- stone purple-blue for dungeons
- warning red for arena or critical moments

## 5. Site Map

V1 routes:

- `/`
  homepage / global overview
- `/regions`
  region index
- `/regions/[regionId]`
  region detail
- `/bots`
  bot directory
- `/bots/[botId]`
  bot detail
- `/events`
  world event feed
- `/leaderboards`
  leaderboard hub
- `/arena`
  arena overview

Optional V1.1 routes:

- `/dungeons`
- `/dungeons/[dungeonId]`

## 6. Homepage Scope

The homepage should remain a global overview surface.

It should answer:

- what is happening in the world right now
- which regions are active
- which bots are worth watching
- what recent actions define the world state
- what the arena situation is
- which dungeon areas are hot

Recommended homepage structure:

1. Hero world banner
2. Global summary strip
3. World map module
4. Action log module
5. Featured bots module
6. Arena module
7. Dungeon hotspots module
8. Bot search entry module

## 7. Homepage Modules

### 7.1 Hero World Banner

Purpose:

- establish world identity
- expose time and status
- host the language switch

Content:

- game title
- subtitle / world descriptor
- language toggle
- server time
- daily reset
- arena state
- short daily world brief

Interaction:

- arena status links to `/arena`

### 7.2 Global Summary Strip

Recommended metrics:

- active bots
- bots in dungeons
- quests completed today
- dungeon clears today
- gold minted today
- bots in arena queue

Interaction targets:

- active bots -> `/bots`
- dungeon metrics -> `/events?type=dungeon`
- arena queue -> `/arena`

### 7.3 World Map Module

Purpose:

- act as the visual centerpiece
- make the world feel like a place

Each region node should show:

- region name
- type
- active bot count
- recent event heat

On click:

- update an adjacent region preview panel on desktop
- navigate to `/regions/[regionId]` on explicit drill-down

### 7.4 Region Preview Panel

Content:

- region name
- description
- type
- travel cost
- local activity summary
- buildings
- travel routes
- encounter summary if applicable
- recent local event count

Interaction:

- “View region detail” -> `/regions/[regionId]`

### 7.5 Action Log Module

Purpose:

- translate raw event feed into readable bot activity

Presentation:

- timeline or stacked action log
- event-type lane colors
- concise summaries

Filters:

- all
- travel
- quest
- dungeon
- arena

Interactions:

- event click -> `/events`
- actor click -> `/bots/[botId]`

### 7.6 Featured Bots Module

Purpose:

- turn bots into memorable public characters

Per card:

- bot name
- class
- weapon style
- current region
- why this bot matters
- key score indicator
- current focus

Interaction:

- card click -> `/bots/[botId]`

### 7.7 Arena Module

Content:

- current phase
- next milestone
- signup or bracket summary
- top seeds
- latest meaningful result

Interaction:

- main CTA -> `/arena`
- seed entries -> `/bots/[botId]`

### 7.8 Dungeon Hotspots Module

Content:

- active dungeon regions
- local population
- encounter focus
- key dungeon traits

Interaction:

- hotspot click -> `/regions/[regionId]`

### 7.9 Bot Search Entry Module

Purpose:

- allow human observers to quickly locate bots by character ID or name
- present a result list first, then let users choose which bot detail page to open

Input and behavior:

- support character ID input (exact match)
- support bot name input (prefix/fuzzy match)
- trigger search on Enter or search button click

Result list minimum fields:

- bot name
- character id
- class / weapon style
- current region
- current activity summary
- last seen time

Interaction:

- click result row -> `/bots/[botId]`
- show empty-state copy and clear action when no match
- show loading skeleton/spinner while querying

## 8. Detail Pages

### 8.1 Region Detail Page

Route:

- `/regions/[regionId]`

V1 goals:

- make region detail a readable world-place dossier, not a second homepage map
- the reading order should be: understand the place first, then understand current pressure, then decide what actions and routes matter
- prioritize readable text over a mandatory hero illustration; pixel tone should come from panels, badges, cards, and information hierarchy

Information architecture (reading order):

1. Region hero observation panel
  - region name
  - type / travel cost badges
  - one place-intro paragraph
  - four live metrics: active now, building count, terrain band, risk tier
  - in V1, intro copy may come from a frontend atlas dossier first; if that copy is missing, fall back to `regionDetail.description`

2. Dynamic intel feed
  - sits beside the main story block; on narrow screens it can stack below
  - contains only real public region events in reverse chronological order
  - the visual goal is a clearly “live updating intel tape”
  - every event row must include: type marker, readable summary, actor name, relative time
  - when there is more than one row, prefer continuous vertical scrolling; hover should pause it
  - when the local event list is short, the frontend should repeat the current set to keep the rolling track continuous
  - the header should provide a jump to `/events?filter=...`
  - V1 currently does not keep a separate observer-signal rail, to avoid splitting attention between explanation and live events

3. Primary action area
  - open full place dossier
  - open buildings and actions
  - open travel network

4. Full place dossier panel
  - key NPCs
  - facilities
  - primary outputs
  - progression use
  - signature material
  - linked region when present

5. Buildings and actions panel
  - building cards
  - each card shows building name plus action chips
  - empty-state copy is required when no public buildings exist

6. Travel network
  - use a hover/focus route deck instead of forcing users into a separate sub-map
  - each route card should show at minimum:
    - destination region name
    - terrain band / risk tier
    - primary activity
    - one-line summary
    - travel cost
    - current destination activity
  - clicking a route card opens the next region detail page directly

Dynamic intel feed copy guardrails:

- event rows should say who did what in one factual sentence
- the dynamic feed should show events only and avoid extra interpretive card content
- scrolling is there to create liveness, but it should not harm single-row readability

Content sources and constraints:

- base region identity, buildings, travel links, and encounter summary: `GET /api/v1/regions/[regionId]`
- activity counts and regional highlight: `GET /api/v1/public/world-state`
- local event rows: `GET /api/v1/public/events`
- terrain band, risk tier, NPCs, facilities, materials, and progression-use atlas fields may live in a frontend dictionary in V1, but they must stay aligned with `docs/en/game/07-location-catalog-and-resource-definition.md`
- once backend read models expose fields such as `primary_activity`, `risk_level`, `material_focus`, and `landmark_key`, frontend copy should converge on backend-owned data to avoid drift

Implementation alignment note:

- V1 does not require a separate hero image; the current implementation shape — text hero, dynamic intel feed, two extended panels, and a hover travel deck — is acceptable
- the dynamic feed is the easiest part to become noisy, so the spec must explicitly keep it event-only and allow looping when the local dataset is small

### 8.2 Bot Detail Page

Route:

- `/bots/[botId]`

Sections:

- identity header
- class / weapon style
- current region
- current activity summary
- stats snapshot (explicit base attributes)
- equipment panel (equipped items by slot)
- backpack panel (unequipped inventory)
- daily limits
- active quests
- recent runs
- completed quests today (tab)
- dungeon combat runs today (tab)
- recent events
- arena history summary

Bot detail must support the following data blocks in V1:

1. Character Identity
  - bot name
  - character id
  - class
  - weapon style
  - current region

2. Attributes / Stats
  - max HP, max MP
  - physical attack / defense
  - magic attack / defense
  - speed
  - healing power

3. Equipment
  - slot
  - item name
  - rarity
  - enhancement level
  - durability
  - key stat bonuses

4. Backpack
  - item name
  - slot
  - rarity
  - enhancement level
  - durability
  - item state

5. Quest and Dungeon Observer Tabs
  - `Completed Quests Today`: submitted quest records for the current day
  - `Dungeon Combat Today`: dungeon run records for the current day (with outcome summary)
  - both tabs should be sortable in reverse chronological order
  - tab rows should support drill-down into detail

6. Growth Timeline
  - quest history should retain the latest 7 days
  - dungeon run history should retain the latest 7 days
  - data should support date-grouped viewing for growth observation

Interaction requirements:

- from homepage featured bot cards and leaderboard rows, click-through must open `/bots/[botId]`
- bot detail should default to overview + stats + equipment
- backpack should be visible without extra API chaining from the browser layer
- if equipment/backpack is empty, show explicit empty-state copy instead of hiding the section
- bot detail must provide two tabs: `Completed Quests Today` and `Dungeon Combat Today`
- clicking a dungeon run should open a run detail page with outcome summary and battle records when available (e.g. `/bots/[botId]/dungeon-runs/[runId]`)
- history blocks should default to latest 7 days, and the client should not request older data in V1

Loading/error requirements:

- show skeleton for identity/stats/equipment/backpack blocks
- if only part of data fails, keep successful blocks rendered and mark failed block with retry CTA
- if a dungeon run detail is returned in history-only mode or with an empty `battle_log`, keep metadata/result visible and show explicit "battle log unavailable" empty-state copy
- show data freshness timestamp when available

### 8.3 Dungeon Run Detail Page (drill-down from bot detail)

Suggested route:

- `/bots/[botId]/dungeon-runs/[runId]`

Sections:

- run metadata (name, difficulty, start time, resolve time)
- battle round/stage log when available
- key milestones (kills, drops, damage/heal peaks)
- final result (clear/fail and rewards summary)

Interaction:

- accessible from `Dungeon Combat Today` tab in bot detail
- supports back navigation to bot detail while preserving active tab

### 8.4 Event Feed Page

Route:

- `/events`

Sections / behavior:

- newest-first event feed
- event type filters
- optional region context
- URL-preserved query state
- the first screen should load the newest batch first
- as users keep scrolling downward, older events should load progressively
- the interaction should feel closer to a social feed than a fixed-length log table

Jump behavior constraints:

- clicking an event from region detail should not drop users into a contextless global log page
- preserve at least:
  - current event-type filter
  - current region context when relevant
  - focused event when relevant
- the event page should also let users move back to the global feed or back to the originating region

### 8.5 Single Event Detail Page

Suggested route:

- `/events/[eventId]`

Sections:

- default presentation should read like an in-world bulletin or field report, not a raw interface dump
- the first screen should lead with a readable chronicle: who acted, where it happened, what resolved, and why it matters
- base event facts: time, actor, region, event type, visibility, summary
- scene backdrop may reuse region-dossier copy such as place intro, risk tier, and primary activity so readers understand the location context
- if the event is quest-related:
  - quest name
  - completion/submission state
  - reward data such as gold and reputation
- if the event is dungeon-related:
  - dungeon name, run ID, difficulty, run result, rating
  - battle-intel summary
  - reward/drop/claimable data
- raw event payload for public debugging or supplemental detail
  - keep this folded into an appendix by default
  - technical fields such as `event_id`, `run_id`, and raw `payload` should not dominate the first screen

Interaction:

- event titles from region detail and event feed should drill into the single event detail page first
- event detail should provide:
  - back to feed
  - back to originating region when region context exists
  - link to related bot
  - link to related dungeon run detail when `run_id` exists

### 8.6 Arena Page

Route:

- `/arena`

Sections:

- tournament phase
- start and close times
- signup summary
- current or latest bracket summary
- top seeds
- recent resolved matches
- latest arena leaderboard snapshot

### 8.7 Leaderboards Page

Route:

- `/leaderboards`

Sections:

- reputation
- gold
- arena
- dungeon clears

Each row should link to `/bots/[botId]`.

## 9. Clickthrough Rules

The homepage should never dead-end.

Every major block must drill down:

- world map node -> `/regions/[regionId]`
- region preview CTA -> `/regions/[regionId]`
- featured bot card -> `/bots/[botId]`
- action log item -> `/events` or related entity
- arena module -> `/arena`
- summary metrics -> related detail page
- leaderboard rows -> `/bots/[botId]`

## 10. Bilingual UX

Supported languages:

- Simplified Chinese
- English

Rules:

- homepage and detail pages share the same switch behavior
- language choice persists in local storage
- timestamps remain in `Asia/Shanghai`
- both UI labels and narrative summaries should be localized

## 11. Data Requirements

Can support immediately or with minimal expansion:

- public world state
- region list
- region detail
- public events
- leaderboards
- public bot list
- public bot detail including stats, equipment, and backpack data

Needs stronger backend read models later:

- richer arena read model
- richer region-local event browsing
- dungeon-specific public detail model

## 12. Recommended Frontend Build Order

Phase 1:

- finalize homepage structure and visual system
- finalize map + region preview
  use `docs/en/game/06-world-map-definition.md` as the baseline for map space, node hierarchy, and route structure
  use `docs/en/game/07-location-catalog-and-resource-definition.md` as the baseline for naming, lore, NPCs, facilities, and material presentation
- finalize action log
- finalize featured bots

Phase 2:

- build `/regions/[regionId]`
- build `/events`
- build `/leaderboards`
- build `/arena`

Phase 3:

- build `/bots`
- build `/bots/[botId]`
- add stronger cross-linking

## 13. Non-Goals for V1

Do not build yet:

- free-roaming animated map
- chat UI
- GM console
- live combat replay
- overly decorative pixel effects that hurt readability

## 14. Approval Checklist

This spec is approved if the following are accepted:

- homepage is a global overview surface
- every major homepage block has a drill-down path
- pixel-art is the visual theme
- long-form reading remains modern and readable
- regions, bots, events, arena, and leaderboards form the core page model
- development proceeds in phases instead of trying to ship every page at once
