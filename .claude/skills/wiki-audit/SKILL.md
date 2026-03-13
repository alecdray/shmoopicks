---
name: wiki-audit
description: Audit the wiki for compliance with its README rules and fix any violations found. Keywords: audit wiki, wiki compliance, check wiki, review wiki.
context: fork
agent: Explore
---

Audit the wiki at `docs/wiki/` for compliance with its conventions and fix any violations.

## Steps

### 1. Read the rules

Read `docs/wiki/README.md` in full. This is the source of truth for what correct looks like.

### 2. Read all pages

Read every file in `docs/wiki/pages/` and `docs/wiki/wiki.md` in parallel.

### 3. Audit each page against the rules

Using the rules from the README, audit every page for compliance.

### 4. Fix all violations

Make corrections directly. Do not ask for confirmation on clear rule violations — apply the fix and note it in the summary.

### 5. Output an audit summary

---

## Wiki Audit Summary

### Violations Fixed

For each issue that was corrected:

**`<page-name>.md`**
- What was wrong and what was changed

### No Issues Found

If a page was clean, list it here.

### Judgement Calls

Anything that was ambiguous or where a decision was made that the user should be aware of.

---
