// Go-Garage - Shared htmx conventions
//
// Implements the browser-side conventions described in spec/htmx-migration.md:
// CSRF header injection, JSON error-envelope rendering, and HX-Trigger driven
// flash messages. Feature templates only need hx-* attributes; all shared
// behavior lives here.

"use strict";

(function () {
    var FLASH_TYPES = ["success", "error", "warning", "info"];
    var GENERAL_ERROR_SELECTOR = "[data-form-errors]";

    // ========================================
    // CSRF Header Injection
    // ========================================

    /**
     * Read the CSRF token rendered by the authenticated layout.
     * @returns {string} token value, or an empty string when absent.
     */
    function csrfToken() {
        var meta = document.querySelector('meta[name="csrf-token"]');
        return meta ? meta.getAttribute("content") || "" : "";
    }

    /**
     * Add the X-CSRF-Token header to every htmx request so that
     * cookie-authenticated API mutations pass CSRF validation.
     */
    function initCSRFHeader() {
        document.addEventListener("htmx:configRequest", function (e) {
            var token = csrfToken();
            if (token) {
                e.detail.headers["X-CSRF-Token"] = token;
            }
        });
    }

    // ========================================
    // Error Envelope Rendering
    // ========================================

    /**
     * Find the form associated with the element that issued the request.
     * Delete buttons may live outside a form, in which case null is returned.
     */
    function requestForm(elt) {
        if (!elt || typeof elt.closest !== "function") {
            return null;
        }
        return elt.closest("form");
    }

    /**
     * Remove field and general error messages left over from a previous
     * submission so that stale text is never shown alongside new errors.
     */
    function clearErrors(scope) {
        if (!scope) {
            return;
        }

        scope.querySelectorAll("[data-field]").forEach(function (target) {
            target.textContent = "";
        });

        scope.querySelectorAll('[aria-invalid="true"]').forEach(function (control) {
            control.removeAttribute("aria-invalid");
        });

        scope.querySelectorAll(GENERAL_ERROR_SELECTOR).forEach(function (target) {
            target.textContent = "";
        });
    }

    /**
     * Parse the {success:false,error:{...}} envelope from a response body.
     * @returns {Object|null} the error object, or null when unavailable.
     */
    function parseErrorEnvelope(responseText) {
        if (!responseText) {
            return null;
        }

        var payload;
        try {
            payload = JSON.parse(responseText);
        } catch (err) {
            return null;
        }

        if (!payload || typeof payload !== "object" || !payload.error) {
            return null;
        }
        return payload.error;
    }

    /**
     * Write each field message to its [data-field] target and mark the
     * matching control invalid. Returns the first invalid control, if any.
     */
    function renderFieldErrors(scope, details) {
        var firstInvalid = null;

        details.forEach(function (detail) {
            if (!detail || !detail.field) {
                return;
            }

            var target = scope.querySelector('[data-field="' + CSS.escape(detail.field) + '"]');
            if (target) {
                target.textContent = detail.message || "";
            }

            var control = scope.querySelector('[name="' + CSS.escape(detail.field) + '"]');
            if (control) {
                control.setAttribute("aria-invalid", "true");
                if (!firstInvalid) {
                    firstInvalid = control;
                }
            }
        });

        return firstInvalid;
    }

    /**
     * Show a message that is not tied to a single field. Falls back to an
     * error flash message when the form has no general error region.
     */
    function renderGeneralError(scope, message) {
        if (!message) {
            return;
        }

        var target = scope ? scope.querySelector(GENERAL_ERROR_SELECTOR) : null;
        if (target) {
            target.textContent = message;
            return;
        }
        showFlash({ type: "error", message: message });
    }

    /**
     * Render an error envelope into the originating form. Formless requests
     * leave existing form errors untouched and show a general error flash.
     */
    function renderError(elt, responseText) {
        var form = requestForm(elt);
        clearErrors(form);

        var error = parseErrorEnvelope(responseText);
        if (!error) {
            renderGeneralError(form, "Something went wrong. Please try again.");
            return;
        }

        var details = Array.isArray(error.details) ? error.details : [];
        var firstInvalid = form ? renderFieldErrors(form, details) : null;

        if (!firstInvalid) {
            renderGeneralError(form, error.message);
        }

        if (firstInvalid && typeof firstInvalid.focus === "function") {
            firstInvalid.focus();
        }
    }

    /**
     * Render API error responses and network failures using the shared
     * conventions rather than swapping JSON into the document.
     */
    function initErrorRendering() {
        document.addEventListener("htmx:responseError", function (e) {
            renderError(e.detail.elt, e.detail.xhr ? e.detail.xhr.responseText : "");
        });

        document.addEventListener("htmx:sendError", function (e) {
            var form = requestForm(e.detail.elt);
            clearErrors(form);
            renderGeneralError(form, "Unable to reach the server. Please check your connection and try again.");
        });
    }

    // ========================================
    // Flash Messages
    // ========================================

    /**
     * Return the flash message region, creating it when the server-rendered
     * partial produced none for the current page.
     */
    function flashRegion() {
        var region = document.querySelector(".flash-messages");
        if (region) {
            return region;
        }

        region = document.createElement("div");
        region.className = "flash-messages";

        var container = document.querySelector("#main-content") ||
            document.querySelector(".auth-container") ||
            document.body;
        container.insertBefore(region, container.firstChild);
        return region;
    }

    /**
     * Display a flash message using the same markup as
     * web/templates/partials/flash-messages.html.
     * @param {Object} detail - {type, message} provided by the server.
     */
    function showFlash(detail) {
        if (!detail || !detail.message) {
            return;
        }

        var type = FLASH_TYPES.indexOf(detail.type) === -1 ? "info" : detail.type;
        var region = flashRegion();
        region.setAttribute("role", type === "error" ? "alert" : "status");
        region.setAttribute("aria-live", type === "error" ? "assertive" : "polite");

        var flash = document.createElement("div");
        flash.className = "flash flash-" + type;

        var text = document.createElement("span");
        text.className = "flash-text";
        text.textContent = detail.message;

        var close = document.createElement("button");
        close.className = "flash-close";
        close.setAttribute("aria-label", "Close");
        close.textContent = "\u00d7";

        flash.appendChild(text);
        flash.appendChild(close);
        region.appendChild(flash);

        if (window.GoGarage && typeof window.GoGarage.initFlashMessage === "function") {
            window.GoGarage.initFlashMessage(flash);
        }
    }

    /**
     * Listen for the showFlash event sent by API handlers through HX-Trigger.
     */
    function initFlashTrigger() {
        document.addEventListener("showFlash", function (e) {
            showFlash(e.detail);
        });
    }

    // ========================================
    // Loading States
    // ========================================

    /**
     * Expose aria-busy on the element making a request. htmx already adds the
     * htmx-request class, which the shared styles use for the spinner.
     */
    function initLoadingState() {
        document.addEventListener("htmx:beforeRequest", function (e) {
            var elt = e.detail.elt;
            if (elt && typeof elt.setAttribute === "function") {
                elt.setAttribute("aria-busy", "true");
            }
        });

        document.addEventListener("htmx:afterRequest", function (e) {
            var elt = e.detail.elt;
            if (elt && typeof elt.removeAttribute === "function") {
                elt.removeAttribute("aria-busy");
            }
        });
    }

    // ========================================
    // Initialize on DOM Ready
    // ========================================

    function init() {
        initCSRFHeader();
        initErrorRendering();
        initFlashTrigger();
        initLoadingState();
    }

    window.GoGarage = window.GoGarage || {};
    window.GoGarage.showFlash = showFlash;

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", init);
    } else {
        init();
    }
})();
