# Listening History — Research

## Summary

The listening history feature (v1) pulls a user's recently played tracks from Spotify on an hourly schedule, persists them, and surfaces this data in two places on the dashboard: a "last played at" column in the albums table (sortable), and a scrolling stock-ticker-style bar at the top of the dashboard showing recently listened-to albums. Spotify caps the recently-played endpoint at 50 tracks, so the hourly polling strategy is essential to build up a longer history over time by accumulating entries before they fall off the API window.

## Relevant Code

### Spotify layer
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/spotify/spotify.go` — `Service.GetRecentlyPlayedTracks()` already exists. It calls `client.PlayerRecentlyPlayedOpt` with a `BeforeEpochMs` cursor, collects up to 50 items per call within a time window, and returns `[]spotify.RecentlyPlayedItem`. Each item has `.Track` (a `SimpleTrack` with `.Album SimpleAlbum`, `.Artists []SimpleArtist`, `.ID`), and `.PlayedAt time.Time`.
- `ScopeUserReadRecentlyPlayed` is already registered in `server.go` when constructing the SpotifyAuth service, so no OAuth scope changes are needed.

### Task / scheduling layer
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/core/task/task.go` — `TaskManager` supports both cron tasks (registered via `RegisterCronTask`) and ad-hoc tasks (via `RegisterAdHocTask`). Cron tasks implement the `Task` interface: `Run(ctx) error`, `Schedule() *CronExpression`, `Name() string`.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/feed/task.go` — `SyncStaleSpotifyFeedsTask` is the reference cron task pattern. It runs `* * * * *` (every minute), queries for stale feeds, and calls the feed service for each. A new `SyncListeningHistoryTask` should follow this exact pattern.

### Feed / sync layer
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/feed/service.go` — `SyncSpotifyFeed` is the reference for how to structure a sync function: mark syncing, do work, mark success/failure. The listening history sync would live in a new module (e.g. `listeninghistory`) with its own service, or could extend the feed service.

### Database
- `/Users/shmoopy/Documents/code/repos/shmoopicks/db/schema.sql` — Current schema has `tracks` (id, spotify_id, title), `albums` (id, spotify_id, title, image_url), `user_tracks` (user_id, track_id, added_at), `user_releases` (user_id, release_id, added_at). There is no `played_at` concept — `added_at` on `user_tracks` is when the track was saved to the library, not when it was played.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/db/queries/user_tracks.sql` — `GetOrCreateUserTrack` and `GetUserTracks` exist but track saved-library membership, not play history.
- A new migration is needed to add a `listening_history` table (or `track_plays`) with columns: `id`, `user_id`, `track_id` (FK → tracks), `album_id` (FK → albums), `played_at datetime`. This lets us efficiently query "last played at" per album for a user.

### Library / dashboard adapter
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/service.go` — `AlbumDTO` and `GetAlbumsInLibrary` are the core data structures for the dashboard. `AlbumDTO` will need a `LastPlayedAt *time.Time` field added so it can be passed to templates and sorted.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/adapters/http.go` — `GetAlbumsTable` handles sort-by parameter dispatch. A new `"lastPlayed"` case needs to be added.
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/library/adapters/dashboard.templ` — `AlbumsTable` and `albumRow` templates need a new "Last Played" column. The ticker bar would be a new templ component in this file (or a dedicated file), rendered above or below the `DashboardHeaderBar`.

### Server wiring
- `/Users/shmoopy/Documents/code/repos/shmoopicks/src/internal/server/server.go` — `NewServices` is where the new listening history service would be instantiated and its cron task registered via `s.taskManager.RegisterCronTask(...)`.

## Architecture

### Data flow for hourly sync

```
TaskManager (cron, every hour)
  → SyncListeningHistoryTask.Run()
    → for each user with a Spotify token:
        spotify.Service.GetRecentlyPlayedTracks(ctx, userId, window)
          → Spotify API: /me/player/recently-played (max 50 tracks)
        listeninghistory.Service.UpsertPlayHistory(ctx, userId, tracks)
          → GetOrCreateAlbum / GetOrCreateTrack (reuse existing queries)
          → INSERT INTO track_plays ... ON CONFLICT DO NOTHING
```

### Data flow for dashboard display

```
GET /app/library/dashboard
  → libraryHandler.GetDashboardPage()
    → libraryService.GetAlbumsInLibrary()  [extended to JOIN last played]
    → listeningHistoryService.GetLastPlayedAtByAlbum(ctx, userId)
    → merge into AlbumDTO.LastPlayedAt
    → DashboardPage template
        → ticker bar component (recently played albums, scrolling)
        → AlbumsTable (with new "Last Played" sortable column)
```

### Module options

