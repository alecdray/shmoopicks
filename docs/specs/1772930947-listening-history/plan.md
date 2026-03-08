# Listening History — Implementation Plan

## Approach

Create a new `listeninghistory` package (parallel to `feed`, `library`, `review`) that owns all play-history logic: DB table, service, and cron task. The package depends on `spotify.Service` for data fetching and shares the existing `GetOrCreateAlbum` / `GetOrCreateTrack` / `GetOrCreateAlbumTrack` / `GetOrCreateArtist` / `GetOrCreateAlbumArtist` queries already used in the library feed. The `library` package gains a `LastPlayedAt *time.Time` field on `AlbumDTO` and a new sort method, populated inside `GetAlbumsInLibrary` via a separate query + in-memory merge — the same pattern used for ratings (via `review`).

**Key decisions:**

- **New package, not extending `feed`**: The feed module is exclusively about saved-library sync. Play history is orthogonal and is best kept separate.
- **`track_plays` table with denormalized `album_id`**: Each `RecentlyPlayedItem` carries the specific album context (`SimpleTrack.Album`). Storing `album_id` directly avoids a join through `album_tracks` and is safe because the album identity comes from Spotify at play time.
- **`listeninghistory` imported directly into `library.Service`**: `listeninghistory` only imports `db` and `spotify` — no cycle exists. This mirrors how `library` already imports `review`: `GetAlbumsInLibrary` calls `listeninghistory.Service.GetLastPlayedAtByAlbumIds` and merges the results into `AlbumDTO` before returning. No handler-level merge is needed.
- **Separate query + in-memory merge for `LastPlayedAt`**: Consistent with how `GetAlbumsInLibrary` already assembles ratings, artists, tracks, and releases. Avoids a complex JOIN that would break the existing query structure.
- **Sync state tracked as a single `last_synced_at` column on a `listening_history_syncs` table**: Simpler than replicating the full `feeds` status machine. The window passed to `GetRecentlyPlayedTracks` is `time.Since(lastSyncedAt) + 1 hour buffer`. For a first sync with no prior state, use a 2-hour window.
- **Ticker shows recently played albums from `track_plays`, deduplicated by album, limited to 20**: Includes albums not in the library. Rendered server-side on page load. Deduplication is done in Go (keep the most recent play per album).
- **Cron schedule `0 * * * *`** (every hour on the hour).
- **`GetUsersWithSpotifyToken`** — new query needed; returns all `users` where `spotify_refresh_token IS NOT NULL AND deleted_at IS NULL`.

## Files to Change

| File | Change |
|------|--------|
| `db/migrations/<ts>_add_track_plays.sql` | New migration: `track_plays` table + `listening_history_syncs` table |
| `db/queries/track_plays.sql` | New file: `UpsertTrackPlay`, `GetLastPlayedAtByAlbumIds`, `GetRecentlyPlayedAlbums`, `GetListeningHistorySyncState`, `UpsertListeningHistorySyncState` |
| `db/queries/users.sql` | Add `GetUsersWithSpotifyToken` query |
| `src/internal/listeninghistory/service.go` | New package: `Service` struct, `UpsertPlayHistory`, `GetLastPlayedAtByAlbumIds`, `GetRecentlyPlayedAlbums`, `GetOrCreateSyncState`, `UpdateSyncState` |
| `src/internal/listeninghistory/task.go` | New: `SyncListeningHistoryTask` cron task (hourly) |
| `src/internal/library/service.go` | Add `LastPlayedAt *time.Time` to `AlbumDTO`; add `listeninghistory.Service` dependency; extend `GetAlbumsInLibrary` to call `GetLastPlayedAtByAlbumIds` and merge into DTOs; add `AlbumDTOs.SortByLastPlayed` |
| `src/internal/library/adapters/http.go` | Add `"lastPlayed"` case in `GetAlbumsTable`; fetch recently played albums in `GetDashboardPage` and pass through `DashboardPageProps` |
| `src/internal/library/adapters/dashboard.templ` | Add "Last Played" sortable column to `AlbumsTable` / `albumRow`; add `RecentlyPlayedTicker` component; update `DashboardPage` and `DashboardPageProps` |
| `src/internal/server/server.go` | Instantiate `listeninghistory.Service`; add to `services` struct; register `SyncListeningHistoryTask`; pass service to `NewLibraryService` |
| `static/src/main.css` | Add `@keyframes` marquee animation for ticker bar |

## Implementation Steps

1. **Database migration** — create migration with `task db/create -- add_track_plays`, define `track_plays` and `listening_history_syncs` tables, run `task db/up`.

2. **SQL queries** — write `db/queries/track_plays.sql` (five queries) and add `GetUsersWithSpotifyToken` to `db/queries/users.sql`. Run `task build/sqlc`.

3. **`listeninghistory` service** — create `src/internal/listeninghistory/service.go` with `Service`, constructor, and all business-logic methods. No HTTP adapters needed in this module.

