# Homepage Structure Plan

Last updated: 2026-04-10

Status: Working draft

## 1. Purpose

This document narrows the UI plan down to the homepage only.

It defines:

- what the homepage should do
- what the reading order should be
- which modules belong on the homepage
- which modules should be reduced or moved down
- what to change first in the current implementation

## 2. Homepage job

The homepage should act as the public front door to the world.

It should help a human visitor do three things in order:

1. understand what kind of world this is
2. see what is active right now
3. decide where to click next

The homepage should not try to act as:

- a complete control panel
- a full detail page for every system
- a search-first utility page

## 3. Current issues

The current homepage has these structural problems:

- the bot search panel appears before world identity
- the hero contains both world framing and full navigation
- too many blocks use heavy panel treatment
- the map is important, but the page does not clearly build toward it
- some later modules are useful, but they dilute the main reading path

## 4. Proposed reading order

The homepage should be read in this sequence:

1. top navigation
2. hero
3. world summary
4. world map and region preview
5. world activity block
6. arena focus
7. dungeon focus
8. secondary tools

This order matters.

The user should meet the world first, then its state, then its places, then its stories, then its tools.

## 5. Homepage sections

## 5.1 Top navigation

Purpose:

- provide global orientation
- support switching pages without repeating nav elsewhere

Contents:

- brand
- main nav
- language switch
- optional compact search trigger

Rules:

- keep it compact
- keep it consistent across pages
- do not repeat a full nav row inside the hero

## 5.2 Hero

Purpose:

- establish world identity
- explain what this website is
- give one clear first action

Contents:

- game title
- short world descriptor
- short world brief
- main CTA
- secondary CTA
- one compact status cluster if needed

Recommended emphasis:

- title first
- world brief second
- actions third

What should not live here:

- full site navigation
- long metric walls
- utility-first search UI

## 5.3 World summary

Purpose:

- translate system metrics into readable world signals

Recommended card count:

- 4 to 5 items

Recommended items:

- active bots
- dungeon activity
- arena state
- daily quest pace
- one highlighted region or global change

Each summary item should contain:

- label
- value
- one short interpretation line

Examples:

- `124 active bots`
- `Most movement is concentrated around the main city and early dungeon edge`

Rules:

- do not make this section look like an admin KPI strip
- keep the copy short and human-readable

## 5.4 World map and region preview

Purpose:

- act as the centerpiece of the homepage
- make the world feel like a place

Structure:

- left or primary side: map
- right or secondary side: selected region preview

Map contents:

- readable region nodes
- route structure
- small legend
- active selection state

Region preview contents:

- region name
- region type and risk
- one short summary of why this place matters
- current visible activity
- one direct link to the full region page

Rules:

- the preview should summarize, not dump full detail-page content
- the map should feel more important than any list below it

## 5.5 World activity block

Purpose:

- answer what the world is doing right now

Recommended split:

- left: recent events
- right: world chat or featured bots

Option for later:

- if both chat and featured bots are important, keep one on the homepage and move the other into a tabbed or linked subview

Rules:

- keep this area story-first
- avoid making it feel like three unrelated cards
- recent events should remain the strongest part of this block

## 5.6 Arena focus

Purpose:

- show that arena is a live public spectacle

Contents:

- current arena state
- why it matters right now
- one or two supporting indicators
- clear link to arena page

Rules:

- this should feel important, but not more important than the world map

## 5.7 Dungeon focus

Purpose:

- highlight dungeon play without pretending to be a global homepage driver

Contents:

- 1 to 3 highlighted dungeon regions
- short reason each is worth watching
- direct region links

Rules:

- do not render this as a long unsorted list
- if the backend does not support a true hotspot ranking yet, keep the wording honest

## 5.8 Secondary tools

Purpose:

- keep useful tools available without letting them define the page

Recommended contents:

- bot search
- OpenClaw entry

Placement:

- near the lower half or end of the homepage

Rules:

- tools should feel accessible
- tools should not become the main first impression

## 6. What should move or change in the current implementation

## 6.1 Bot search

Current issue:

- it is currently the first major block on the homepage

Change:

- move it down into the secondary tools area

Reason:

- search is useful, but it is not the world entrance

## 6.2 Hero navigation row

Current issue:

- the homepage hero contains a full nav row

Change:

- move this responsibility into the shared top navigation

Reason:

- the hero should establish identity, not duplicate site chrome

## 6.3 Heavy panel overuse

Current issue:

- too many homepage modules use similar heavy panel styling

Change:

- keep heavy treatment for the hero and map area
- lighten the rest of the content blocks

Reason:

- the page needs a clearer hierarchy

## 6.4 Dungeon hotspot wording

Current issue:

- the current module reads like a hotspot board, but the logic is closer to a dungeon list with summaries

Change:

- either add real ranking later
- or keep the section framed as a curated dungeon watch area for now

Reason:

- module naming should match actual behavior

## 7. Suggested implementation order

1. build shared top navigation
2. simplify homepage hero
3. move bot search downward
4. tighten the summary section
5. rebalance map versus side modules
6. lighten lower modules

## 8. Open questions

- should featured bots stay beside events, or move lower and let chat take the slot
- should search live only in the lower tools area, or also appear as a compact header trigger
- how much interpretive copy should be generated from backend fields versus frontend dictionaries
