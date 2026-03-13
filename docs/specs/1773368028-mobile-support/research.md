# Mobile Support — Research

## Summary

Mobile support aims to make Wax fully usable on phones and tablets. The current UI is desktop-first — the library table renders all albums at once (a known performance issue on mobile), the header bar is dense, and interactions like dropdowns and hover tooltips don't translate to touch. The goal is responsive layouts and interactions across the full app, with lazy-loading on the dashboard as the critical performance fix.

## Relevant Code

### Layout & Shell
- `/src/internal/core/templates/root.templ` — HTML `<head>`, sets `viewport` meta (already present: `width=device-width, initial-scale=1`), loads Tailwind, HTMX, Alpine, idiomorph, Font Awesome
- `/src/internal/core/templates/layout.templ` — minimal shell used by the login/pre-app pages; does not use `NavBarComponent`
- `/src/internal/core/templates/navbar.templ` — `NavBarComponent` with `navbar-start/center/end` structure; nav links are commented out; logout button sits in `navbar-end`

### Dashboard Page (main focus)
- `/src/internal/library/adapters/dashboard.templ` — all dashboard templates:
  - `DashboardPage` — top-level page; `h-screen w-full flex flex-col overflow-hidden`
  - `DashboardHeaderBar` — fixed 44px header; wax logo, home icon, feeds dropdown, user dropdown
  - `LibraryStats` — horizontal stats strip (Artists / Albums / Tracks)
  - `CarouselSection` / `carouselStrip` — horizontally scrolling album carousel
  - `AlbumsTable` — the main library table; 8 columns (Album, Artists, Rating, Formats, Date Added, Last Played, Tags, actions menu); loads all albums at once, no pagination or lazy load
  - `albumRow` — individual row; uses tooltips (`data-tip`), `dropdown-end` menus for row actions
  - `FeedsDropdownContent` — polls the server every 5–30s via `hx-trigger="every Ns"`

### Modal System
- `/src/internal/core/templates/modal.templ` — DaisyUI `<dialog>` modal injected into `#global-modal-container` at the bottom of `<body>`; no explicit sizing — inherits DaisyUI `.modal-box` defaults
- Modals used: Rating (`RatingModal`), Review Notes (`ReviewNotesModal`), Tags (`TagsModal`)

### Auth / Login
- `/src/internal/auth/components.templ` — login page; uses `NavBarComponent`

### CSS
- `/static/src/main.css` — Tailwind v4 + DaisyUI v5; custom `wax` theme; `font-brand` for headings; ticker animation; no existing responsive breakpoint overrides
- `/static/public/main.css` — compiled output (not hand-edited)

### Routing
- `/src/internal/server/server.go` — all routes; `/app/library/dashboard` is the main page; partial endpoints for table, carousel, and feeds dropdown

## Architecture

The app is a server-rendered HTML app. Every view is a Go templ template compiled to HTML. HTMX handles partial updates in-place. Alpine.js handles client-side interactivity (tags autocomplete, rating label display). There is no client-side routing and no JavaScript framework.

**Dashboard data flow:**
1. `GET /app/library/dashboard` → `libraryHandler.GetDashboardPage` → renders `DashboardPage` with all albums pre-loaded
2. Sorting: `GET /app/library/dashboard/albums-table?sortBy=X&dir=Y` → swaps `#album-table` outerHTML
3. Carousel toggle: `GET /app/library/dashboard/carousel?view=X` → swaps `#carousel-section` outerHTML
4. Feeds dropdown: polls `/app/library/dashboard/feeds-dropdown-content` on a timer

**Modal pattern:** HTMX GET triggers return a `Modal(...)` component which injects a `<dialog>` into `#global-modal-container` via `hx-swap-oob`. Closing is driven by DaisyUI dialog form submit or `ForceCloseModal` which uses `hx-on:htmx:load` to call `.close()`.

**Performance problem:** `DashboardPage` calls `AlbumsTable(props.Library.Albums, ...)` with the entire slice — no virtual scrolling, no pagination, no lazy loading. On mobile, rendering hundreds of table rows causes a large DOM and slow initial paint.

## Existing Patterns

