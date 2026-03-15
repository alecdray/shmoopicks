# My Library v1.1 — Implementation Plan

## Approach

Extend the existing dashboard handlers and templates in place. The core library data layer requires no changes — `AlbumDTO` already carries all the data needed for filtering. The work is split into three coherent areas:

1. **New list view** — replace `AlbumsTable` (table layout) with `AlbumsList` (DaisyUI list layout), a new `AlbumListRow` component leading with album art and a prominent rating number. The existing DOM id on the rating element is preserved so OOB swaps from the review handlers continue to work without a context-discrimination step.

2. **Filter + sort chip UI** — add a chip bar above the list that drives sort and filter params. Each chip opens the existing centered `Modal` component. Sort chip controls sort field and direction. Rating chip controls min/max rating and rated/unrated/only-rated. Format chip is a single-select. Artist chip is a multi-select autocomplete following the `TagsForm` Alpine.js pattern.

3. **Dashboard UX cleanup** — remove Spotify outlinks from album rows; remove the tags ellipsis menu from rows. Tags remain read-only badges in the row footer.

The approach deliberately avoids creating new routes or a new endpoint. `GetAlbumsTable` and `GetAlbumsPage` are extended in place to parse filter params and apply an in-memory filter pass after sorting, before pagination. The sentinel URL builder in the template is extended to forward all active filter params.

Alternatives considered:
- SQL-level filtering: rejected, adds complexity and migration risk for a personal-scale library where full in-memory load is already the pattern.
- New endpoint for filtered list: rejected, unnecessary — the same handler can serve both filtered and unfiltered via URL params, matching the existing sort pattern.

---

## Files to Change

| File | Change |
|------|--------|
| `src/internal/library/adapters/dashboard.templ` | Replace `AlbumsTable`, `albumRow`, `albumsTableBody`, `albumsTableSortableHeader` with `AlbumsList`, `albumListRow`, `albumsListBody`. Add `AlbumListRating` component (numeric-only, same DOM id). Add `filterChipBar` and per-chip modal content components. Remove Spotify outlinks and tags ellipsis from row. |
| `src/internal/library/adapters/http.go` | Extend `GetAlbumsTable` and `GetAlbumsPage` to parse filter params (`minRating`, `maxRating`, `rated`, `format`, `artist`) and apply `FilterAlbums` helper. Update `GetDashboardPage` to pass artist list and default filter state to the template. |
| `src/internal/library/service.go` | Add `FilterParams` struct and `AlbumDTOs.Filter(FilterParams)` method for the in-memory filter pipeline. |
| `src/internal/review/adapters/http.go` | In `SubmitRatingRecommenderRating` and `DeleteRatingLogEntry`, replace `adapters.AlbumRating(*album, true)` OOB swap call with `adapters.AlbumListRating(*album, true)` — the new numeric-only component for the list view. |

---

## Implementation Steps

1. **Add `FilterParams` and `AlbumDTOs.Filter` to `service.go`**
   - Define `FilterParams` struct: `MinRating *float64`, `MaxRating *float64`, `Rated string` (`"only"` | `"unrated"` | `""`), `Formats []models.ReleaseFormat`, `ArtistIDs []string`.
   - Add `Filter(p FilterParams) AlbumDTOs` method that applies each active filter as an in-memory pass and returns the matching subset. Unset/zero fields are treated as no filter (pass-through).
   - Run `task build/sqlc` is not required (no SQL changes).

2. **Extend `GetAlbumsTable` and `GetAlbumsPage` in `http.go`**
   - Parse filter params from `r.URL.Query()`: `minRating`, `maxRating`, `rated`, `format` (single), `artist` (repeated via `r.URL.Query()["artist"]`).
   - After the sort switch, call `albums = albums.Filter(filterParams)` before `.Page(offset)`.
   - Change the rendered component from `AlbumsTable` to `AlbumsList`.
   - Pass `filterParams` and `lib.Artists` (for artist chip modal) to the component so the chip bar can display current state.

3. **Update `GetDashboardPage` in `http.go`**
   - Default filter is empty (no active filters). Pass `lib.Artists` and empty `FilterParams` to `DashboardPage` props.
   - Update `DashboardPageProps` to include `Artists []library.ArtistDTO` and `FilterParams library.FilterParams`.

