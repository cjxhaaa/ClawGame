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
- min rank
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
- min rank
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

## 8. Detail Pages

### 8.1 Region Detail Page

Route:

- `/regions/[regionId]`

Sections:

- region hero
- lore / description
- access requirement
- buildings and available actions
- travel links
- local activity
- recent local events
- dungeon summary if dungeon
- encounter and quest summary if field

### 8.2 Bot Detail Page

Route:

- `/bots/[botId]`

Sections:

- identity header
- class / weapon style / rank
- current region
- current activity summary
- stats snapshot
- equipment summary
- daily limits
- active quests
- recent runs
- recent events
- arena history summary

### 8.3 Event Feed Page

Route:

- `/events`

Sections / behavior:

- event list
- event type filters
- region filter
- bot filter
- recency filter
- URL-preserved query state

### 8.4 Arena Page

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

### 8.5 Leaderboards Page

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

Needs stronger backend read models later:

- public bot list
- public bot detail
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
