# Mobile Support — Implementation

## What was built

All four areas from the plan were delivered:

1. **Font Awesome removed** — dead `<script>` tag deleted from `root.templ`.
2. **Viewport / layout fix** — body changed from `h-screen` to `min-h-dvh`. `DashboardPage` outer wrapper changed to `md:h-dvh md:overflow-hidden` (desktop retains fixed-height app layout; mobile uses natural document scroll). Inner scroll div changed to `md:flex-1 md:overflow-y-auto`.
3. **Sticky header** — `sticky top-0 z-10` added to `DashboardHeaderBar`.
4. **Column collapse** — Formats, Date Added, Last Played, and Tags columns (both `<th>` and `<td>`) hidden with `hidden md:table-cell` on mobile. The three-dot actions dropdown expanded with mobile-only (`md:hidden`) info items: format badges, date added, last played.
5. **Infinite scroll** — `AlbumsPageSize = 20` constant and `Page(offset int) AlbumDTOs` method added to `service.go`. `albumsTableBody` updated to accept `offset`, `sortBy`, `sortDir` params and render an HTMX `hx-trigger="revealed"` sentinel `<tr>` when a full page is returned. New `GET /app/library/dashboard/albums-page` endpoint added. `GetDashboardPage` and `GetAlbumsTable` both pass `Page(0)` as the initial album slice.

## Differences from the plan

- **`albumsTableSortableHeader` required a new `extraClass` parameter** — the plan described adding `hidden md:table-cell` to sortable header `<th>` elements but didn't account for the fact that the class is generated inside the template. Added `extraClass string` param and used `class={ "cursor-pointer hover:bg-base-200", extraClass }` to merge classes.
- **Pagination implemented via `Page` method on `AlbumDTOs`, not a service method** — the plan described a `GetAlbumsInLibraryPaged` service method. Instead, a `Page(offset int)` method was added to the `AlbumDTOs` type. This keeps sorting in the handler (consistent with `GetAlbumsTable`) and avoids passing sort params into the service layer.
- **Dropdown width changed from responsive to fixed `w-48`** — plan mentioned `w-48 md:w-36`. A fixed `w-48` was used since the desktop layout is unaffected at that width and it avoids an unnecessary responsive class.
- **`fmt.Sscanf` used for offset parsing** — no query param parsing helper exists in the codebase; `fmt.Sscanf` parses the integer offset with a zero default on failure.

## Future Considerations

The current implementation makes the album table usable on mobile via column collapse and horizontal scroll, but a table is fundamentally a desktop-oriented pattern. A proper mobile-first approach would likely involve a purpose-built album list or card view for small screens — either as a replacement for the table on mobile or as an alternative view mode the user can switch between. This is worth revisiting as mobile usage grows.

## Plan inaccuracies

- The plan said to change `body` to `min-h-dvh` and also mentioned `h-dvh` on the outer wrapper; both were done but the body had `h-screen` (not just in `DashboardPage` — it was on the `<body>` tag in `root.templ`). The plan correctly identified both locations.
- `albumsTableSortableHeader` had a duplicate `class` attribute in the original template (`class="cursor-pointer hover:bg-base-200"` and `class="inline-block mr-1"` — the second was vestigial and was removed during the edit).
- The plan referenced `albumsTableBody` being called from `AlbumsTable` — correct. But it didn't mention that `GetAlbumsPage` renders `albumsTableBody` directly (a private templ function), which works because both are in the same `adapters` package.
