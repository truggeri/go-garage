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

The shared htmx module listens for `showFlash`, creates the same flash markup used by `web/templates/partials/flash-messages.html`, creates a `.flash-messages` region when the partial rendered none, appends the message there, and ensures that region has `role="status"` and `aria-live="polite"` for non-error notifications.
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
