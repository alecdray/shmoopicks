# My Library v1.1 — Research

## Summary

My Library v1.1 redesigns the core library experience in three parts: (1) replace the dense table view with a visual list view leading with album art and a prominent rating number, (2) add a chip-based faceted filter and sort UI where each chip opens the existing centered modal with its specific settings, and (3) clean up dashboard UX by removing all Spotify outlinks from the library list and removing the tags ellipsis (⋯) menu from rows. All filtering and sorting is in-memory on the server, paginated via infinite scroll, with state reflected in URL params.

---

## Relevant Code

### Library module
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/library/service.go` — `Library`, `AlbumDTO`, `AlbumDTOs`, and all sort methods (`SortByTitle`, `SortByArtist`, `SortByRating`, `SortByDate`, `SortByLastPlayed`). `GetAlbumsInLibrary` loads all albums for a user in one pass and hydrates ratings, tags, and last-played timestamps in bulk. `Library.Artists` provides the distinct artist list needed for the artist filter.
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/library/adapters/dashboard.templ` — all current templates: `albumRow`, `albumsTableBody`, `AlbumsTable`, `albumsTableSortableHeader`, `AlbumRating`, `AlbumTagsCell`, `CarouselSection`, `LibraryStats`, `DashboardPage`.
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/library/adapters/http.go` — HTTP handlers: `GetDashboardPage`, `GetAlbumsTable`, `GetAlbumsPage`, `GetCarousel`, `GetAlbumDetailPage`.

### Review / rating modal
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/review/adapters/http.go` — `SubmitRatingRecommenderRating` and `DeleteRatingLogEntry` both render OOB swaps: `adapters.AlbumRating(*album, true)` (targets `album-rating-{albumId}`) and `adapters.AlbumRatingHistory(*album, true)`. Both handlers use the library adapter's `AlbumRating` component directly.
- `AlbumRating` in `dashboard.templ` — the existing combined component renders `{rating} - {label}` as a single badge (rated) or a "Rate" button (unrated), both keyed by DOM id `album-rating-{albumId}`. In the new list view this component must be replaced by a new component that shows only the numeric value and still carries the same DOM id so existing OOB swaps continue to work.

### Tags
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/tags/adapters/tags.templ` — `TagsForm` uses Alpine.js for chip input and autocomplete. The multi-select searchable pattern here is a reference for the artist filter chip's modal UI.

### Routing
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/server/server.go` — registered routes:
  - `GET /app/library/dashboard` → `GetDashboardPage`
  - `GET /app/library/dashboard/albums-table` → `GetAlbumsTable` (full list reset)
  - `GET /app/library/dashboard/albums-page` → `GetAlbumsPage` (infinite scroll, appends rows)
  - `GET /app/library/albums/{albumId}` → `GetAlbumDetailPage`

### DB / schema
- `/Users/shmoopy/Documents/code/repos/wax/db/schema.sql` — `releases.format` is a SQLite CHECK enum: `digital`, `vinyl`, `cd`, `cassette`. Ratings live in `album_rating_log`; the current rating is the most recent entry. No server-side filtering queries exist — all filtering is in-memory after fetching the full library.

### Templates infrastructure
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/core/templates/modal.templ` — `Modal` / `ForceCloseModal` pattern using `<dialog>`. The modal container lives in the page body (`ModalContainer`); filter chips will reuse this same centered modal pattern rather than a custom bottom sheet.
- `/Users/shmoopy/Documents/code/repos/wax/src/internal/core/templates/root.templ` — page shell includes HTMX, Idiomorph (morph extension), Alpine.js, and response-targets extension.

---

## Architecture

### Data flow (current)

1. `GetDashboardPage` calls `GetLibrary` → fetches all user albums in one bulk pass (releases, artists, tracks, ratings, tags, last-played).
2. Library is sorted in-memory (`SortByDate(false)` default), then paginated: only `Page(0)` (20 items) is passed to the template.
3. Infinite scroll: the last row in `albumsTableBody` renders a sentinel element with `hx-trigger="revealed"` pointing to `/albums-page?offset=N&sortBy=X&dir=Y`. On trigger, the handler re-fetches the full library, re-sorts, and slices the correct page.
4. Sort changes hit `GetAlbumsTable`, which re-fetches, re-sorts, and returns the entire list component (replacing `#album-table`).