4. **Create `AlbumListRating` templ component in `dashboard.templ`**
   - Renders a large numeric rating value (rated) or a "Rate" ghost button (unrated), carrying `id={ GetAlbumRatingID(album.ID) }` and the same HTMX click action as `AlbumRating`.
   - Accepts `isOobSwap bool` and conditionally adds `hx-swap-oob="true"`.
   - This is the component that will appear in the list row.
   - The existing `AlbumRating` component is left untouched (still used on the detail page).

5. **Create `albumListRow` and `albumsListBody` templ components**
   - `albumListRow(album library.AlbumDTO)` — DaisyUI `list-row` structure:
     - Left: `<a href="…">` wrapping avatar (album art) — links to detail page.
     - Middle: `<a href="…">` wrapping title + artist names as plain text (no Spotify links).
     - Right: `@AlbumListRating(album, false)` (outside any `<a>`).
     - Footer: read-only tag badges from `album.Tags`, format badge(s) from `album.Releases`.
     - No ellipsis menu. No Spotify outlinks anywhere.
   - `albumsListBody(albums []library.AlbumDTO, offset int, sortBy, sortDir string, fp library.FilterParams)` — renders rows; if `len(albums) == library.AlbumsPageSize`, appends an infinite scroll sentinel `<li>` with `hx-get` pointing to `/app/library/dashboard/albums-page` with all active sort + filter params URL-encoded.

6. **Create `filterChipBar` and chip modal components**
   - `filterChipBar(sortBy, sortDir string, fp library.FilterParams, artists []library.ArtistDTO)` — renders a horizontal scrollable row of chips:
     - Sort chip: always visible; shows current sort label + direction indicator. Opens sort modal.
     - Rating chip: visible when `fp.MinRating`, `fp.MaxRating`, or `fp.Rated` is set; otherwise shows as inactive. Opens rating modal.
     - Format chip: visible when `fp.Formats` is non-empty; otherwise inactive. Opens format modal.
     - Artist chip: visible when `fp.ArtistIDs` is non-empty; otherwise inactive. Opens artist modal.
   - Each chip is a `<button>` styled as `btn btn-sm` (active) or `btn btn-sm btn-ghost btn-outline` (inactive) that triggers opening the chip's modal.
   - Chip modals use `hx-get` to load modal content into the existing `ModalContainer` via the `Modal` pattern, or render inline via Alpine.js `x-data` open/close if simpler. The filter form inside each modal posts to `GET /app/library/dashboard/albums-table` (replacing `#album-list`) with the appropriate params.
   - Sort modal: radio-style buttons for sort field, toggle for asc/desc.
   - Rating modal: two `input[type=number] input-sm` for min/max (0–10), radio for rated/unrated/all.
   - Format modal: radio for `digital` / `vinyl` / `cd` / `cassette` / (all).
   - Artist modal: searchable multi-select using Alpine.js autocomplete pattern from `TagsForm`, listing `Library.Artists`.

7. **Create `AlbumsList` templ component**
   - Replaces `AlbumsTable`. DOM id: `album-list` (instead of `album-table`).
   - Renders `filterChipBar` above the list.
   - Renders a `<ul class="list">` (DaisyUI list) containing `@albumsListBody(...)`.

8. **Update `DashboardPage` template**
   - Replace `@AlbumsTable(props.FirstPageAlbums, "date", "desc")` with `@AlbumsList(props.FirstPageAlbums, "date", "desc", props.FilterParams, props.Artists)`.

9. **Update review OOB swap calls**
   - In `SubmitRatingRecommenderRating`: replace `adapters.AlbumRating(*album, true)` with `adapters.AlbumListRating(*album, true)`.
   - In `DeleteRatingLogEntry`: same replacement.

10. **Run `task build/templ` after all `.templ` edits.**

---

## Database Changes

None. All filtering is in-memory. No new SQL queries or migrations are required.

---

## Feature Specs