Option A: New `listeninghistory` package (parallel to `feed`, `library`) — cleanest separation, owns its DB table, service, task.

Option B: Extend `feed` package — somewhat forced since feed is about saved-library sync, not play history.

Option A is strongly preferred given the existing module structure.

## Existing Patterns

- **Cron task**: implement `task.Task` interface with a `Schedule()` returning a `CronExpression`. Register via `taskManager.RegisterCronTask()` in `server.go`. Reference: `feed.SyncStaleSpotifyFeedsTask`.
- **Service struct**: business logic in a `Service` struct with a `*db.DB` field and a `*spotify.Service` dependency. Constructor is `NewService(db, spotifyService)`.
- **DB migrations**: `task db/create -- migration_name` creates a new goose migration file. `task db/up` applies it.
- **SQL queries**: `.sql` files in `db/queries/`, regenerated with `task build/sqlc`. Use `ON CONFLICT ... DO UPDATE` for upserts.
- **DTO pattern**: domain models are DTOs (e.g. `AlbumDTO`), not raw sqlc types. Sorting methods live on the slice type (e.g. `AlbumDTOs.SortByDate`).
- **Templates**: `.templ` files, regenerated with `task build/templ`. Component props are plain Go structs. HTMX is used for interactivity.
- **Context**: `contextx.ContextX` is used throughout; `ctx.UserId()` extracts the authenticated user.
- **Error handling**: wrap errors with `fmt.Errorf("...: %w", err)`, return early. HTTP handlers call `http.Error` directly (some modules use `httpx.HandleErrorResponse()`).
- **Scrolling ticker**: no existing example in the codebase. Will need a CSS `@keyframes` marquee animation added to `static/src/main.css`, or use Tailwind's animation utilities. The ticker would be an HTMX-polled component (like `FeedsDropdownContent`) or rendered server-side on page load.

## Constraints & Risks

- **Spotify 50-track cap**: `GetRecentlyPlayedTracks` already handles this with a cursor loop, but is still bounded by whatever Spotify stores. If the cron runs less frequently than the user's listening rate * 50 tracks, history gaps will form. Running hourly should be sufficient for most users, but heavy listeners could saturate the 50-item window quickly.
- **Multi-user task**: The current `SyncStaleSpotifyFeedsTask` iterates all stale feeds (which implicitly handles multiple users). The new listening history task needs to iterate all users who have a Spotify token. Need a query like `GetUsersWithSpotifyToken` — does not currently exist.
- **Albums not in library**: Recently played tracks may include albums not in the user's saved library. The ticker and "last played at" column should work for these as well if we store album data, but the dashboard table only shows albums in the user's library (via `user_releases`). The "last played at" column on the table would only apply to albums already in the library. The ticker can show any recently played album.
- **Track → Album mapping**: `RecentlyPlayedItem.Track` is a `SimpleTrack` which has `.Album SimpleAlbum` (with `.ID`, `.Name`, `.Images`). Album data is readily available per track without extra API calls.
- **`AlbumDTO.LastPlayedAt` join**: `GetAlbumsInLibrary` currently does several separate queries. Adding last-played-at either requires a new query that JOINs `track_plays` with albums, or a separate `GetLastPlayedAtByAlbumIds` query that returns a map. The latter (separate query + in-memory merge) matches the existing pattern in `GetAlbumsInLibrary`.
- **Cron timing**: the existing feed cron runs every minute. Running listening history sync every hour means using `0 * * * *` as the cron expression. The `robfig/cron/v3` library supports standard 5-field cron syntax.
- **Ticker UI**: no existing scrolling animation component in the codebase. This is the highest-effort UI piece and may need custom CSS.

## Open Questions

1. Should listening history be its own package (`listeninghistory`), or does it make sense to fold it into `feed` or `library`?
2. Should the `track_plays` table store the album_id directly (denormalized for query efficiency), or just track_id and join through album_tracks? Given that a track can appear on multiple albums and `SimpleTrack.Album` gives us the specific album context from the play, denormalizing with album_id seems right.
3. Should the ticker show only albums that are in the user's library, or all recently played albums (including ones not in their library)? The feature doc doesn't specify.
4. What window should the hourly sync use? Since the last sync timestamp will be tracked, using `time.Since(lastSyncAt) + buffer` (matching the feed pattern with a 1-hour buffer) makes sense — but what is the initial sync window if there's no prior sync (all-time is not possible; Spotify only retains ~24h of recent plays regardless).
5. Is a separate sync-state table needed for listening history (similar to `feeds` with `last_sync_status`, `last_sync_completed_at`), or is tracking just `last_synced_at` per user sufficient?
6. How many albums should the ticker display? Should it deduplicate (show unique albums rather than every track play)?

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research  -->
