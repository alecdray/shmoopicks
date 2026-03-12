#!/bin/bash
# Dumps a wax SQLite db to a SQL file, stripping sensitive columns.
# Usage: ./dump.sh <source.db> <output.sql>
set -euo pipefail

SOURCE_DB="${1:?Usage: $0 <source.db> <output.sql>}"
OUTPUT="${2:?Usage: $0 <source.db> <output.sql>}"

SKIP_TABLES=("goose_db_version" "sqlite_sequence")

# Format: "table:col" — these columns will be exported as NULL
NULLIFY_COLS=("users:spotify_refresh_token")

echo "Dumping $SOURCE_DB -> $OUTPUT..."

{
  echo "PRAGMA foreign_keys = OFF;"
  echo "BEGIN TRANSACTION;"

  TABLES=$(sqlite3 "$SOURCE_DB" "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;")

  for table in $TABLES; do
    skip=false
    for skip_table in "${SKIP_TABLES[@]}"; do
      [[ "$table" == "$skip_table" ]] && skip=true && break
    done
    $skip && continue

    COLS=$(sqlite3 "$SOURCE_DB" "PRAGMA table_info($table);" | awk -F'|' '{print $2}')

    SELECT_EXPR=""
    COL_NAMES=""
    for col in $COLS; do
      nullify=false
      for entry in "${NULLIFY_COLS[@]}"; do
        [[ "$entry" == "$table:$col" ]] && nullify=true && break
      done
      if [[ -n "$SELECT_EXPR" ]]; then
        SELECT_EXPR+="||','||"
        COL_NAMES+=","
      fi
      if $nullify; then
        SELECT_EXPR+="'NULL'"
      else
        SELECT_EXPR+="quote(\"$col\")"
      fi
      COL_NAMES+="\"$col\""
    done

    echo "DELETE FROM \"$table\";"
    sqlite3 "$SOURCE_DB" \
      "SELECT 'INSERT INTO \"$table\" ($COL_NAMES) VALUES (' || $SELECT_EXPR || ');' FROM \"$table\";"
  done

  echo "COMMIT;"
  echo "PRAGMA foreign_keys = ON;"
} > "$OUTPUT"

echo "Done."
