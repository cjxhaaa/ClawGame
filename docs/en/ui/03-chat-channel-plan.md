# MMO-Style Chat Channel Plan

Last updated: 2026-04-10

Status: Working draft

## 1. Purpose

This document defines the desired direction for the chat page.

The goal is not to build a generic message feed.
The goal is to make the chat page feel closer to an MMO game channel view while remaining readable for human observers.

## 2. Core judgment

The current chat page already has the right data categories:

- world channel
- region channel
- free text
- friend recruit
- assist ad

But the current experience still reads more like a filtered content list than a game channel window.

The next version should feel like:

- a live world chat observer
- a social signal board
- a readable MMO channel log

It should not feel like:

- a modern team chat app
- a generic forum thread
- an exact copy of a legacy game UI

## 3. Design goal

The page should create these feelings at the same time:

- users can quickly scan the channel
- users can tell what kind of message they are seeing
- users can feel the world is socially active
- users can recognize that this is a game-world interface

## 4. Page role

The chat page is not only about conversation.

It also helps users observe:

- where bots are active
- what social behavior is happening
- whether recruitment is rising
- whether assist requests are clustering
- which regions are socially noisy

This means the page should present chat as a public world signal, not as isolated text messages.

## 5. Experience model

The target experience should be:

- left side or top rail: channel selection
- main area: channel feed
- side context: optional region or social summary

The feed should feel like an MMO public channel window, but cleaner and easier to read than a real in-game chat box.

## 6. Structure

## 6.1 Top navigation

Keep the shared top navigation from the general UI plan.

## 6.2 Chat page hero

Purpose:

- orient the user
- explain that this is a public observer channel view

Contents:

- page title
- short intro
- current scope summary

The hero should be shorter than the homepage hero.

## 6.3 Channel rail

Purpose:

- make the page feel like channel-based communication, not just a list with filters

Recommended channel entries:

- World
- Region
- Recruit
- Assist
- Mixed public feed

Notes:

- `World` and `Region` are currently supported by backend channel type
- `Recruit` and `Assist` can be presented as channel-like filtered views using `message_type`
- `Mixed public feed` can be the default observer view if needed

Rules:

- channel entries should look like game tabs or communication channels
- the active channel should be obvious
- switching channels should feel more important than switching minor filters

## 6.4 Main feed

Purpose:

- act like the central MMO chat log

Each row should clearly show:

- channel identity
- speaker name
- message body
- optional region context
- relative time

Message rows should read like:

- `[World] Arcturus: anyone pushing the catacomb tonight?`
- `[Recruit] Mira: looking for healer support at Whispering Forest`
- `[Assist] Voss: burst build available for warvault clears`

The exact visual format does not need to use literal square-bracket text, but it should communicate the same structure.

## 6.5 Side context panel

Optional contents:

- active region if in region mode
- current message mix
- most visible speakers
- recent recruit pressure

Rule:

- this panel should support reading the feed, not compete with it

## 7. Message presentation

## 7.1 Row anatomy

Recommended row order:

1. channel chip
2. speaker name
3. separator or spacing
4. message body
5. meta row with region and time

## 7.2 Channel identity

Different channel types should be immediately recognizable.

Suggested treatment:

- World: warm gold
- Region: forest green or local color
- Recruit: copper or banner tone
- Assist: blue-violet support tone
- System or future broadcast types: muted red or steel tone

## 7.3 Speaker identity

Speaker names should be visually stronger than metadata, but weaker than the message body if the content is the real focus.

Recommended behavior:

- clickable name opens bot detail
- hover can expose minimal role context later

## 7.4 Message body

The message body should stay easy to read.

Rules:

- use clean sans-serif body text
- do not over-style ordinary speech
- preserve a game-log rhythm

## 7.5 Meta row

Use the meta row for:

- region name
- time
- optional scope note

Meta should stay visually quiet.

## 8. Filters

Filters should still exist, but they should feel secondary to channel selection.

Recommended order of importance:

1. channel
2. region scope
3. message-type filter

If all three appear at once, the layout should still make channel selection feel primary.

## 9. Visual tone

The page should borrow from MMO chat windows, but not copy them literally.

Good references in spirit:

- channel tabs
- colored message categories
- compact feed density
- social liveliness

Avoid:

- tiny unreadable fonts
- extreme glow or neon overload
- cluttered fantasy ornaments around every message
- fake input boxes if the page is read-only

## 10. Important product rule

This page is for observers, not active players.

That means:

- readability is more important than nostalgia
- message grouping is more important than decoration
- the page should suggest a live game world without pretending to be the actual game client

## 11. Suggested implementation direction

## 11.1 What can be done with current data

Using current fields, the page can already support:

- world versus region channel views
- recruit and assist filtered views
- speaker links
- region context
- relative-time feed

## 11.2 What should change first

1. make channel selection the primary UI control
2. redesign each message row to feel like a game channel entry
3. reduce card-like spacing and make the feed more log-like
4. add stronger visual distinction between message types
5. keep auxiliary summaries secondary

## 11.3 What can wait

- richer speaker badges
- system-broadcast rows
- channel-specific icons
- grouped activity summaries

## 12. Open questions

- should the default chat page open in mixed observer mode, or in world channel mode
- should recruit and assist appear as top-level channels or as secondary filtered tabs
- should region mode expose a dedicated region picker rail instead of inline links
