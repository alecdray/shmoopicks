# Mobile Support — Implementation Plan

## Approach

The implementation is a progressive enhancement of the existing server-rendered stack — no new frameworks or client-side routing. Changes fall into four areas:

1. **Layout responsiveness** — swap `h-screen overflow-hidden` for `h-dvh` on desktop and let mobile scroll naturally; make the header sticky at mobile breakpoints.
2. **Albums table column collapse** — hide the Formats, Date Added, Last Played, and Tags columns below `md:` and expose those fields through the existing actions dropdown.
3. **Infinite scroll (lazy loading)** — serve only the first batch on page load; a hidden sentinel row at the end of `<tbody>` triggers HTMX to load the next page when it enters the viewport.
4. **Housekeeping** — remove the dead Font Awesome CDN `<script>` tag.

This approach keeps all rendering server-side, reuses existing HTMX partial-swap patterns, and requires no new JavaScript. The feeds dropdown polling issue (fires when closed) is noted but is a pre-existing limitation deferred to future work — fixing it requires Alpine.js visibility tracking or a mutation observer.

## Files to Change

| File | Change |
|------|--------|
| `src/internal/core/templates/root.templ` | Remove Font Awesome `<script>` tag; change `body` to `min-h-dvh` instead of `h-screen` |
| `src/internal/library/adapters/dashboard.templ` | Column collapse in `albumRow` / `AlbumsTable`; infinite-scroll sentinel row; `DashboardPage` layout switch (`h-dvh` on desktop, natural scroll on mobile); sticky header on mobile |
| `src/internal/library/adapters/http.go` | `GetDashboardPage` passes first batch only; new `GetAlbumsPage` handler for paginated batches; `GetAlbumsTable` passes `offset` param through |
| `src/internal/server/server.go` | Register new `GET /app/library/dashboard/albums-page` route |
| `src/internal/library/service.go` | Add `GetAlbumsPagedInLibrary(ctx, userId, sortBy, dir, limit, offset)` method (or add `limit`/`offset` to existing `GetAlbumsInLibrary`) |

No database migrations are needed. The existing queries already return full album data; pagination is applied in the service layer using Go slice operations (all albums are fetched from the DB once and sliced) — this avoids complex SQL pagination with the current `sqlc.slice` join pattern.

## Implementation Steps

1. **Remove dead Font Awesome script tag**
   In `root.templ`, delete the `<script src="https://kit.fontawesome.com/03c1599055.js" ...>` line. Run `task build/templ`.

2. **Fix viewport height on iOS Safari**
   In `root.templ`, change `body` classes from `h-screen` to `min-h-dvh`. In `DashboardPage` in `dashboard.templ`, change the outer wrapper from `h-screen w-full flex flex-col overflow-hidden` to `h-dvh w-full flex flex-col overflow-hidden md:h-dvh` and remove `overflow-hidden` at small breakpoints using `overflow-auto md:overflow-hidden`. The inner scroll div should remain `overflow-y-auto` only at `md:` and above.

3. **Make the header bar sticky on mobile**
   In `DashboardHeaderBar`, add `sticky top-0 z-10` classes (already has `bg-base-100` to avoid bleed-through). On desktop the outer `flex flex-col` + fixed-height header already achieves the same effect.

4. **Collapse album table columns on mobile**
   In `albumRow`, wrap the Formats `<td>`, Date Added `<td>`, Last Played `<td>`, and Tags `<td>` with `class="hidden md:table-cell"`. In the `AlbumsTable` header row, wrap the matching `<th>` elements the same way. The three-dot actions dropdown already contains Notes and Tags; add Formats (as `releaseFormatBadge` badges) and Date Added / Last Played as read-only lines inside the dropdown for mobile-only display using `class="md:hidden"` items in the `<ul>`.

5. **Add pagination constant and service method**
   In `service.go`, add:
   ```go
   const AlbumsPageSize = 20
   ```
   Add a `GetAlbumsInLibraryPaged(ctx, userId, limit, offset int)` method that calls the existing `GetAlbumsInLibrary` and slices the result. This is intentionally simple — the full album list is already fetched from the DB in one round-trip, so slicing in Go is cheap and correct with the existing sort logic.

