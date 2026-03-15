# My Library v1.1 — Implementation

## What Was Built

All scoped features from the plan were delivered:

1. **Visual list view** — `AlbumsTable` (table layout) replaced by `AlbumsList` (DaisyUI `list-row` layout, `max-w-3xl`). Each row has:
   - A vertical column of all 4 format icons (Digital, Vinyl, CD, Cassette) — full opacity if owned, `opacity-15` if not
   - Album art (`size-24`, links to detail page)
   - Title + artist text block (links to detail page), with a footer row showing rating label badge and tag badges (`text-nowrap`)
   - A large numeric rating (`text-4xl`, `text-primary`, `min-w-14`, full row height for easy clicking) — shows `--` if unrated
   - No Spotify outlinks. No tags ellipsis menu.

2. **AlbumListRating** — renders a large numeric rating (`%.1f` format) or `--` for unrated, both styled as clickable. Carries the same DOM id (`album-rating-{albumId}`) so OOB swaps from review handlers continue to work.

3. **Chip-based filter + sort UI** — `filterChipBar` renders four chips above the list:
   - Sort chip: always active, shows current sort field + direction. Opens a dialog with radio sort fields and asc/desc toggle.
   - Rating chip: active when min/max rating or rated filter is set. Opens a dialog with number inputs and rated/unrated/all radio.
   - Format chip: active when a format filter is set. Opens a dialog with format radio (single-select).
   - Artist chip: active when artist filter is set. Opens a dialog with a searchable checkbox list using Alpine.js.
   - All chip modals submit via HTMX `hx-get` to `/app/library/dashboard/albums-table`, replacing `#album-list`.

4. **In-memory filtering** — `FilterParams` struct and `AlbumDTOs.Filter` method added to `service.go`. Filter applies after sort, before pagination. `GetAlbumsTable` and `GetAlbumsPage` both parse and apply filter params.

5. **Infinite scroll with filter preservation** — sentinel builds its URL via `buildAlbumsPageURL` helper which encodes all active sort + filter params.

6. **Review OOB swap updated** — `SubmitRatingRecommenderRating` and `DeleteRatingLogEntry` now call `AlbumListRating(*album, true)` instead of `AlbumRating(*album, true)`.

7. **Unit tests** — 12 new table-driven tests for `AlbumDTOs.Filter` covering all filter types and combinations.

**Known rough edge:** The chip-based filter/sort UI is functional but the UX and visual polish are not finalized. The dialogs feel clunky and the overall chip bar styling needs iteration. This is deferred to a follow-up.

## Differences from the Plan

**Row layout** — the plan described a left/middle/right structure (art → title → rating). The final layout is: format icons column → art → title block → rating. Format icons moved out of footer badges into their own leading column showing all 4 formats always (dimmed if not owned), giving a quick visual summary of library formats per album.

**Rating display** — plan said "large numeric value". Final: `text-4xl text-primary`, `%.1f` format (e.g. "7.5", "10.0"), `--` for unrated (clickable, dimmed). Entire right column is clickable (`min-w-14 h-full`).

**Format badges in footer removed** — formats are represented only by the leading icon column. Footer shows only rating label badge + tag badges.

**Artist modal search wiring** — the plan referenced the `TagsForm` Alpine.js autocomplete pattern. The actual implementation uses server-rendered checkboxes with an Alpine.js `x-data="{ search: '' }"` wrapper and `x-show="!search || $el.dataset.name.includes(search.toLowerCase())"` for client-side filtering. HTMX serializes checked checkboxes as repeated `artist=<id>` params naturally.

**Chip modals use `<dialog>` opened via Alpine.js** rather than the existing `Modal` / `ModalContainer` templ components. Reason: `Modal` requires an HTMX request to load content into `ModalContainer`, which would require new endpoints. Using `<dialog x-ref="...">` toggled via `@click="$refs.X.showModal()"` is simpler and works inline.

**`AlbumTagsCell` not used in list rows** — tags are rendered inline in the row footer to avoid the OOB-swap id that `AlbumTagsCell` carries.

**`releaseFormatBadge` removed** — was unused after the format display was redesigned.

## Plan Inaccuracies

- **`AlbumDTOs.Filter` passthrough optimization**: the plan did not mention an early-return when no filters are set. Added to avoid unnecessary allocation on unfiltered requests.
- **Artist chip modal `x-data` scope**: the plan's reference to `TagsForm` pattern was not directly applicable. The inline scoped `x-data` pattern is cleaner for a self-contained dialog.
- **`Formats []models.ReleaseFormat`**: the plan defines it as a slice for forward compatibility, but the format chip is single-select. In practice only 0 or 1 value is set.