```gherkin
Feature: My Library v1.1

  The dashboard displays the user's library as a visual list view with album art and rating,
  with chip-based filters and sort, and without Spotify outlinks or the tags ellipsis menu.

  Scenario: Library displays as visual list with album art
    Given the user has albums in their library
    When they view the dashboard
    Then each album is shown as a list row with album art, title, artist, and rating

  Scenario: Default sort is date added descending
    Given the user has albums in their library
    When they view the dashboard without any sort params
    Then albums are ordered newest-added first

  Scenario: Sort by rating descending via sort chip
    Given the user has rated and unrated albums
    When they click the sort chip and select "Rating" descending
    Then the album list reloads with highest-rated albums first and unrated albums last

  Scenario: Filter by format via format chip
    Given the user has albums in multiple formats (vinyl, digital, CD)
    When they click the format chip and select "vinyl"
    Then only albums with a vinyl release are shown in the list

  Scenario: Filter by rating range via rating chip
    Given the user has albums with ratings between 1 and 10
    When they click the rating chip and set min rating to 7
    Then only albums rated 7 or above are shown

  Scenario: Filter to unrated albums via rating chip
    Given the user has rated and unrated albums
    When they click the rating chip and select "Unrated only"
    Then only albums without a rating are shown

  Scenario: Filter by artist via artist chip
    Given the user has albums from multiple artists
    When they click the artist chip and select one or more artists
    Then only albums by those artists are shown

  Scenario: Infinite scroll loads next page respecting active filters
    Given the user has filtered to vinyl albums and there are more than 20
    When they scroll to the bottom of the list
    Then the next page of vinyl albums is appended, maintaining the format filter

  Scenario: Rating OOB swap updates the list row after rating submission
    Given the user opens the rating modal for an album
    When they submit a rating
    Then the album row's rating display updates in place without a full page reload

  Scenario: No Spotify outlinks appear in album rows
    Given the user views the library list
    Then no external Spotify links are visible for album titles or artist names

  Scenario: Tags ellipsis menu is not present on album rows
    Given the user views the library list
    Then there is no ellipsis (⋯) dropdown menu on any album row
    And tags are displayed as read-only badges in the row
```

---

## Testing

**Unit tests (Go):**
- `AlbumDTOs.Filter` in `service.go`: table-driven tests covering each filter type in isolation and in combination — min/max rating, rated-only, unrated-only, format, single artist, multi-artist, no filters (passthrough).

**E2E specs (Playwright, `e2e/feat/`):**
- Happy path: dashboard loads with list view, album art visible, no Spotify outlinks, no ellipsis menu.
- Sort chip: clicking sort chip, selecting a field, verifying order.
- Format chip: selecting vinyl, confirming only vinyl albums remain.
- Rating chip: setting min rating, confirming filtered results.
- Artist chip: selecting an artist, confirming filtered results.
- Infinite scroll with active filter: verifying next page preserves filter.
- OOB rating swap: submit rating via modal, verify row rating updates without reload.

---

## Risks & Mitigations

**OOB rating swap coupling (highest risk)**
`SubmitRatingRecommenderRating` and `DeleteRatingLogEntry` currently call `adapters.AlbumRating(*album, true)`. This must be changed to `adapters.AlbumListRating(*album, true)`. If missed, the rating element in list rows will be replaced with the old combined badge, breaking the visual. Mitigation: step 9 is explicit; the E2E spec for OOB rating swap will catch a regression.

**Sentinel URL builder missing filter params**
If `albumsListBody` omits any active filter param from the sentinel's `hx-get` URL, paginated results will diverge from the filtered view. Mitigation: `FilterParams` is passed as a struct to the template; the sentinel builds its URL from the same struct so all params are forwarded consistently.

**Alpine.js artist multi-select complexity**
The artist chip modal needs an autocomplete multi-select. The `TagsForm` pattern is the reference, but it's wired to a different form submission mechanism. Mitigation: keep the artist modal form simple — a `GET` form targeting `#album-list`. Alpine.js state tracks selected artist IDs and serializes them as repeated `artist=<id>` hidden inputs on submit.

**Filter interaction with empty result set**
When filters produce zero results, `albumsListBody` must show a meaningful empty state rather than a blank list. Mitigation: add an explicit empty-state message branch in `albumsListBody` that reads "No albums match your filters."

**DOM id rename (`album-table` → `album-list`)**
The `hx-target="#album-table"` in the old sortable headers is being replaced entirely (sort moves to a chip). However any other template or JS that targets `#album-table` must be updated. The sortable headers are fully removed; a codebase search confirms `#album-table` appears only in `dashboard.templ`. Mitigation: search for `album-table` across all templates before marking step complete.

---

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan docs/wiki/pages/my-library-v1.1.md -->