### Data flow (v1.1)

Same as current, with these additions:
- Chip selections are encoded as URL params (e.g. `?sortBy=rating&dir=desc&minRating=7&maxRating=10&format=vinyl&artist=<id1>&artist=<id2>&rated=only`).
- `GetAlbumsTable` (renamed `GetAlbumsList` or extended in place) parses those params, applies filtering after sort, and returns the full first page of the filtered set in the new `AlbumsList` component.
- `GetAlbumsPage` receives the same filter params forwarded by the infinite scroll sentinel, re-applies the same filter pipeline, and returns the next page.
- Filter order: sort → filter → paginate (all in-memory on `AlbumDTOs`).

### Key observation: in-memory everything

There are no SQL-level filter or sort queries. The full library is always loaded from the DB on every request and sorted/filtered/paginated in Go. Filtering adds another pass on top of sorting. Acceptable for personal-scale libraries; no caching layer is needed for v1.1.

### HTMX patterns in use

- `hx-swap="outerHTML"` — replaces a whole component (list, carousel).
- `hx-swap-oob="true"` — OOB fragments for rating and tags cells after modal submission.
- `hx-trigger="revealed"` — infinite scroll sentinel row.
- `hx-ext="morph"` — idiomorph-based morphing for feeds dropdown.
- Alpine.js — used in tags form (chip/autocomplete); will also drive the open/close state for filter chip modals.

---

## Existing Patterns

### Component structure
Templates are Templ components in `adapters/`. Each visual unit is its own `templ` function, composed into larger page components. Small components that need OOB-swap capability accept an `isOobSwap bool` parameter and conditionally render `hx-swap-oob="true"` (see `AlbumRating`, `AlbumTagsCell`).

### Row / link structure
The new list view uses native DaisyUI list structure. Individual interactive-free containers within a row (album art block, title/artist block) are wrapped in `<a href="...">` links to the detail page. The rating number has its own HTMX action and is not wrapped in a link. This avoids the nested-interactive-in-`<a>` validity issue entirely while meeting the UX goal.

### Sort and filter state in URL params
Sort state is already encoded as `?sortBy=X&dir=Y` and threaded through all pagination requests. Filter params follow the same pattern, using repeated params for multi-value fields (e.g. `artist=id1&artist=id2` — idiomatic Go via `r.URL.Query()["artist"]`). The infinite scroll sentinel builds its URL with `fmt.Sprintf` — this must be extended to include all active filter params so pagination remains consistent with the active filter set.

### Modal for chip settings
Clicking a chip opens the existing centered `Modal` component (via `modal.templ`). Each chip type gets its own modal content. No new bottom sheet primitive is needed for v1.1; the existing `<dialog>` behavior and Alpine.js open/close state are sufficient and can be iterated on later.

### Default sort
Date added descending is the default. No chip is pre-selected to represent it; chips reflect explicit user selections only.

### Page size constant
`library.AlbumsPageSize = 20` controls pagination. Both the page handler and the infinite-scroll sentinel use this constant.

### DaisyUI / Tailwind
The app uses DaisyUI component classes. The new list row uses DaisyUI list primitives; filter chips use `btn` / `badge` styles; chip modals use the existing `modal` classes; rating min/max inputs use `input input-sm`; artist multi-select uses the Alpine.js autocomplete pattern from `TagsForm`.

### New rating display components
`AlbumRating` (the existing combined component) is used in the dashboard and — critically — as an OOB swap target in both `SubmitRatingRecommenderRating` and `DeleteRatingLogEntry` in the review handler. For the new list view, a new component will be created that shows only the large numeric value and carries the same stable DOM id (`album-rating-{albumId}`). The existing `AlbumRating` component continues to be used on the album detail page unchanged. The review handler's OOB swap will need to be updated to call the new list-view component when rendering the dashboard context, or the new component must be shaped so the shared id approach still works. The simplest path: the new list-view rating component uses the same id, and the review handler's OOB swap is updated to call the new component for that slot.