4. **`listeninghistory` cron task** — create `src/internal/listeninghistory/task.go` implementing `task.Task`.

5. **`AlbumDTO` extension** — add `LastPlayedAt *time.Time` to `AlbumDTO` in `library/service.go`; add `listeninghistory.Service` as a dependency on `library.Service`; extend `GetAlbumsInLibrary` to call `GetLastPlayedAtByAlbumIds` and merge into DTOs (same location as the existing ratings merge); add `SortByLastPlayed` to `AlbumDTOs`.

6. **HTTP handler update** — add `"lastPlayed"` sort case in `GetAlbumsTable`; fetch recently played albums via `listeninghistory.Service.GetRecentlyPlayedAlbums` in `GetDashboardPage` and pass through `DashboardPageProps`.

7. **Template update** — add "Last Played" column header and cell to `AlbumsTable` / `albumRow`; add `RecentlyPlayedTicker` templ component; insert ticker into `DashboardPage` between header bar and library stats. Run `task build/templ`.

8. **CSS** — add marquee `@keyframes` + `.ticker-track` animation class to `static/src/main.css`.

9. **Server wiring** — update `server.go` to instantiate `listeninghistory.Service`, pass it into `NewLibraryService`, and register `SyncListeningHistoryTask`.

10. **Build & smoke test** — run `task build` to verify no compile errors, then manually test.

## Database Changes

### New migration: `add_track_plays`

```sql
-- +goose Up

CREATE TABLE track_plays (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    played_at datetime not null,
    unique(user_id, track_id, played_at)
);

CREATE TABLE listening_history_syncs (
    user_id text primary key references users(id) on delete cascade,
    last_synced_at datetime not null
);

-- +goose Down

DROP TABLE track_plays;
DROP TABLE listening_history_syncs;
```

### New file: `db/queries/track_plays.sql`

```sql
-- name: UpsertTrackPlay :exec
INSERT INTO track_plays (id, user_id, track_id, album_id, played_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (user_id, track_id, played_at) DO NOTHING;

-- name: GetLastPlayedAtByAlbumIds :many
SELECT album_id, MAX(played_at) as last_played_at
FROM track_plays
WHERE user_id = ? AND album_id IN (/*SLICE:album_ids*/?)
GROUP BY album_id;

-- name: GetRecentlyPlayedAlbums :many
SELECT albums.*, MAX(track_plays.played_at) as last_played_at
FROM track_plays
JOIN albums ON albums.id = track_plays.album_id
WHERE track_plays.user_id = ?
GROUP BY albums.id
ORDER BY last_played_at DESC
LIMIT 20;

-- name: GetListeningHistorySyncState :one
SELECT * FROM listening_history_syncs WHERE user_id = ?;

-- name: UpsertListeningHistorySyncState :exec
INSERT INTO listening_history_syncs (user_id, last_synced_at)
VALUES (?, ?)
ON CONFLICT (user_id) DO UPDATE SET last_synced_at = EXCLUDED.last_synced_at;
```

### Addition to `db/queries/users.sql`

```sql
-- name: GetUsersWithSpotifyToken :many
SELECT * FROM users
WHERE spotify_refresh_token IS NOT NULL AND deleted_at IS NULL;
```

## Testing

**Cron task / sync:**
- Connect a Spotify account, wait for (or manually trigger) the hourly task.
- Verify rows appear in `track_plays` with correct `user_id`, `track_id`, `album_id`, `played_at`.
- Run the task twice; confirm `ON CONFLICT DO NOTHING` prevents duplicates.

**Last Played column:**
- Play an album on Spotify, trigger a sync, reload dashboard.
- Confirm the "Last Played" column shows the correct date for albums in your library.
- Click the column header; confirm sort ascending/descending works. Albums with no plays should sort to the bottom.

**Ticker:**
- After a sync, verify the ticker bar shows recently played album art/titles scrolling.
- Verify albums NOT in the library still appear in the ticker.
- Verify deduplication: the same album appears only once even if many tracks were played.

**Edge cases:**
- User with no play history: ticker should be hidden or show empty state; "Last Played" cells should be empty.
- First sync (no prior `listening_history_syncs` row): should succeed with a 2-hour window.

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| `GetAlbumsInLibrary` becomes slow with an extra query per page load | The new `GetLastPlayedAtByAlbumIds` is a single batched query (using sqlc `/*SLICE:*/`). Impact should be minimal. |
| `GetRecentlyPlayedAlbums` returns albums whose `album_artists` aren't populated (needed for ticker display) | Ticker only needs album title + image URL, both on the `albums` row — no join to `album_artists` required. |
| Ticker CSS animation interacts poorly with existing layout (height, overflow) | Contain the ticker in a fixed-height `overflow-hidden` div; test across screen sizes. |
| Multi-user task fails for one user and aborts remaining users | Log the error and `continue` to the next user rather than returning early. |
| Spotify `Client()` returns `ErrFailedToGetToken` for users with an expired/revoked token | Detect and skip those users in the task; log a warning. |

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan listening-history -->
