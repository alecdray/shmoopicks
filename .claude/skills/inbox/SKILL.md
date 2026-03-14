---
name: inbox
description: Pull unread wax emails from Gmail and ingest them into the wiki in one step. Keywords: inbox, process email, check and ingest, full inbox run.
---

# Inbox Skill

Run the full inbox pipeline in sequence:

1. Run `/inbox-pull` to fetch unread wax-labeled emails from Gmail into `./tmp/inbox.md`
2. If any emails were pulled, present a summary of what was found — each note's date and content — and ask the user to confirm before proceeding
3. On confirmation, run `/inbox-ingest` to ingest them into the wiki and clear the file
4. Report the outcome of both steps