### Endpoint: extend in place
`GetAlbumsTable` and `GetAlbumsPage` are extended in place to parse filter params and pass them through the filter pipeline. The `AlbumsTable` component is replaced by `AlbumsList` (new component, new DOM id e.g. `#album-list`). The route for `albums-table` remains the same; the handler is renamed if needed. No need to preserve the old table behavior.

### Error handling
HTTP handlers use `httpx.HandleErrorResponse()` for non-template error paths.

---

## Constraints & Risks

### 1. Full library reload on every request
Every sort, filter, or pagination event re-fetches the entire library from the DB. Filtering adds another in-memory pass on top of sorting. Acceptable for v1.1 scope.

### 2. Infinite scroll + filtering interaction
Every active filter param must be forwarded to `/albums-page` to keep the paginated set consistent. The sentinel element builds its URL via `fmt.Sprintf` — this must include all filter params. Missing any param will cause the paginated set to diverge from the filtered view.

### 3. AlbumDTO already has everything needed
`AlbumDTO` carries `Rating`, `Tags`, `Releases` (with `Format`), `LastPlayedAt`, and `Artists` — all data needed for v1.1 filters. No new DB queries required.

### 4. Rating display ID for OOB swap — split component risk
The new list view's rating number element must carry the DOM id `album-rating-{albumId}` (via `GetAlbumRatingID`). The review modal's OOB swap in `SubmitRatingRecommenderRating` and `DeleteRatingLogEntry` calls `adapters.AlbumRating(*album, true)` — the existing combined badge. After switching to the new list view component, these OOB swap calls must be updated to render the new numeric component instead. If they still render the old combined badge into the new list row, the rating display will break visually. This is the highest-risk coupling point.

### 5. Spotify outlinks removal
The current `albumRow` has two Spotify links: (a) an external link icon next to the album title, and (b) artist names linked to Spotify artist pages. Both are removed in the new list view. The carousel's conditional Spotify outlink for out-of-library items is unaffected.

### 6. Tags ellipsis menu removed
The current `albumRow` has a `dropdown` ellipsis (⋯) menu that contains the tags action and mobile-only format/date info. This entire dropdown is removed in v1.1. Tags are shown as read-only badge(s) in the row footer. No tags action on the row — may be added back in a future pass.

### 7. AlbumDTOs sort methods mutate in place
All `SortBy*` methods call `sort.Slice` on the slice, mutating it. The filter pass must be applied after sorting (order doesn't matter semantically), and must come before pagination.

### 8. Artist filter: multi-select via repeated params
The artist chip's modal uses a multi-select searchable input, following the Alpine.js autocomplete pattern in `TagsForm`. `Library.Artists` provides the full distinct artist list. The URL param for artist filter carries multiple values (`artist=<id1>&artist=<id2>`) — the handler must collect all values with `r.URL.Query()["artist"]`.

### 9. Rating range filter: float comparison
Ratings are `float64`. Min rating filter: `album.Rating != nil && album.Rating.Rating != nil && *album.Rating.Rating >= min`. Unrated-only filter: `album.Rating == nil || album.Rating.Rating == nil`. Rated-only filter: inverse of unrated.

---

## Open Questions

No open questions remain. All previously identified questions were resolved via inline feedback:

1. **Filter param encoding**: use repeated params (`artist=id1&artist=id2`) — idiomatic Go via `r.URL.Query()["artist"]`.
2. **Modal for chip settings**: reuse the existing centered modal behavior; iterate on UI (e.g. bottom sheet on mobile) in a later pass.
3. **Tags ellipsis menu**: removed entirely for v1.1; no tags action on the row.
4. **AlbumsTable endpoint**: extend in place, rename if needed; no need to preserve existing table behavior.

---

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research  -->
