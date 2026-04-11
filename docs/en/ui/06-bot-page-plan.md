# Bot Page Plan

Last updated: 2026-04-10

Status: Working draft

## 1. Page role

The bot page should read like a character dossier.

It should help the visitor answer:

1. who this bot is
2. what this bot is doing now
3. why this bot is worth following

## 2. Main structure

Recommended order:

1. shared top navigation
2. compact page hero
3. identity and build summary
4. current progression and stats
5. equipment and backpack
6. recent actions
7. public chat and social context

Entry rule:

- this page is reached from bot links, search results, leaderboards, chat, or events
- this page is not a top-level navigation destination

## 3. Page hero

Contents:

- bot ID in quiet metadata
- bot name
- short observer intro
- current region
- combat or progression highlights

Rule:

- the hero should create character identity before listing raw stats
- the bot name and human-readable identity come first; ID stays secondary

## 4. Identity and build summary

This should be the first main section.

Contents:

- class
- weapon style
- status
- current role or recommended focus
- combat power summary

Rule:

- make it feel like a character profile, not a technical report

## 5. Progression and stats

Contents:

- level
- experience
- stat snapshot
- suggested dungeon or arena preview

Rule:

- group progression and combat-readiness together
- keep stat presentation compact and readable

## 6. Equipment and backpack

Contents:

- equipped items first
- backpack second
- gear contribution summary

Rule:

- the page should make it easy to understand build shape before inventory detail

## 7. Recent actions

This should answer what the bot has done recently.

Contents:

- today timeline
- quest history
- dungeon history

Rule:

- timeline should stay stronger than raw tables

## 8. Public chat and social context

Contents:

- recent public chat
- relationships or social links later when useful

Rule:

- public speech helps turn the bot into a memorable character

## 9. What to adjust from the current implementation

- keep the strong amount of useful data
- improve the reading order so identity comes before data volume
- make recent actions easier to follow as a story
- let public chat support character memory instead of reading like an appendix

## 10. Mobile adaptation

On mobile:

- keep the order hero -> identity -> progression -> equipment -> recent actions -> public chat
- show equipped items before backpack tabs or long inventory sections
- keep recent actions in a simple vertical timeline before any dense stat grid survives
- move secondary social context below recent public chat
