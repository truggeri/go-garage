# Web Interface Documentation

This document describes the Go-Garage web interface: its template system, CSS classes, JavaScript modules, and UI patterns. Use it as a reference when building or modifying pages.

## Table of Contents

- [Template System](#template-system)
  - [Directory Structure](#directory-structure)
  - [Layouts](#layouts)
  - [Partials](#partials)
  - [Page Templates](#page-templates)
  - [Error Pages](#error-pages)
  - [Rendering Templates](#rendering-templates)
  - [Page Data Structs](#page-data-structs)
- [CSS Reference](#css-reference)
  - [Design Tokens](#design-tokens)
  - [Layout](#layout)
  - [Typography](#typography)
  - [Buttons](#buttons)
  - [Forms](#forms)
  - [Cards](#cards)
  - [Tables](#tables)
  - [Flash Messages](#flash-messages)
  - [Navigation](#navigation)
  - [Dashboard](#dashboard)
  - [Vehicle Components](#vehicle-components)
  - [Detail Pages](#detail-pages)
  - [Pagination](#pagination)
  - [Empty States](#empty-states)
  - [Utility Classes](#utility-classes)
  - [Accessibility](#accessibility)
  - [Responsive Breakpoints](#responsive-breakpoints)
  - [Dark Mode](#dark-mode)
- [JavaScript Modules](#javascript-modules)
  - [main.js](#mainjs)
  - [form-validation.js](#form-validationjs)
- [Style Guide](#style-guide)
  - [Naming Conventions](#naming-conventions)
  - [Template Patterns](#template-patterns)
  - [Form Patterns](#form-patterns)
  - [Accessibility Patterns](#accessibility-patterns)

---

## Template System

### Directory Structure

```text
web/templates/
├── layouts/
│   ├── base.html          # Main layout (navbar + footer)
│   └── auth.html          # Auth layout (centered card, no navbar)
├── partials/
│   ├── navigation.html    # Top navigation bar
│   ├── flash-messages.html# Flash message alerts
│   ├── header.html        # Page header partial
│   └── footer.html        # Site footer
├── pages/
│   ├── home.html          # Public landing page
│   ├── dashboard.html     # Authenticated dashboard
│   ├── login.html         # Login form
│   ├── register.html      # Registration form
│   ├── vehicles/
│   │   ├── list.html      # Vehicle list with cards
│   │   ├── new.html       # Add vehicle form
│   │   ├── detail.html    # Vehicle detail view
│   │   └── edit.html      # Edit vehicle form
│   ├── maintenance/
│   │   ├── list.html      # Maintenance records table
│   │   ├── new.html       # Add maintenance form
│   │   ├── detail.html    # Maintenance detail view
│   │   └── edit.html      # Edit maintenance form
│   └── profile/
│       ├── view.html      # View profile
│       ├── edit.html      # Edit profile form
│       └── password.html  # Change password form
└── errors/
    ├── 403.html           # Forbidden
    ├── 404.html           # Not found
    └── 500.html           # Server error
```

### Layouts

#### `base.html`

The main application layout. Includes the navigation bar, flash messages, a `<main>` content area, and the site footer. Used for all authenticated pages and the public home page.

**Blocks available:**

| Block | Purpose | Default |
|-------|---------|---------|
| `title` | Page `<title>` | `Go-Garage` |
| `extra_head` | Additional `<head>` content | _(empty)_ |
| `content` | Main page content | _(empty)_ |
| `extra_scripts` | Additional `<script>` tags | _(empty)_ |

**Usage in a page template:**

```html
{{define "title"}}My Page - Go-Garage{{end}}

{{define "content"}}
<h1>Page Heading</h1>
<p>Content goes here.</p>
{{end}}
```

#### `auth.html`

A centered, minimal layout for login and registration pages. Displays a brand link, flash messages, and content inside an `auth-container`. No navigation bar.

Provides the same blocks as `base.html`.

### Partials

| Partial | Description |
|---------|-------------|
| `navigation` | Top navbar with brand, links (Dashboard, Vehicles, Maintenance), user dropdown (Profile, Logout), theme toggle, and mobile hamburger menu. Uses `{{.IsAuthenticated}}`, `{{.ActiveNav}}`, and `{{.UserName}}` from page data. |
| `flash-messages` | Iterates over `{{.Flash}}` to render alert banners. Each flash has a `.Type` (`success`, `error`, `warning`, `info`) and `.Message`. |
| `footer` | Simple site footer with copyright. |
| `header` | Optional page header partial. |

### Page Templates

Each page template defines a `title` block and a `content` block. The handler chooses which layout to use when rendering (see [Rendering Templates](#rendering-templates)).

### Error Pages

Error pages (`403.html`, `404.html`, `500.html`) use the `error-page` CSS class and provide links back to the dashboard or home page.

### Rendering Templates

Templates are rendered through the **template engine** in `internal/templateengine/engine.go`.

```go
// Signature
func (e *Engine) Render(w io.Writer, name, layout string, data interface{}) error
```

**Parameters:**

| Parameter | Description | Examples |
|-----------|-------------|----------|
| `name` | Template path relative to `pages/` | `"dashboard.html"`, `"vehicles/list.html"` |
| `layout` | Layout block name | `"base"` or `"auth"` |
| `data` | Page data struct | `dashboardPageData{...}` |

**Example handler call:**

```go
h.engine.Render(w, "vehicles/list.html", "base", data)
```

The engine combines all layouts, partials, and the named page template, then executes the specified layout block with the provided data struct.

### Page Data Structs

Every page handler defines a data struct that is passed to the template. All authenticated page structs share these common fields:

| Field | Type | Description |
|-------|------|-------------|
| `Flash` | `interface{}` | Flash messages for the `flash-messages` partial |
| `IsAuthenticated` | `bool` | Whether the user is logged in |
| `UserName` | `string` | Display name for the navbar |
| `ActiveNav` | `string` | Highlights the active nav link (`"dashboard"`, `"vehicles"`, `"maintenance"`, `"profile"`) |

Form pages add:

| Field | Type | Description |
|-------|------|-------------|
| `CSRFToken` | `string` | CSRF token for the hidden form field |
| `Errors` | `map[string]string` | Field-specific validation error messages (keyed by field name) |

Page data structs are defined in the corresponding `internal/handlers/page_*.go` files.

---

## CSS Reference

All styles are in `web/static/css/main.css` (served minified as `main-minified.css`). The stylesheet is organized into clearly labeled sections.

### Design Tokens

CSS custom properties defined on `:root`. Use these variables instead of hard-coded values.

**Colors:**

| Variable | Value | Usage |
|----------|-------|-------|
| `--color-primary` | `#2563eb` | Primary actions, links |
| `--color-primary-hover` | `#1d4ed8` | Primary hover state |
| `--color-secondary` | `#64748b` | Secondary elements |
| `--color-success` | `#16a34a` | Success states |
| `--color-danger` | `#dc2626` | Errors, delete actions |
| `--color-warning` | `#d97706` | Warning alerts |
| `--color-info` | `#0891b2` | Info alerts |

**Typography:**

| Variable | Value |
|----------|-------|
| `--font-size-sm` | `0.875rem` |
| `--font-size-base` | `1rem` |
| `--font-size-lg` | `1.125rem` |
| `--font-size-xl` | `1.25rem` |
| `--font-size-2xl` | `1.5rem` |
| `--font-size-3xl` | `2rem` |

**Spacing:**

| Variable | Value |
|----------|-------|
| `--spacing-xs` | `0.25rem` |
| `--spacing-sm` | `0.5rem` |
| `--spacing-md` | `1rem` |
| `--spacing-lg` | `1.5rem` |
| `--spacing-xl` | `2rem` |
| `--spacing-2xl` | `3rem` |

**Layout & Effects:**

| Variable | Value |
|----------|-------|
| `--container-max-width` | `1200px` |
| `--border-radius` | `0.375rem` |
| `--border-radius-lg` | `0.5rem` |
| `--shadow-sm` | subtle shadow |
| `--shadow-md` | medium shadow |
| `--shadow-lg` | large shadow |

### Layout

| Class | Description |
|-------|-------------|
| `.container` | Centered content container, max-width `1200px`, horizontal padding |
| `main.container` | Flex-grow main area with top/bottom padding |

### Typography

Headings (`h1`–`h6`) use `--color-gray-900` with tight line-height. Size mapping:

| Element | Size |
|---------|------|
| `h1` | `--font-size-3xl` (2rem) |
| `h2` | `--font-size-2xl` (1.5rem) |
| `h3` | `--font-size-xl` (1.25rem) |

### Buttons

| Class | Description |
|-------|-------------|
| `.btn` | Base button styles (padding, border-radius, cursor, transitions) |
| `.btn-primary` | Blue background, white text |
| `.btn-secondary` | White background, gray border |
| `.btn-danger` | Red background, white text |
| `.btn-sm` | Smaller padding and font size |
| `.btn-block` | Full-width block button |
| `.btn-loading` | Loading state — text becomes transparent, a CSS spinner appears. Applied automatically by `main.js` on form submit. |

### Forms

| Class | Description |
|-------|-------------|
| `.form-group` | Wrapper for a label + input + error message |
| `.form-label` | Styled `<label>` element |
| `.form-input` | Styled `<input>`, `<select>`, or `<textarea>` |
| `.form-input.is-invalid` | Red border for invalid fields |
| `.form-input.is-valid` | Green border for valid fields |
| `.form-error` | Red error message text below a field |
| `.form-hint` | Gray hint text below a field |
| `.form-actions` | Flex container for submit/cancel buttons |
| `.form-section` | Groups related fields with bottom margin |
| `.form-section-title` | Section heading with a bottom border |
| `.form-row` | CSS Grid row that auto-fits columns (min 200px) for side-by-side fields |
| `.required` | Red asterisk for required fields |

**Form template pattern:**

```html
<div class="form-group">
    <label class="form-label" for="make">Make <span class="required" aria-hidden="true">*</span></label>
    <input class="form-input{{if .Errors.make}} is-invalid{{end}}"
           type="text" id="make" name="make"
           value="{{.Make}}"
           required placeholder="e.g. Toyota"
           {{if .Errors.make}}aria-invalid="true" {{end}}aria-describedby="make-error">
    {{if .Errors.make}}
    <p class="form-error" id="make-error">{{.Errors.make}}</p>
    {{end}}
</div>
```

### Cards

| Class | Description |
|-------|-------------|
| `.card` | White background container with border, shadow, and padding |
| `.card-header` | Card header with bottom border |
| `.card-header--flex` | Flex header with space-between alignment |
| `.card-title` | Card heading (no bottom margin) |
| `.card-body` | Card body with top padding |
| `.card-footer` | Card footer with top border |

### Tables

| Class | Description |
|-------|-------------|
| `.table` | Full-width table with collapsed borders |
| `.table-responsive` | Responsive table — on mobile (≤768px), rows become stacked cards. Requires `data-label` attributes on `<td>` elements. |

**Responsive table pattern:**

```html
<table class="table table-responsive">
    <thead>
        <tr>
            <th>Date</th>
            <th>Type</th>
            <th>Actions</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td data-label="Date">2024-01-15</td>
            <td data-label="Type">Oil Change</td>
            <td data-label="Actions"><a href="#">View</a></td>
        </tr>
    </tbody>
</table>
```

### Flash Messages

| Class | Description |
|-------|-------------|
| `.flash-messages` | Container for flash message list |
| `.flash` | Individual flash message (flex layout) |
| `.flash-success` | Green success message |
| `.flash-error` | Red error message |
| `.flash-warning` | Yellow warning message |
| `.flash-info` | Blue info message |
| `.flash-close` | Close button inside a flash |
| `.flash-dismiss` | Applied during fade-out animation |

Flash messages auto-dismiss after 5 seconds (except errors). See `main.js` for behavior.

### Navigation

| Class | Description |
|-------|-------------|
| `.navbar` | Top navigation bar |
| `.navbar-content` | Flex container inside navbar |
| `.navbar-brand` | Logo / brand link |
| `.navbar-toggle` | Hamburger menu button (hidden on desktop) |
| `.navbar-menu` | Menu container (links + user dropdown + theme toggle) |
| `.navbar-links` | Flex row of navigation links |
| `.navbar-link` | Individual nav link |
| `.navbar-link--active` | Active/current page indicator |
| `.navbar-user-dropdown` | User dropdown wrapper |
| `.navbar-user-toggle` | Dropdown trigger button |
| `.navbar-dropdown-menu` | Dropdown menu panel |
| `.navbar-dropdown-item` | Dropdown menu link |
| `.navbar-dropdown-divider` | Horizontal rule in dropdown |

### Dashboard

| Class | Description |
|-------|-------------|
| `.dashboard-header` | Dashboard heading area |
| `.stat-grid` | CSS Grid for statistics cards (auto-fit, min 180px) |
| `.stat-card` | Individual stat card (white, centered) |
| `.stat-value` | Large number display |
| `.stat-label` | Small text label under stat value |
| `.dashboard-actions` | Flex container for quick-action buttons |

### Vehicle Components

| Class | Description |
|-------|-------------|
| `.vehicle-grid` | CSS Grid for vehicle cards (auto-fill, min 300px) |
| `.vehicle-card` | Flex column card for a vehicle |
| `.vehicle-card-header` | Header with title and status badge |
| `.vehicle-card-title` | Vehicle name heading |
| `.vehicle-status` | Status badge (pill shape) |
| `.vehicle-status-active` | Green badge |
| `.vehicle-status-sold` | Yellow badge |
| `.vehicle-status-scrapped` | Red badge |
| `.vehicle-card-body` | Card body area |
| `.vehicle-detail` | Small detail text line |
| `.vehicle-detail-label` | Bold label in detail line |
| `.vehicle-card-actions` | Action buttons row with top border |

### Detail Pages

| Class | Description |
|-------|-------------|
| `.detail-layout` | Grid layout — 1 column on mobile, 2fr + 1fr on desktop (>1024px) |
| `.detail-main` | Primary content column |
| `.detail-sidebar` | Sidebar column (stats, quick actions) |
| `.detail-list` | Grid of detail items (1 col on mobile, 2 cols on ≥640px) |
| `.detail-item` | Label + value pair |
| `.detail-label` | Small gray label text |
| `.detail-value` | Value text |
| `.detail-notes` | Pre-wrapped notes text |
| `.stat-list` | Vertical list of stat items |
| `.stat-item` | Stat label + value with bottom border |
| `.stat-item-label` | Stat item label |
| `.stat-item-value` | Stat item value (large, bold) |
| `.action-list` | Vertical list of action buttons |
| `.page-header` | Flex header with title and action buttons |
| `.page-header-actions` | Grouped action buttons in page header |

### Pagination

| Class | Description |
|-------|-------------|
| `.pagination` | Centered flex container for pagination controls |
| `.pagination-info` | Small text showing page info (e.g., "Page 1 of 5") |

### Empty States

| Class | Description |
|-------|-------------|
| `.empty-state` | Centered text container for empty lists |
| `.empty-state-icon` | Large icon/emoji above the message |
| `.empty-state-hero` | Boxed empty-state variant with background and border |

### Utility Classes

| Class | Description |
|-------|-------------|
| `.text-center` | Center-aligned text |
| `.text-muted` | Gray muted text |
| `.mt-md` | `margin-top: 1rem` |
| `.mb-md` | `margin-bottom: 1rem` |
| `.mb-lg` | `margin-bottom: 1.5rem` |
| `.sr-only` | Screen-reader only (visually hidden) |

### Accessibility

| Class / Feature | Description |
|-----------------|-------------|
| `.skip-link` | "Skip to content" link — hidden until focused (keyboard navigation) |
| `.sr-only` | Visually hidden content for screen readers |
| Focus indicators | `focus-visible` outlines on buttons, links, inputs, and dropdown items |

### Responsive Breakpoints

| Breakpoint | Target | Key Changes |
|------------|--------|-------------|
| ≤768px | Mobile | Hamburger nav menu, single-column forms, stacked table rows, touch-friendly 44px min-height inputs |
| ≤480px | Small mobile | Single-column stat grid, stacked action buttons, stacked pagination |
| ≤1024px | Tablet | Single-column detail layout, adjusted vehicle grid |

### Dark Mode

Dark mode is controlled by a `data-theme` attribute on `<html>`. The theme is detected from `localStorage` or the OS `prefers-color-scheme` media query via an inline script in the `<head>` (prevents flash of wrong theme).

Dark mode overrides all design tokens — background colors, text colors, borders, shadows, and component-specific colors. The theme toggle button in the navbar (or on auth pages) switches between light and dark modes and saves the preference to `localStorage`.

---

## JavaScript Modules

All scripts are in `web/static/js/` and served minified. Both files use an IIFE pattern to avoid global scope pollution.

### main.js

Core UI behaviors initialized on `DOMContentLoaded`:

| Function | Description |
|----------|-------------|
| `initFlashMessages()` | Adds close-button handlers to `.flash` elements. Auto-dismisses non-error flashes after 5 seconds with a fade-out transition. |
| `initNavbarToggle()` | Toggles `.navbar-menu.active` class when the hamburger button (`.navbar-toggle`) is clicked. Updates `aria-expanded` for accessibility. |
| `initUserDropdown()` | Opens/closes the user dropdown menu (`.navbar-dropdown-menu`). Supports click-to-toggle, outside-click-to-close, Escape key, and arrow-key navigation between `[role="menuitem"]` elements. |
| `initThemeToggle()` | Manages the dark/light mode toggle button (`#theme-toggle`). Reads theme from `localStorage`, toggles `data-theme` on `<html>`, and updates `aria-pressed` / `aria-label`. |
| `initFormSubmitLoading()` | Listens for `submit` events on all forms. Adds `.btn-loading` class to the submit button, sets `disabled` and `aria-busy="true"` to prevent double-submissions. |
| `initDeleteButtons()` | Handles buttons with `data-confirm-delete` attribute. Shows a `confirm()` dialog with the `data-confirm-message` text, then submits the referenced hidden form with a loading spinner. |

**Delete button pattern:**

```html
<button class="btn btn-danger"
        data-confirm-delete="delete-form"
        data-confirm-message="Are you sure you want to delete this vehicle?">
    Delete
</button>
<form id="delete-form" method="POST" action="/vehicles/123/delete" style="display:none">
    <input type="hidden" name="csrf_token" value="...">
</form>
```

### form-validation.js

Client-side form validation for forms with the `novalidate` attribute.

**Validation rules:**

| Rule | Trigger | Check |
|------|---------|-------|
| Required | `required` attribute | Field must not be empty |
| Email | `type="email"` | Must match email regex |
| Min/Max length | `minlength` / `maxlength` | Character count bounds |
| Number range | `type="number"` + `min` / `max` | Numeric bounds |
| Date | `type="date"` | Must be a parseable date |
| Password strength | `data-validate-password="true"` | Requires uppercase, lowercase, and digit; min 8 chars |
| Password match | `data-match-field="fieldId"` | Must match the referenced field's value |

**Event handling:**

- **`blur`**: Validates the field if it has a value or is already marked invalid.
- **`input`**: Re-validates if the field is currently marked invalid (real-time correction feedback).
- **`submit`**: Validates all fields; prevents submission and focuses the first invalid field if any fail.

**UI feedback:**

| State | CSS Class | ARIA |
|-------|-----------|------|
| Invalid | `.is-invalid` on input, `.form-error` message shown | `aria-invalid="true"`, error linked via `aria-describedby` |
| Valid | `.is-valid` on input, error message hidden | `aria-invalid` removed |

**Custom validation attributes:**

```html
<!-- Password with strength validation -->
<input type="password" data-validate-password="true" required minlength="8">

<!-- Password confirmation -->
<input type="password" data-match-field="new_password" required>
```

---

## Style Guide

### Naming Conventions

- **CSS classes**: Use lowercase with hyphens (`kebab-case`). Component names use BEM-like naming with `--` for modifiers (e.g., `.navbar-link--active`).
- **Template blocks**: Use lowercase with underscores (e.g., `extra_head`, `extra_scripts`).
- **Template definitions**: Use lowercase with hyphens (e.g., `{{define "flash-messages"}}`).
- **Page data fields**: Use `PascalCase` for exported fields (Go convention).
- **HTML IDs**: Use lowercase with hyphens (e.g., `id="make-error"`).
- **Data attributes**: Use lowercase with hyphens (e.g., `data-confirm-delete`, `data-label`).

### Template Patterns

1. **Every page template** defines `title` and `content` blocks.
2. **CSRF tokens** are included in all forms as a hidden input:
   ```html
   <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
   ```
3. **Server-side errors** are displayed conditionally per field:
   ```html
   {{if .Errors.fieldname}}
   <p class="form-error" id="fieldname-error">{{.Errors.fieldname}}</p>
   {{end}}
   ```
4. **Active navigation** is set via the `ActiveNav` field in the page data struct. The navigation partial highlights the matching link.
5. **Forms use `novalidate`** to delegate validation to the JavaScript validation module.

### Form Patterns

- Wrap each field in a `.form-group`.
- Use `.form-row` for side-by-side fields (auto-fits into columns).
- Group related fields in a `.form-section` with a `.form-section-title`.
- Place submit and cancel buttons in `.form-actions`.
- Mark required fields with `<span class="required" aria-hidden="true">*</span>`.
- Connect error messages to inputs with `aria-describedby`.
- Add `aria-invalid="true"` conditionally when server-side errors exist.

### Accessibility Patterns

- Use semantic HTML elements (`<nav>`, `<main>`, `<form>`, `<table>`).
- Include `aria-label` on navigation and toggle buttons.
- Use `role="menu"` and `role="menuitem"` for dropdown menus.
- Use `role="alert"` on flash message containers.
- Provide `aria-expanded` on toggle buttons (navbar, dropdown).
- Use `aria-current="page"` on the active navigation link.
- Include a `.skip-link` at the top of every layout for keyboard users.
- Ensure all interactive elements have visible focus indicators (`:focus-visible`).
- Use `aria-hidden="true"` on decorative elements (icons, required asterisks).
