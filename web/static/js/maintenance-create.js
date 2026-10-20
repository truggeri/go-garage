// Maintenance creation batches one REST request per record without replaying saved cards.
"use strict";

(function () {
    document.addEventListener("DOMContentLoaded", function () {
        var container = document.getElementById("bulk-maintenance-form");
        if (!container) return;
        var vehicle = document.getElementById("vehicle_id");
        var submit = document.getElementById("submit-records-btn");
        var status = container.querySelector("[data-bulk-status]");
        var pending = [];
        var saved = 0;
        var busy = false;

        function setBusy(value) {
            busy = value;
            container.setAttribute("aria-busy", String(value));
            vehicle.disabled = value;
            container.querySelectorAll("button").forEach(function (button) {
                button.disabled = value;
            });
            if (!value) {
                document.getElementById("add-record-btn").disabled =
                    container.querySelectorAll(".record-card").length >= 50;
            }
        }

        function start() {
            if (busy) return;
            if (!vehicle.reportValidity()) return;
            pending = Array.from(container.querySelectorAll(".record-card"));
            if (pending.some(function (form) { return !form.reportValidity(); })) return;
            saved = 0;
            status.textContent = "";
            setBusy(true);
            htmx.trigger(pending[0], "maintenance-submit");
        }

        submit.addEventListener("click", start);
        container.addEventListener("submit", function (event) {
            event.preventDefault();
            start();
        });
        container.addEventListener("change", function (event) {
            if (event.target.name !== "service_type") return;
            var group = event.target.closest(".record-card").querySelector(".custom-service-type-group");
            var input = group.querySelector("input");
            var custom = event.target.value === "other";
            group.classList.toggle("hidden", !custom);
            input.required = custom;
            if (!custom) input.value = "";
        });
        container.addEventListener("htmx:configRequest", function (event) {
            event.detail.path = "/api/v1/vehicles/" + encodeURIComponent(vehicle.value) + "/maintenance";
        });
        container.addEventListener("htmx:beforeOnLoad", function (event) {
            // Intermediate successes must not follow the API's detail redirect.
            if (busy && pending.length > 1 && event.detail.xhr.status === 201) {
                event.preventDefault();
            }
        });
        container.addEventListener("htmx:afterRequest", function (event) {
            // htmx also notifies a surviving ancestor when the issuing card is removed.
            if (!busy || event.detail.elt !== pending[0]) return;
            if (event.detail.xhr.status !== 201) {
                pending = [];
                setBusy(false);
                if (saved) {
                    status.textContent = saved + " record(s) saved. Only unsaved records remain; correct the error and retry.";
                }
                return;
            }
            var completed = pending.shift();
            saved++;
            if (pending.length) {
                completed.remove();
                document.getElementById("records-container").dispatchEvent(new Event("maintenance:recordsChanged"));
                status.textContent = saved + " record(s) saved.";
                // Let htmx finish the previous request before issuing the next.
                window.setTimeout(function () {
                    htmx.trigger(pending[0], "maintenance-submit");
                }, 0);
            } else {
                setBusy(false);
            }
        });
        container.addEventListener("htmx:validation:halted", function () {
            pending = [];
            setBusy(false);
            status.textContent = "Check the remaining records before retrying.";
        });
    });
})();
