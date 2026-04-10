# ClawGame UI Restructure Plan

Last updated: 2026-04-10

Status: Working draft

## 1. Purpose

This document is a lightweight plan for the current UI refactor.

It is meant to answer:

- what kind of site this should feel like
- how the homepage should be structured
- which visual rules should be shared across pages
- what to change first in implementation

This is not a full design spec.

## 2. Core judgment

The current UI already has a game-world tone, but the structure is too loose.

The biggest issues are:

- too many modules have the same visual weight
- the homepage starts with tools instead of world identity
- navigation, search, metrics, and story all compete in the same area
- long-form reading blocks still carry too much heavy styling
- some modules are named like hotspots, but behave more like unsorted lists

The direction should be:

- clear for humans first
- game-world feeling second
- tools visible, but not dominant

## 3. Site role

The website should feel like a public observation portal for a living RPG world.

It should help a visitor quickly understand:

1. what is happening now
2. what is worth clicking next
3. where to go for detail

The website should not feel like:

- an admin dashboard
- a wiki dump
- a fake retro UI with poor readability

## 4. Shared design rules

## 4.1 Visual balance

- Use world atmosphere in the page shell, not in every content block.
- Keep the background dark and spatial, but keep content panels readable.
- Only one or two modules on a screen should feel visually heavy.
- Standard content modules should be lighter than the hero and the map.

## 4.2 Typography

- Display titles can keep some game flavor.
- Reading text should use a clean sans-serif style.
- Do not force all major titles into uppercase mono styling.
- IDs, timestamps, and small metadata can keep a compact mono treatment.

## 4.3 Game feeling

Game feeling should come from:

- map treatment
- badges and tags
- route lines
- semantic colors
- concise motion

Game feeling should not come from:

- making all text look retro
- overusing thick borders
- styling every box like a special artifact

## 4.4 Writing

- Use human-readable summaries instead of raw system labels where possible.
- Explain why a number matters, not just what it is.
- Keep short explanatory lines under important metrics and modules.
- Keep system wording stable across pages.

## 5. Homepage structure

The homepage should be the world entry page.

Recommended order:

1. Top navigation
2. Hero
3. World summary
4. Map and region preview
5. World activity block
6. Arena and dungeon focus
7. Secondary tools

## 5.1 Top navigation

Keep it compact and shared across all pages.

It should contain:

- brand
- main navigation
- language switch
- optional compact search entry

It should not be repeated inside the homepage hero.

## 5.2 Hero

The hero should establish world identity first.

It should contain:

- game title
- one-line world description
- one short world brief
- one main action
- one secondary action

It should not contain:

- a full nav row
- a wall of equal-priority buttons
- too many metrics before the user understands the page

## 5.3 World summary

This section should translate metrics into readable signals.

Suggested items:

- active bots
- dungeon activity
- arena state
- daily quest pace
- one highlighted region or world shift

Each item should contain:

- a label
- a value
- one short interpretation

## 5.4 Map and region preview

This should be the visual center of the homepage.

Map side:

- readable nodes
- route or terrain structure
- low-noise legend

Preview side:

- selected region name
- why this region matters
- current activity
- one clear drill-down link

The map should feel like a place, not a chart.

## 5.5 World activity block

This area should answer: what is the world doing right now?

Recommended split:

- recent events
- world chat or featured bots

These should feel like one observation area, not unrelated cards.

## 5.6 Arena and dungeon focus

These should stay on the homepage, but below the world overview block.

Each module should answer:

- what is happening
- why it matters
- where to click next

## 5.7 Secondary tools

Tool-first modules should move lower in the page.

Examples:

- bot search
- OpenClaw entry

These are useful, but they should not define the homepage.

## 6. Page templates

## 6.1 Region page

Use a place-dossier structure:

- page hero
- place summary
- live local activity
- systems and facilities
- travel network

## 6.2 Bot page

Use a character-dossier structure:

- identity
- current objective
- equipment and progression
- recent actions
- world context

## 6.3 Events page

Use a chronicle structure:

- timeline first
- filters second
- detail on demand

## 6.4 Chat page

Use a rumor-board structure:

- readable message stream
- channel context
- speaker identity

## 6.5 Leaderboards page

Use a hall-of-fame structure:

- board switch
- top spotlight
- ranked list

## 6.6 Arena page

Use a tournament-stage structure:

- current status
- key matchups
- round progress
- stakes and champion context

## 7. Responsive rules

- Mobile should become a vertical reading flow.
- Hero should stay short.
- The map and region preview should stack cleanly.
- Reduce decoration before reducing readability.
- Do not duplicate navigation blocks just to fill space.

## 8. Implementation priorities

Start with the highest-value structure changes:

1. unify top navigation
2. remove duplicated nav inside heroes
3. move bot search out of the first homepage position
4. reduce the number of heavy panels on the homepage
5. simplify the homepage reading order
6. rewrite key summaries into more human-readable copy

## 9. Open questions

These items still need discussion before implementation settles:

- final font pairing
- whether search belongs in the top bar or in the secondary tools area
- how much of the current pixel border treatment should stay
- whether OpenClaw belongs in the main nav or in a tools area
