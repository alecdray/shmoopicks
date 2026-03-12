#!/bin/bash
# Uploads a SQL dump to a remote SQLite db over SSH.
# Usage: ./upload.sh <dump.sql> <user@host:/path/to/db.sql>
set -euo pipefail

DUMP_FILE="${1:?Usage: $0 <dump.sql> <user@host:/path/to/db.sql>}"
DEST="${2:?Usage: $0 <dump.sql> <user@host:/path/to/db.sql>}"

HOST="${DEST%%:*}"
REMOTE_PATH="${DEST#*:}"

read -r -s -p "sudo password for $HOST: " SUDO_PASS
echo

REMOTE_TMP=$(ssh "$HOST" "TMPFILE=\$(mktemp) && chmod 600 \"\$TMPFILE\" && echo \"\$TMPFILE\"")

echo "Uploading $DUMP_FILE to $HOST:$REMOTE_TMP..."
scp "$DUMP_FILE" "$HOST:$REMOTE_TMP"

echo "Backing up and applying on $HOST..."
printf '%s\n' "$SUDO_PASS" | ssh "$HOST" "sudo -S sh -c \"cp '$REMOTE_PATH' '${REMOTE_PATH}.bak' && sqlite3 '$REMOTE_PATH' < '$REMOTE_TMP'\" && rm '$REMOTE_TMP'"

echo "Done."
