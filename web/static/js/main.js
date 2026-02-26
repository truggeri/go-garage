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
        document.querySelectorAll(".flash-close").forEach(function (btn) {
            btn.addEventListener("click", function () {
                var flash = btn.closest(".flash");
                if (flash) {
                    flash.remove();
                }
            });
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
     */
    function initUserDropdown() {
        var toggle = document.querySelector(".navbar-user-toggle");
        var menu = document.querySelector(".navbar-dropdown-menu");
        if (!toggle || !menu) {
            return;
        }

        toggle.addEventListener("click", function (e) {
            e.stopPropagation();
            var expanded = toggle.getAttribute("aria-expanded") === "true";
            toggle.setAttribute("aria-expanded", String(!expanded));
            menu.classList.toggle("active");
        });

        document.addEventListener("click", function () {
            if (menu.classList.contains("active")) {
                menu.classList.remove("active");
                toggle.setAttribute("aria-expanded", "false");
            }
        });

        document.addEventListener("keydown", function (e) {
            if (e.key === "Escape" && menu.classList.contains("active")) {
                menu.classList.remove("active");
                toggle.setAttribute("aria-expanded", "false");
                toggle.focus();
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
    // Initialize on DOM Ready
    // ========================================

    document.addEventListener("DOMContentLoaded", function () {
        initFlashMessages();
        initNavbarToggle();
        initUserDropdown();
        initThemeToggle();
    });
})();
