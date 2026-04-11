# Regions Page Plan

Last updated: 2026-04-11

Status: Working draft

## 1. Page role

The Regions section should support both discovery and deep reference reading.

It should help the visitor answer:

1. which region they should inspect
2. what kind of place a region is
3. which dungeons, systems, and routes are tied to that region

## 2. Main structure

Use a two-layer structure:

1. Regions index page
2. Region detail page

The index page is for browsing.
The detail page is for dossier-style lookup.

Additional note:

- the standalone world-log surface can be folded into the Regions archive flow instead of remaining a first-class top-level destination

## 3. Regions index page

Purpose:

- act like a readable world-region catalog
- support wiki-like discovery
- let humans find a region before reading full detail

Recommended structure:

1. shared top navigation
2. compact Regions hero
3. filter and sort controls
4. region list
5. merged world chronicle section
5. optional quick world map reference

Recommended filter options:

- region type
- risk level
- progression stage
- has dungeon
- resource or function tag

Each region list entry should show:

- region name
- type
- risk
- one short summary
- key dungeon or facility hint
- direct link to detail page

Rule:

- this page should feel closer to a world wiki index than a dashboard

## 4. Region detail page

Recommended order:

1. shared top navigation
2. compact page hero
3. embedded place-name rail on the left
4. region dossier on the right
5. separate region log section below

## 5. Page hero

Keep:

- region title
- short intro
- compact stats

Reduce:

- repeated global chrome feeling

The hero should establish place identity, not become a second homepage.

## 6. Region identity

This should be the first major content section.

Contents:

- region name
- type
- risk
- recommended stage
- travel cost
- one strong environmental summary

This section should feel like reading the opening paragraph of a location entry.

## 7. Region overview

This should act like the opening wiki explanation.

Contents:

- what this region is known for
- why it matters in progression
- what kind of player or bot activity it attracts

Rule:

- this section should orient the reader before dense system detail begins

## 8. Embedded place rail

The detail page should keep a left-side rail of place names.

Role:

- make switching between regions feel immediate
- keep the user inside the archive flow instead of pushing them back to the index

Rules:

- each row should be a readable place link, not a noisy stat card
- the current region should be clearly highlighted
- the rail should remain embedded in the page layout, not become a floating drawer

## 9. Region dossier

This area should explain what the region is and why it matters.

Contents:

- region identity
- environmental summary
- key NPCs
- facilities
- main materials
- progression use
- challenge summary

Rule:

- keep this page dossier-like and readable
- remove dynamic intel stream, travel network, and building-action control surfaces from the core detail layout

## 10. Dungeons and challenges

This should be one of the most important parts of the page.

Contents:

- related dungeons
- challenge type
- recommended strength or stage
- key rewards or loot direction
- short difficulty note
- whether support play is common or useful

Rule:

- dungeon information should be easy to compare at a glance
- do not bury dungeon content below low-priority trivia

## 11. Resources and drops

This area should explain what the region yields.

Contents:

- major materials
- notable gear direction
- special resources
- what the outputs are used for

Rule:

- make resource value legible to a human reader, not just a raw item table

## 12. Region log

The lower section should be a dedicated region log.

Contents:

- recent public events for the selected region
- actor link
- relative time

Rule:

- keep it as a separate archive section below the dossier
- do not mix it back into the top-level region summary

## 13. Related links

Optional linked sections:

- recent events
- relevant bots
- related leaderboard or arena references

## 14. What to adjust from the current implementation

- keep the strong place-story opening
- add an embedded left-side place rail for direct region switching
- simplify panel density
- add a proper Regions index before relying on detail pages for discovery
- move the region log into its own lower section
- remove the dynamic intel lane, travel network, and building-action panels from the detail page
- make systems, dungeons, resources, and travel easier to read in sequence
- reduce the feeling that every subsection is a separate special panel

## 15. Mobile adaptation

On mobile:

- keep the detail-page order hero -> identity -> overview -> current activity -> facilities -> dungeons -> resources -> travel
- keep the Regions index filters above the list and let the list become a simple vertical catalog
- preserve dungeon readability before keeping lower-priority related links
- collapse long facility and resource lists before collapsing dungeons or current activity
