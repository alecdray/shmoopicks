---
description: >
  What Wax can do today — shipped, production features and their scope. Belongs here: feature
  descriptions, user-facing behaviour, and constraints of live functionality. Does not belong here:
  planned or in-queue work (→ roadmap), implementation internals (→ architecture), or data
  structure definitions (→ data-model).
links:
  - roadmap
  - data-model
  - frontend
  - integrations
---

[wiki](../wiki.md)

# Features

Shipped features in production today.

## My Library

The core of the app. A user's library is their collection of music — albums, artists, tracks, and releases (format variants: digital, vinyl, CD, cassette).

- Digital media is automatically synced from Spotify on a recurring schedule
- Users can browse and sort their library by title, artist, rating, date added, and last played
- Albums open in Spotify directly from the library

---

## Rankings & Reviews

Users can rate and review albums in their library.

- 0–10 score per album, manually set or guided by a 3-question questionnaire
- The questionnaire evaluates: track consistency, emotional impact, and gut reaction
- Free-text review notes attached to the rating
- Rating and review are independently updatable

---

## Tagging

Users can apply custom tags to albums for flexible organization and discovery.

- Tags belong to tag groups or stand alone
- No limit on tags per album or tags per group
- Two built-in tag group concepts: **Sound** (genre, style, influences) and **Mood** (context, feeling, occasion)
- Users define their own tags within these groups

