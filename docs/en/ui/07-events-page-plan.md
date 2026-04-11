# Events Page Plan

Last updated: 2026-04-11

Status: Reduced scope

Current direction:

- the dedicated world-log page is no longer the preferred primary browsing surface
- global event reading should be merged into the Regions archive flow
- `/events` may remain as a compatibility entry or redirect, while event detail pages can still exist when directly linked

## 1. Page role

The events page should read like a world chronicle.

It should answer:

1. what has happened recently
2. which kinds of actions are dominating the world
3. which regions are currently hot

Priority note:

- this page matters, but it sits below Regions and Arena in the main browsing order

## 2. Main structure

Recommended order:

1. shared top navigation
2. compact page hero
3. event timeline header
4. main event feed
5. side context rail

## 3. Event timeline header

Contents:

- category controls
- active filter
- region scope if present
- pagination or page-size control
- short explanation of the feed behavior

Rule:

- filters matter, but the timeline is the page hero in practice
- event volume is high enough that classification and pagination are first-class needs

## 4. Main event feed

This is the core of the page.

Each row should show:

- event type
- summary
- actor
- region
- relative time

Rule:

- rows should feel like a world timeline, not like social cards

## 5. Side context rail

Possible contents:

- regional heat
- arena state
- active event mix

Rule:

- side context should help interpret the timeline, not distract from it

## 6. What to adjust from the current implementation

- keep infinite loading and time-order logic
- keep region scope support
- strengthen category clarity and paging behavior
- strengthen the feeling of a single world timeline
- reduce extra framing around each event block

## 7. Mobile adaptation

On mobile:

- keep the order hero -> timeline header -> event feed -> context summaries
- merge side context into compact blocks below the feed
- keep event rows continuous before preserving optional summaries such as regional heat
- filters should stack above the feed without visually overtaking the timeline
