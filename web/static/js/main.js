// Go-Garage - Main JavaScript

"use strict";

(function () {
    // ========================================
    // Flash Message Dismissal
    // ========================================

    /**
     * Initialize close buttons on flash messages.
     * Clicking the close button removes the flash element from the DOM.
     */
    function initFlashMessages() {
        var autoDismissDelay = 5000;
        var flashes = document.querySelectorAll(".flash");

        flashes.forEach(function (flash) {
            var closeBtn = flash.querySelector(".flash-close");
            if (closeBtn) {
                closeBtn.addEventListener("click", function () {
                    flash.remove();
                });
            }

            // Auto-dismiss after delay; skip for error messages.
            if (!flash.classList.contains("flash-error")) {
                setTimeout(function () {
                    flash.classList.add("flash-dismiss");
                    flash.addEventListener("transitionend", function () {
                        flash.remove();
                    }, { once: true });
                }, autoDismissDelay);
            }
        });
    }

    // ========================================
    // Mobile Navbar Toggle
    // ========================================

    /**
     * Toggle the mobile navigation menu visibility.
     * Updates aria-expanded attribute for accessibility.
     */
    function initNavbarToggle() {
        var toggle = document.querySelector(".navbar-toggle");
        var menu = document.querySelector(".navbar-menu");
        if (!toggle || !menu) {
            return;
        }

        toggle.addEventListener("click", function () {
            var expanded = toggle.getAttribute("aria-expanded") === "true";
            toggle.setAttribute("aria-expanded", String(!expanded));
            menu.classList.toggle("active");
        });
    }

    // ========================================
    // User Dropdown Menu
    // ========================================

    /**
     * Initialize the user dropdown menu in the navigation bar.
     * Clicking the toggle button opens/closes the dropdown.
     * Clicking outside the dropdown or pressing Escape closes it.
     * Arrow keys navigate between menu items when the dropdown is open.
     */
    function initUserDropdown() {
        var toggle = document.querySelector(".navbar-user-toggle");
        var menu = document.querySelector(".navbar-dropdown-menu");
        if (!toggle || !menu) {
            return;
        }

        function getMenuItems() {
            return menu.querySelectorAll('[role="menuitem"]');
        }

        function closeMenu() {
            menu.classList.remove("active");
            toggle.setAttribute("aria-expanded", "false");
        }

        toggle.addEventListener("click", function (e) {
            e.stopPropagation();
            var expanded = toggle.getAttribute("aria-expanded") === "true";
            toggle.setAttribute("aria-expanded", String(!expanded));
            menu.classList.toggle("active");
            if (!expanded) {
                var items = getMenuItems();
                if (items.length > 0) {
                    items[0].focus();
                }
            }
        });

        document.addEventListener("click", function () {
            if (menu.classList.contains("active")) {
                closeMenu();
            }
        });

        document.addEventListener("keydown", function (e) {
            if (e.key === "Escape" && menu.classList.contains("active")) {
                closeMenu();
                toggle.focus();
            }
        });

        menu.addEventListener("keydown", function (e) {
            var items = getMenuItems();
            if (items.length === 0) { return; }
            var current = document.activeElement;
            var index = Array.prototype.indexOf.call(items, current);

            if (e.key === "ArrowDown") {
                e.preventDefault();
                var next = (index + 1) % items.length;
                items[next].focus();
            } else if (e.key === "ArrowUp") {
                e.preventDefault();
                var prev = (index - 1 + items.length) % items.length;
                items[prev].focus();
            }
        });
    }

    // ========================================
    // Dark Mode Toggle
    // ========================================

    /**
     * Initialize the dark mode toggle button.
     * Reads saved theme from localStorage, applies it, and handles toggle clicks.
     */
    function initThemeToggle() {
        var btn = document.getElementById("theme-toggle");
        if (!btn) {
            return;
        }

        function isDark() {
            var theme = document.documentElement.getAttribute("data-theme");
            if (theme === "dark") { return true; }
            if (theme === "light") { return false; }
            return window.matchMedia("(prefers-color-scheme: dark)").matches;
        }

        function updateButton() {
            var dark = isDark();
            btn.setAttribute("aria-pressed", dark ? "true" : "false");
            btn.setAttribute("aria-label", dark ? "Switch to light mode" : "Switch to dark mode");
        }

        updateButton();

        btn.addEventListener("click", function () {
            var newTheme = isDark() ? "light" : "dark";
            document.documentElement.setAttribute("data-theme", newTheme);
            localStorage.setItem("theme", newTheme);
            updateButton();
        });
    }

    // ========================================
    // Form Submit Loading States
    // ========================================

    /**
     * Intercepts form submissions and adds a loading spinner to the
     * submit button to prevent double submissions.
     */
    function initFormSubmitLoading() {
        document.addEventListener("submit", function (e) {
            if (e.defaultPrevented) { return; }
            var form = e.target;
            if (form.tagName !== "FORM") { return; }
            var btn = form.querySelector('button[type="submit"]');
            if (!btn || btn.classList.contains("btn-loading")) { return; }

            btn.classList.add("btn-loading");
            btn.setAttribute("disabled", "disabled");
            btn.setAttribute("aria-busy", "true");
        });
    }

    // ========================================
    // Delete Button Loading States
    // ========================================

    /**
     * Sets up delete buttons that use data-confirm-delete attribute.
     * Shows a confirmation dialog, then adds a loading spinner while
     * the hidden delete form is being submitted.
     */
    function initDeleteButtons() {
        var buttons = document.querySelectorAll("[data-confirm-delete]");
        buttons.forEach(function (btn) {
            var formId = btn.getAttribute("data-confirm-delete");
            var form = document.getElementById(formId);
            if (!form) { return; }

            btn.addEventListener("click", function () {
                var message = btn.getAttribute("data-confirm-message") ||
                    "Are you sure? This action cannot be undone.";
                if (confirm(message)) {
                    btn.classList.add("btn-loading");
                    btn.setAttribute("disabled", "disabled");
                    btn.setAttribute("aria-busy", "true");
                    form.submit();
                }
            });
        });
    }

    // ========================================
    // Initialize on DOM Ready
    // ========================================

    document.addEventListener("DOMContentLoaded", function () {
        initFlashMessages();
        initNavbarToggle();
        initUserDropdown();
        initThemeToggle();
        initFormSubmitLoading();
        initDeleteButtons();
    });
})();
