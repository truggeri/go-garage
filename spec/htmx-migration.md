# HTMX and REST Migration Conventions

This document defines the conventions for migrating state-changing web
interface actions from full-page POST handlers to the REST API with
[htmx](https://htmx.org). It is the implementation contract for the work
tracked by issue #230.

## Scope and goals

The browser continues to use Go `html/template` for GET/list/detail pages.
Only create, update, and delete actions move to `/api/v1/...` endpoints. API
handlers remain JSON-only: they must not render templates or know about page
layout markup.

The following actions are in scope:

| Area | Actions |
| --- | --- |
| Vehicles | Create, edit, delete from the list and detail pages |
| Maintenance | Create, edit, delete from the list and detail pages |
| Fuel | Create, edit, delete from the list and detail pages |
| Profile | Edit profile |
| Password | Change password |

Login, registration, and logout remain full-page cookie flows. They are not
resource CRUD and are outside this migration.

## Authentication

The protected `/api/v1` router accepts both authentication mechanisms:

1. Prefer an `Authorization: Bearer <token>` header. This preserves the
   existing contract for programmatic API clients.
2. If no bearer header is present, accept the `access_token` and
   `refresh_token` cookies using the same validation and refresh behavior as
   protected page routes.

The authentication middleware records which mechanism succeeded in request
context. This is required by CSRF middleware: bearer-authenticated requests do
not use ambient browser credentials and therefore do not require the
cookie-specific CSRF check; cookie-authenticated state-changing requests do.

When authentication fails:

- A normal API client receives the existing JSON `401` response.
- An htmx/browser request receives the same JSON error envelope plus
  `HX-Redirect: /login`. The shared htmx module follows that redirect instead
  of inserting the JSON error into the page.

No existing bearer-token response shape or status code is changed.

Implemented by `middleware.HybridAuthGuard` in
`internal/middleware/authentication.go` and applied to the protected
`/api/v1` subrouter (PR #249).

## CSRF protection

Cookie-authenticated `POST`, `PUT`, and `DELETE` requests to `/api/v1` must include a valid `X-CSRF-Token` header. The API router must apply CSRF middleware after hybrid authentication; that middleware must validate this header for JSON requests while retaining `csrf_token` form-field validation for full-page forms. Bearer-authenticated requests skip this additional check.

After header support is implemented, authenticated layouts must expose the generated token as:

```html
<meta name="csrf-token" content="{{.CSRFToken}}">
```

The shared htmx module reads this value and adds it to every htmx request.
The existing `csrf_token` hidden form field remains the convention for
full-page forms, including login and registration.

CSRF failures from API requests use the normal JSON error envelope with HTTP
`403`; they must not render an HTML error page.

Implemented by `middleware.APICSRFProtection` in `internal/middleware/csrf.go`,
applied to the protected `/api/v1` subrouter after `HybridAuthGuard`. The shared
`CSRFProtection` middleware now accepts the token from the `X-CSRF-Token` header
in addition to the `csrf_token` form field, and `web/templates/layouts/base.html`
renders the `csrf-token` meta tag (PR #250).

## Front-end assets

htmx and its `json-enc` extension are vendored in `web/static/js/vendor/`
(`htmx.min.js` and `json-enc.js`); see that directory's `README.md` for the
upstream versions. `web/templates/layouts/base.html` loads both files plus the
shared module `web/static/js/htmx-conventions.js`, so every authenticated page
has htmx available. The login and registration layout stays a full-page cookie
flow and does not load htmx.

The shared module implements CSRF header injection, error-envelope rendering,
`showFlash` handling, and the `aria-busy` loading hook. Flash markup reuses
`window.GoGarage.initFlashMessage()` exported from `web/static/js/main.js` for
close and auto-dismiss behavior rather than duplicating it (PR #251).

## Request encoding and form migration

Mutation forms use the `json-enc` htmx extension and submit JSON using the
field names expected by the corresponding API request type:

```html
<form hx-post="/api/v1/vehicles"
      hx-ext="json-enc"
      hx-swap="none">
```

Use `hx-put` and `hx-delete` for updates and deletion. The API URL must
contain resource identifiers that are represented by the route rather than
the JSON body, such as a vehicle ID for nested maintenance or fuel creation.
PUT request bodies follow the corresponding OpenAPI input schema and include
its required fields. The existing PUT handlers update only fields supplied in
the request, so omitted optional properties remain unchanged; do not describe
or implement PUT as a full replacement.

The existing maintenance creation form can submit multiple records, but
`POST /api/v1/vehicles/{vehicleId}/maintenance` accepts one record per request.
Migrating that form therefore requires more advanced client-side batching, with
one POST per maintenance record.

Keep existing HTML validation attributes and labels. Each field-error target
uses the submitted API field name:

```html
<span class="field-error" data-field="vin"></span>
```

Forms also provide a general error region for errors without a field:

```html
<div data-form-errors role="alert" aria-live="assertive"></div>
```

Mutation requests use `hx-swap="none"` because API handlers return JSON, not
HTML fragments. This prevents a JSON response from being inserted into the
document. Successful requests navigate using `HX-Redirect`; failed requests
are rendered by the shared htmx module while preserving the submitted form.

## Success and navigation

Successful mutation responses retain the existing JSON success envelope and
may add these htmx response headers:

- `HX-Redirect: /destination` navigates to the canonical page after create,
  update, or delete.
- `HX-Trigger: {"showFlash":{"type":"success","message":"..."}}` displays a
  flash message after navigation or before an in-place UI update.

The default convention is to redirect after every mutation. This keeps page
data, navigation, and server-rendered sections consistent without making API
handlers render HTML. A future in-place update is allowed only when the
endpoint and template explicitly define an HTML-fragment response; it must
not be implemented by swapping the API's JSON body.

Recommended destinations:

| Action | Destination |
| --- | --- |
| Vehicle create/edit | Vehicle detail |
| Vehicle delete | Vehicle list |
| Maintenance create/edit | Maintenance detail |
| Maintenance delete | Parent vehicle detail or maintenance list |
| Fuel create/edit | Fuel detail |
| Fuel delete | Parent vehicle detail or fuel list |
| Profile edit | Profile |
| Change password | Profile |

Vehicle list and detail delete controls use `hx-delete` against
`/api/v1/vehicles/{id}` with `hx-swap="none"`. Successful htmx deletes return
`HX-Redirect: /vehicles`; API clients without `HX-Request` still receive the
JSON success envelope without a redirect (PR #257).

The add-vehicle form uses `hx-post="/api/v1/vehicles"` with `json-enc` and
`hx-swap="none"`. The create API converts htmx's string-valued numeric inputs
and returns field-level errors, including duplicate VIN errors. Successful
htmx creation redirects to `/vehicles/{id}?added=true`; the detail page renders
the success flash after navigation. The old `POST /vehicles/new` route is
removed; the now-unused `VehicleCreate` page handler remains for cleanup in
#246 (PR #258).

The [add-vehicle form preview](../docs/screenshots/vehicle-create-htmx.png)
shows the inline duplicate-VIN error state.

The edit-vehicle form uses `hx-put="/api/v1/vehicles/{id}"` with `json-enc` and
`hx-swap="none"`. It submits every editable form field, including unchanged
values, so the server-rendered form remains a complete edit representation.
The API's supplied-field update behavior remains unchanged for other clients;
the htmx form applies all submitted fields, treating blank optional values as
clears. Numeric inputs are normalized from `json-enc` strings and validation
errors use the shared inline field-error rendering. Successful edits redirect
to `/vehicles/{id}?updated=true`, where the detail page displays the success
flash. The old `POST /vehicles/{id}/edit` route is removed; the now-unused
`VehicleUpdate` page handler remains for cleanup in #246 (PR #263).
The [vehicle edit form preview](../docs/screenshots/vehicle-edit-htmx.png)
shows the pre-populated htmx form.

### Profile editing

- [x] Migrate profile editing to `PUT /api/v1/users/me` (PR #269).

The edit-profile form uses `hx-put="/api/v1/users/me"` with `json-enc` and
`hx-swap="none"`. It submits `username`, `email`, `first_name`, and `last_name`
as strings. Supplied strings are trimmed; blank optional names clear their
values, while blank usernames or emails are validated rather than ignored.
Omitted fields remain unchanged for partial API updates. Validation errors and
htmx duplicate username/email conflicts include field details for the shared
inline error renderer. Other API clients retain the existing JSON response
shape without redirect headers.

Successful htmx edits return `HX-Redirect: /profile?updated=true`, where the
profile page displays the success flash. The legacy `POST /profile/edit` route
and `ProfileUpdate` page handler are removed.

The [profile edit form preview](../docs/screenshots/profile-edit-htmx.png)
shows the inline duplicate-username error state.

### Password changes

- [x] Migrate the change-password form to `PUT /api/v1/users/me/password` (PR #270).

The form uses `hx-put`, `json-enc`, `hx-swap="none"`, and the shared CSRF
header injection. Incorrect current passwords, mismatched confirmations, and
password-strength failures use inline field errors. Confirmation is required
for htmx requests; other API clients may omit it, but supplied confirmations
must match. Passwords are passed unchanged, never trimmed.

Successful htmx changes return `HX-Redirect: /profile?password_changed=true`
so the profile page displays the existing success flash. Existing access and
refresh tokens remain valid, matching the current authentication behavior;
only subsequent password-based logins require the new password.
The legacy `POST /profile/password` route and `ChangePasswordSubmit` handler
are removed. A native `method="POST"` fallback prevents passwords from entering
the URL if htmx cannot load; it fails safely rather than changing the password.
Password values are never included in responses or request logs,
and the shared error renderer uses only error messages, not submitted values.

The [password form preview](../docs/screenshots/password-htmx.png)
shows an inline validation error without exposing password values.

### Maintenance creation

The add-maintenance page preserves the shared vehicle dropdown (including
`vehicle` and `vehicle_id` query preselection) and "Include More Records" flow.
Each record card is a separate form with `hx-post`, `json-enc`, and
`hx-swap="none"`. `maintenance-create.js` validates the vehicle and all cards
before saving, sets the nested API URL from the current dropdown selection
during `htmx:configRequest`, and submits one record at a time. Numeric form
strings are normalized by the create API; blank cost and mileage stay unset.

Intermediate successful responses do not follow `HX-Redirect`. Their saved
cards are removed, and the next card is submitted only after the previous
request completes. Errors stop the batch, render through the shared inline
error handler on the failing card, preserve unsaved values, and re-enable
controls. A retry submits only the remaining cards, not confirmed saves.
As with any non-idempotent POST, a lost response after a server commit has an
ambiguous outcome; users should check existing records before retrying a
network failure. The final success follows
`HX-Redirect: /maintenance/{id}?added=true`; the detail page displays the
creation flash once through the shared layout.

The old `POST /maintenance/new` route is removed. The unused
`MaintenanceCreate` page handler and its parsing helpers remain for #246,
consistent with the vehicle migrations. The [maintenance creation preview](../docs/screenshots/maintenance-create-htmx.png)
shows a stopped bulk submission with a field error and only unsaved cards
remaining. Focused JavaScript batching tests run with
`node --test tests/maintenance-create.test.cjs` (PR #265).

### Maintenance editing

The edit-maintenance form uses `hx-put="/api/v1/maintenance/{id}"` with
`json-enc`, `hx-swap="none"`, and native HTML validation. It submits all seven
editable fields, including unchanged custom service type, provider, and notes;
the disabled vehicle display is not submitted and ownership stays unchanged.
The htmx API path parses string-valued numeric inputs, validates the complete
record, and treats blank optional values as clears. Explicit service update
flags clear nullable cost and mileage without changing the existing
supplied-field behavior for non-htmx API clients.

Successful edits return `HX-Redirect: /maintenance/{id}?updated=true`; the
detail page uses the layout's flash partial to show one success notification.
API validation errors use the shared inline targets without replacing the
form. The old `POST /maintenance/{id}/edit` route is removed; the unused
`MaintenanceUpdate` page handler remains for cleanup in #246.
The [maintenance edit preview](../docs/screenshots/maintenance-edit-htmx.png)
shows a future-date validation error with the submitted values preserved
(PR #266).

### Maintenance deletion

Maintenance list and detail delete controls use
`hx-delete="/api/v1/maintenance/{id}"`, the irreversible-action `hx-confirm`,
and `hx-swap="none"`. Successful htmx deletion redirects to the parent vehicle
detail at `/vehicles/{vehicleId}`; non-htmx API clients retain the JSON success
envelope without a redirect. Cookie-authenticated deletion requires the shared
`X-CSRF-Token` header. The old hidden POST delete form is removed; there is no
registered maintenance page-delete route or handler to retain for #246.
The [maintenance delete preview](../docs/screenshots/maintenance-delete-htmx.png)
shows the migrated controls (PR #267).

The fuel forms and delete button use the Fuel REST API. The add form uses
`hx-post="/api/v1/vehicles/{vehicleId}/fuel"`; its vehicle `<select>` rewrites
the request path in `hx-on::config-request`, so the URL always targets the
selected vehicle (and honors the `?vehicle=` preselection). The edit form uses
`hx-put="/api/v1/fuel/{id}"` and the detail page uses
`hx-delete="/api/v1/fuel/{id}"` with `hx-confirm`; the fuel list has no per-row
delete. For `HX-Request` calls the API normalizes `json-enc` string values
(blank optional numbers are omitted, an unchecked `partial_fill` checkbox
becomes `false`) and returns inline field errors. Success redirects with
`HX-Redirect` are preserved: create to `/fuel?added=true`, edit to
`/fuel/{id}?updated=true`, delete to `/fuel?vehicle={vehicleId}`. The
`POST /fuel/new`, `POST /fuel/{id}/edit`, and `POST /fuel/{id}/delete` routes
and their `FuelCreate`/`FuelUpdate`/`FuelDelete` page handlers are removed
(PR #264). The [fuel form preview](../docs/screenshots/fuel-create-htmx.png) shows an inline validation error.

## Error responses and rendering

API handlers use the existing JSON envelope:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Please correct the highlighted fields",
    "details": [
      {"field": "vin", "message": "VIN is invalid"}
    ]
  }
}
```

The `details` array is optional. `field` values must match form control names
and `data-field` values exactly. The API handler remains responsible for
mapping domain errors to this envelope; it does not render templates.

The shared htmx module listens for `htmx:responseError` and:

1. Parses the JSON error envelope.
2. Clears stale field and general errors from the submitted form.
3. Writes each field message to the matching `[data-field="<field>"]` element
   and marks the related control invalid.
4. Writes the general `error.message` to `[data-form-errors]` when no
   field-specific target exists.
5. Focuses the first invalid control and keeps the response accessible through
   the form's live error region.

Error rendering must never echo submitted values. This is especially important
for password forms: password fields are never included in API errors,
notifications, or DOM content.

Authentication redirects and CSRF failures are handled before field rendering:
`HX-Redirect` wins for unauthenticated htmx requests, while a general error
message is shown for a CSRF failure.

## Flash messages

The existing flash-message classes and close/auto-dismiss behavior remain the
visual convention. API mutations do not set flash cookies. Instead, handlers
send an `HX-Trigger` event:

```http
HX-Trigger: {"showFlash":{"type":"success","message":"Vehicle saved"}}
```

The shared htmx module listens for `showFlash`, creates the same flash markup used by `web/templates/partials/flash-messages.html`, creates a `.flash-messages` region when the partial rendered none, appends the message there, and ensures that region has `role="status"` and `aria-live="polite"` for non-error notifications. It must also initialize the new message's close and auto-dismiss behavior (or use delegated handlers), since the existing one-time `initFlashMessages()` does not cover dynamically appended nodes.
`aria-live="polite"` for non-error notifications. Error notifications use
`role="alert"` and `aria-live="assertive"`.

Flash `type` values are limited to the existing styles: `success`, `error`,
`warning`, and `info`. Messages are server-provided text and must be inserted
with `textContent`, never `innerHTML`.

## Delete and loading behavior

Delete controls use htmx's built-in confirmation and request attributes:

```html
<button type="button"
        hx-delete="/api/v1/vehicles/{{.Vehicle.ID}}"
        hx-confirm="Are you sure you want to delete this vehicle? This action cannot be undone."
        hx-swap="none">
    Delete
</button>
```

Do not add new hidden delete forms or `data-confirm-delete` wiring for
migrated actions. The confirmation text must explain that the action cannot be
undone, and the control must remain keyboard reachable.

Use htmx's `htmx-request` class for loading states where possible. The
existing shared styles may disable the submitting control and expose
`aria-busy="true"` during a request. Custom JavaScript should be limited to
error rendering, CSRF header injection, flash handling, and behavior that
htmx cannot provide.

## Contributor checklist

For each mutation being migrated:

1. Confirm the matching API endpoint, request fields, ownership checks, and
   PUT's supplied-field update behavior in `spec/openapi.yaml` and the API
   handler.
2. Keep the GET page handler and server-rendered form; replace only the
   mutation action with `hx-post`, `hx-put`, or `hx-delete`.
3. Add `json-enc` and `hx-swap="none"` to JSON mutation forms.
4. Ensure form control names match API fields and inline error `data-field`
   values.
5. Preserve HTML validation, labels, focus behavior, and keyboard access.
6. Verify the API supplies `HX-Redirect` and, where appropriate, an
   `HX-Trigger` flash event.
7. Verify cookie-authenticated requests include `X-CSRF-Token`, while bearer
   API clients continue to work unchanged.
8. Test success, validation errors, authorization failures, CSRF failures,
   network failures, and delete confirmation behavior.
9. Remove the old mutating page handler and route only after the htmx flow is
   verified, or track the removal for the cleanup issue.
10. Update the relevant milestone/spec checkbox with the landing PR number.

## Layer review

These conventions preserve the architecture boundaries in
`spec/architecture.md`:

- Page handlers render GET pages and provide form/page data.
- API handlers parse requests, call services, set status/headers, and encode
  JSON; they do not render templates.
- Services retain business rules, validation orchestration, and ownership
  checks.
- Shared JavaScript translates API responses into browser presentation only.
