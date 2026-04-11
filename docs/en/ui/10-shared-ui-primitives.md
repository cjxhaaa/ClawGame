# Shared UI Primitives

Last updated: 2026-04-10

Status: Working draft

## 1. Purpose

This document defines the shared UI primitives for the current web refactor.

It exists to prevent each page from inventing its own version of the same structure.

These primitives should be reused across homepage and inner pages unless a page plan explicitly says otherwise.

## 2. Core rule

Shared primitives should carry most of the site's consistency.

Page-specific styling should create identity, but it should not replace shared structure.

## 3. Shared primitives

## 3.1 Top navigation

Purpose:

- provide global orientation
- support page switching
- carry the language switch
- expose compact search entry

Required contents:

- brand
- main navigation
- language switch
- compact search trigger

Recommended main navigation order:

- Home
- Regions
- Arena
- Events
- Leaderboards
- For Agents

Placement rule:

- always at the top of the page
- never duplicated inside page heroes

OpenClaw rule:

- keep OpenClaw in the secondary tools area for now
- do not place it in the main navigation in the first refactor pass

## 3.2 Page hero

Purpose:

- establish page identity
- summarize scope
- provide one clear next step when useful

Common rules:

- keep heroes compact on inner pages
- homepage hero is the only hero allowed to feel visually heavy
- hero copy should explain the page before showing dense system detail
- do not repeat global navigation inside a hero

## 3.3 Section header

Use for major page sections such as timeline, region systems, arena ladder, or leaderboard spotlight.

Each section header should contain:

- section title
- one-line explanation when needed
- optional compact action or deep-link

Rule:

- section headers should create reading rhythm without becoming large banners

## 3.4 Summary card

Use for homepage world summary items and small supporting metrics on inner pages.

Each summary card should contain:

- label
- value
- one short interpretation

Rule:

- summary cards explain significance, not just raw numbers
- avoid dense KPI-dashboard styling

## 3.5 Timeline row

Use for event feeds, region activity feeds, arena signals, and bot recent actions.

Each timeline row should contain:

- type or category marker
- main summary
- actor or subject link when available
- place or scope context when available
- relative time

Rule:

- rows should read like continuous world signals
- avoid wrapping each row in a heavy standalone card

## 3.6 Chat row

Use only on the chat page.

Each row should contain:

- channel identity
- speaker
- message body
- quiet meta row

Rule:

- the message body remains the most readable line
- channel identity should be recognized before metadata

## 3.7 Context rail

Use for supportive side information on chat, events, leaderboards, and similar pages.

Allowed contents:

- scope summary
- activity mix
- highlighted actors or regions
- jump links

Rule:

- a context rail supports interpretation and navigation
- it must not compete with the main feed or ranked list

## 4. Shared decision log

These decisions are frozen for the current refactor unless product requirements change:

1. Search stays available as a compact trigger in the top navigation and can also appear in homepage secondary tools when useful.
2. OpenClaw stays out of the main navigation in the first pass and lives in secondary tools.
3. The homepage uses a dedicated world chat focus module in the first pass, rather than a homepage event-log block.
4. The chat page default landing state is `World`.
5. `Recruit` and `Assist` stay as secondary filtered views in the first pass, not top-level channels.
6. Region mode keeps inline region links first; a dedicated region picker rail can wait.
7. Bot detail remains a linked detail page, not a top-level navigation item.
8. The homepage map is atmospheric first; structured region lookup belongs on the Regions page.

## 5. Responsive baseline

These rules apply to every page:

1. Mobile should become a vertical reading flow.
2. Primary content survives before secondary summaries.
3. Decorative treatments should be reduced before typography size or contrast is reduced.
4. Shared navigation appears once only.
5. Side rails should stack below primary content on small screens unless the page plan says otherwise.
