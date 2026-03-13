---
name: wiki-ingest
description: Ingest new information into the wiki. Takes notes, a file, or raw text and places the content in the right page(s), updating outdated information where found. Keywords: update wiki, add to wiki, ingest wiki, wiki update.
argument-hint: "<notes or file path>"
---

Ingest the following input into the wiki: {{args}}

## Steps

### 1. Read the wiki rules

Read `docs/wiki/README.md` to understand the wiki's conventions, scope, and structure before touching anything.

### 2. Read all page frontmatter

Read the frontmatter of every file in `docs/wiki/pages/` — the `description` field is sufficient to route information to the right page. Do not read full page content yet.

### 3. Parse the input

If the argument looks like a file path, read the file. Otherwise treat the argument as raw notes.

Extract every discrete piece of information from the input. For each piece, ask:
- Is this new information that doesn't exist in the wiki yet?
- Does this correct or update something already in the wiki?
- Does this contradict anything currently in the wiki?

### 4. Route each piece of information

For each piece of information, determine where it belongs using the `description` frontmatter of candidate pages as the routing guide. A piece of information belongs on the page whose description says it belongs there — and does *not* belong on pages whose descriptions explicitly exclude it.

If a piece of information doesn't fit any existing page, follow the guidance in the README.

### 5. Read target pages in full

Read the complete content of only the pages identified in step 4 as targets for changes. Check for anything the input might update, replace, or contradict.

### 6. Apply changes

Make all edits:
- Insert new content into the appropriate section of the appropriate page, following the existing style and tone
- Update or replace any outdated or incorrect content with the new information
- If a piece of information changes a link relationship, update the `links` frontmatter on affected pages
- Create new pages where warranted per the routing decision in step 4

### 7. Run wiki-audit

Run `/wiki-audit` to verify the wiki is fully compliant after the changes. Any violations it finds and fixes should be included in the summary below.

### 8. Output a review summary

After all edits are made, output a summary in this format:

---

## Wiki Ingest Summary

### Changes Made

For each page that was modified:

**`<page-name>.md`**
- What was added, changed, or removed and why

### New Pages Created

List any new pages created, with the rationale for why the "When to Break Out a New Page" criteria were met.

### Potential Issues

Flag anything that looked like a contradiction, a scope boundary question, or a judgement call that the user should review.

---
