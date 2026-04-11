# Chat Page Wireframe Notes

Last updated: 2026-04-11

Status: Working draft

## 1. Purpose

This document turns the chat channel plan into a page-level wireframe description.

It focuses on:

- page zones
- layout hierarchy
- message row structure
- which current elements should stay
- which current elements should move or be removed

## 2. Page layout

Recommended desktop structure:

1. shared top navigation
2. compact chat hero
3. chat layout grid

Recommended chat layout grid:

- top tab rail: World plus region tabs
- center: main feed
- right rail: optional context panel

Recommended mobile structure:

1. shared top navigation
2. compact chat hero
3. horizontal channel tabs
4. main feed
5. collapsible context area

## 3. Desktop zones

## 3.1 Compact hero

Role:

- set page identity
- summarize current scope

Contents:

- title
- one short intro line
- 2 to 3 compact stats

Keep:

- page title
- freshness
- current channel scope

Reduce:

- oversized hero height
- too much emphasis compared with the feed itself

## 3.2 Top channel rail

Role:

- primary navigation inside the chat page

Recommended order:

1. World
2. one tab for each observable region

Each tab should show:

- channel name
- active state

Important rule:

- this top tab row replaces the feeling currently carried by the old channel switch and region link cluster

## 3.3 Main feed column

Role:

- central reading area

Structure inside the feed column:

1. current channel header
2. secondary filters
3. feed list

### Current channel header

Contains:

- active channel title
- one-line explanation
- optional scope note

### Secondary filters

Contains:

- message type filter for Recruit and Assist views
- region scope control when needed

Rules:

- keep these visually lighter than the top tab rail
- do not let them look like the main navigation

### Feed list

Contains:

- continuous chat rows
- fixed-height channel window
- bottom-anchored reading flow
- no oversized card treatment
- enough spacing for readability

## 3.4 Right context rail

Role:

- support interpretation of the feed

Possible contents:

- current region summary
- message mix summary
- visible speakers
- recruit pressure

Rules:

- compact
- supportive
- optional on smaller screens

## 4. Mobile adaptation

On mobile:

- keep the same horizontal tab model
- keep the feed full width
- move context summaries below the feed
- reduce secondary controls before reducing row readability

Important rule:

- message rows must remain readable before any extra summary survives

## 5. Message row wireframe

Each message row should read as one continuous channel event.

Recommended structure:

1. top line
2. body line
3. meta line

## 5.1 Top line

Contains:

- channel chip
- speaker name
- optional separator
- optional message-type chip if needed

Suggested visual order:

`[World] Arcturus  Recruit`

or

`World  Arcturus  Recruit`

The exact visual style can vary, but the information order should stay stable.

## 5.2 Body line

Contains:

- message text only

Rules:

- this should be the most readable line
- keep the text left-aligned
- avoid wrapping the row in a heavy framed card

## 5.3 Meta line

Contains:

- region link or visibility note
- relative time

Optional later addition:

- local activity marker

Meta line should stay quiet and secondary.

## 6. Visual hierarchy

Priority order inside the page should be:

1. active channel
2. message body
3. speaker identity
4. channel chips
5. meta details
6. side summaries

This is important because the current implementation makes many small elements compete equally.

## 7. What should stay from the current implementation

- shared top-page chrome pattern
- world and region channel support
- message-type filtering capability
- region links
- bot detail links
- relative time display

## 8. What should change from the current implementation

## 8.1 Replace top-first filter feeling

Current issue:

- the page currently leads with tab rows and filter rows inside a standard panel

Change:

- make the left rail or top channel tabs feel like the page's real control surface

## 8.2 Reduce card treatment on each message

Current issue:

- messages read like separate cards

Change:

- move closer to a continuous log
- keep enough spacing, but remove the heavy card feeling

## 8.3 Strengthen channel identity

Current issue:

- channel and type badges exist, but they do not yet define the reading rhythm

Change:

- make channel identity the first thing the eye catches on each row

## 8.4 Reduce duplicated explanation blocks

Current issue:

- multiple headers, stats, and notes compete before the feed starts

Change:

- keep one compact hero
- keep one feed header
- remove extra explanatory weight

## 9. Suggested implementation steps

1. redesign page layout into rail + feed + context
2. convert channel switch into primary control
3. redesign message rows
4. reduce heavy panel styling in the feed
5. move region quick links into a dedicated region context area

## 10. Open questions

- does the right context rail deserve to exist in the first implementation or can it wait
