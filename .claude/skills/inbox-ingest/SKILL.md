---
name: inbox-ingest
description: Ingest pulled inbox notes from tmp/inbox.md into the wiki, then clear the file. Keywords: ingest inbox, process inbox, inbox to wiki.
allowed-tools: [Read, Write]
---

# Inbox Ingest Skill

Process the contents of `./tmp/inbox.md` by ingesting each note into the wiki, then clearing the file.

## Steps

1. Read `./tmp/inbox.md`. If the file is empty or contains no substantive notes (only test messages or the file doesn't exist), stop and tell the user.

2. Run `/wiki-ingest ./tmp/inbox.md` to ingest the content into the wiki.

3. After wiki-ingest completes, clear `./tmp/inbox.md` by overwriting it with empty content:

```
# Wax Inbox
```

4. Tell the user what was ingested and confirm the inbox was cleared.