- **Tailwind utility classes** throughout; Tailwind v4 breakpoints (`sm:`, `md:`, `lg:`) are available but not yet used
- **DaisyUI components**: `navbar`, `table`, `carousel`, `dropdown`, `stats`, `modal`, `badge`, `btn` — all have mobile-aware responsive variants available
- **HTMX partial swaps** for progressive updates; adding `hx-trigger="intersect"` or `hx-trigger="revealed"` would support lazy loading
- **Alpine.js** for client state (tags, rating label); minimal, focused usage
- **`hx-boost="true"`** on the header nav for smooth page transitions
- **No existing breakpoint usage** — all layouts are implicitly desktop-width
- **`overflow-hidden` on the body** — the dashboard uses `h-screen overflow-hidden` + inner `overflow-y-auto` scroll, which may need adjustment on mobile (especially iOS Safari's dynamic viewport height)

## Constraints & Risks

**Critical — Performance:**
- The entire library is rendered on page load with no lazy loading. On mobile, large libraries will cause slow paint and janky scroll. Lazy loading is explicitly called out in the roadmap as required for mobile.
- Approach options: (a) server-side pagination with HTMX page/offset params, (b) HTMX `hx-trigger="intersect"` sentinel at the bottom of the table body to load the next batch, (c) infinite scroll via a "load more" button.

**Layout — Albums Table:**
- An 8-column table is unusable on a 375px screen. This is the hardest part of the mobile redesign. Options range from hiding non-essential columns (Formats, Date Added) at small breakpoints, to replacing the table with a card/list layout on mobile, to a hybrid where the table persists but columns collapse responsively.
- DaisyUI's `table` does not offer a built-in responsive collapse pattern — any card/list view on mobile would require a separate `md:hidden` / `hidden md:block` block in templ.

**Touch interactions:**
- `dropdown-end` menus on row actions are triggered by `tabindex/focus` — works on desktop, but requires a tap on mobile; should be fine with DaisyUI's implementation, but needs verification.
- Hover tooltips (`data-tip`) are invisible on touch devices. They're used for format icons and feed status. On mobile these would silently lose their labels.
- The carousel (`carousel carousel-end`) uses CSS scroll snapping — should work on touch natively, but `overscroll-x-none` on the strip may interfere with body scroll.

**Viewport height — iOS Safari:**
- `h-screen` uses `100vh` which does not account for the iOS Safari address bar. The dashboard's `h-screen flex flex-col overflow-hidden` layout could be cut off or cause double scrollbars. Tailwind v4 exposes `dvh` (`h-dvh`) which fixes this.

**Modal sizing:**
- DaisyUI `modal-box` defaults to a max-width that fits desktop. On small screens the modal may fill the screen poorly. DaisyUI supports `modal-bottom` for a sheet-style modal on mobile.

**Feeds dropdown polling:**
- The `hx-trigger="every 5s"` / `every 30s"` polling on the dropdown will fire even when the dropdown is closed, since the element is in the DOM. On mobile networks this is wasteful. Should be conditioned on visibility or moved to a different trigger mechanism.

**Font Awesome CDN:**
- Icons are loaded from `https://kit.fontawesome.com/03c1599055.js` in `root.templ:22`. This is a third-party CDN call on every page load. Investigation shows Font Awesome classes are not used anywhere in the codebase — the script tag is dead weight and should be removed.

## Decisions

1. **Mobile experience scope:** Full feature parity. Long-term direction is mobile-first with desktop as a power-user mode — this spec is the first step toward that.

2. **Table layout:** Keep the table. On mobile, collapse to three columns: Album, Artist, Rating — plus a three-dot actions menu at the end to access hidden functionality (Formats, Date Added, Last Played, Tags).

3. **Lazy loading strategy:** Infinite scroll using HTMX's `hx-trigger="revealed"` sentinel pattern (see https://htmx.org/examples/infinite-scroll/).

4. **Initial batch size:** 20 albums per batch. Defined as a named constant so it's easy to adjust.

5. **Modal behavior:** Keep centered modal on mobile. Bottom-sheet (`modal-bottom`) noted as a future enhancement.

6. **Scroll architecture:** Use natural document scroll on mobile — remove the `h-screen overflow-hidden` outer container at small breakpoints, let the page scroll normally, and make the header sticky. This avoids iOS Safari viewport height bugs and is more mobile-native. Desktop retains the current fixed inner-scroll layout.

7. **Tooltips on touch:** Drop format icon tooltips on mobile for now. Inline labels noted as a future enhancement.

8. **PWA:** Out of scope for this spec. Noted as a future enhancement.