6. **Add infinite-scroll sentinel to the table template**
   In `albumsTableBody`, after the last `@albumRow(album)`, add a sentinel `<tr>` when the current batch might have more rows (i.e., `len(albums) == AlbumsPageSize`):
   ```html
   <tr id="albums-load-more-sentinel"
       hx-get="/app/library/dashboard/albums-page?offset=N&sortBy=...&dir=..."
       hx-trigger="revealed"
       hx-swap="outerHTML"
       hx-target="#albums-load-more-sentinel">
     <td colspan="8"></td>
   </tr>
   ```
   The `offset` and sort params need to be threaded through. `albumsTableBody` must accept an `offset int`, `sortBy string`, `sortDir string` so it can construct the next URL. When `len(albums) < AlbumsPageSize`, the sentinel is omitted (end of list).

7. **Add `GetAlbumsPage` HTTP handler**
   In `http.go`, add `GetAlbumsPage` that reads `offset`, `sortBy`, `dir` query params, fetches the paged slice, and renders only `albumsTableBody(albums, nextOffset, sortBy, dir)` (no table wrapper — it replaces only the sentinel row and appends new rows via `hx-swap="outerHTML"` on the sentinel itself). Register `GET /app/library/dashboard/albums-page` in `server.go`.

8. **Update `GetDashboardPage` to serve first batch only**
   Replace `AlbumsTable(props.Library.Albums, ...)` with `AlbumsTable(props.Library.Albums[:min(AlbumsPageSize, len)], ...)` — passing only the first page to the initial render. `GetAlbumsTable` (sort re-fetch) should also be updated to slice to `AlbumsPageSize` and include the sentinel.

9. **Run `task build/templ` and smoke-test**
   Verify the desktop layout is unchanged, then test on a mobile-width viewport (375px) in browser dev tools: columns should collapse, infinite scroll should fire, header should stay sticky.

## Database Changes

None. Pagination is implemented by slicing the in-memory album slice returned by the existing DB queries. If performance profiling later shows the full-library DB fetch is too slow for large libraries, a SQL-level `LIMIT`/`OFFSET` query can be added as a follow-up.

## Testing

**Manual — Desktop (baseline)**
- Dashboard loads and looks identical to current state at `md:` and above.
- Sort headers work and reload the table correctly.
- Modals (rating, notes, tags) open and close correctly.

**Manual — Mobile (375px viewport, Chrome DevTools)**
- Page scrolls naturally (no double scrollbar, no cut-off by iOS address bar simulation).
- Header sticks to top when scrolling down.
- Album table shows only Album, Artists, Rating, and the actions menu column.
- Three-dot menu on a row reveals format, date added, last played, and tags entries.
- On page load, only 20 rows are rendered. Scrolling to the bottom triggers a new batch to appear.
- Subsequent batches append correctly; sentinel disappears after the last batch.

**Manual — Carousel**
- Carousel strip scrolls horizontally with swipe on mobile.

**Manual — Housekeeping**
- Browser network tab shows no request to `kit.fontawesome.com`.

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| `hx-trigger="revealed"` fires immediately if the sentinel is already in viewport on first load (small library) | Acceptable — it will just fetch an empty next page and remove the sentinel harmlessly |
| Slicing the full in-memory album list works now but will be slow for very large libraries (1000s of albums) | Noted as a known tradeoff; SQL-level pagination is a clear follow-up path |
| Sort state is carried only via query params; switching sort columns reloads the full table (resetting infinite scroll back to page 1) | This is the existing behavior and is correct — a full table swap resets to offset 0 |
| `overflow-hidden` removal on mobile could break the inner `overflow-y-auto` scroll behavior on some browsers | Test carefully; the `md:overflow-hidden` guard preserves desktop behavior |
| Feeds dropdown polling fires while closed | Pre-existing issue, out of scope for this PR |

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan mobile-support -->
