# Events Page Plan

Last updated: 2026-04-10

Status: Working draft

## 1. Page role

The events page should read like a world chronicle.

It should answer:

1. what has happened recently
2. which kinds of actions are dominating the world
3. which regions are currently hot

## 2. Main structure

Recommended order:

1. shared top navigation
2. compact page hero
3. event timeline header
4. main event feed
5. side context rail

## 3. Event timeline header

Contents:

- active filter
- region scope if present
- short explanation of the feed behavior

Rule:

- filters matter, but the timeline is the page hero in practice

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
- strengthen the feeling of a single world timeline
- reduce extra framing around each event block
