---
name: wiki-search
description: Search the wiki for information about the product, architecture, or features. Use when the user asks what the wiki says about a topic, wants to find relevant wiki pages, or needs context from the wiki before answering a question. Keywords: search wiki, find in wiki, wiki says, what does wiki, wiki contains, look up wiki.
argument-hint: "<query>"
context: fork
agent: Explore
---

Search the wiki at `docs/wiki/` for information related to: {{args}}

## Steps

### 1. Read the entry point

Read `docs/wiki/wiki.md` to get the list of top-level pages. Then read the `description` frontmatter of each of those pages to understand their scope.

### 2. Identify candidate pages

Based on the query and each top-level page's `description` frontmatter, identify which pages are likely to contain relevant information.

### 3. Search page contents

Read each candidate page. Use Grep to search for specific terms if the query contains keywords. Focus on finding:
- Direct mentions of the topic
- Sections that discuss related concepts
- Decisions or rationale connected to the query

### 4. Follow relevant links

Each page has a `links` frontmatter field listing related pages by filename. After reading a candidate page, check its `links` field. For any linked pages not yet visited, read their frontmatter `description` to assess relevance before reading the full page. If the description suggests the page is relevant to the query, read it in full. Repeat until no new relevant links are found or all pages have been visited.

If the initial candidates don't yield results, broaden to other pages.

### 5. Return findings

Summarize what the wiki says about the topic. Structure the output as:

---

## Wiki Search: {{args}}

### Relevant Pages

For each page that contains relevant information:

**`<page-name>.md`**
> Quoted or paraphrased content, with enough context to be useful

### Summary

A brief synthesis of what the wiki collectively says about this topic.

### Not in Wiki

If the topic was not found or only partially covered, note what's missing and which page it would likely belong in if it were added.

---
