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
    // Initialize on DOM Ready
    // ========================================

    document.addEventListener("DOMContentLoaded", function () {
        initFlashMessages();
        initNavbarToggle();
    });
})();
