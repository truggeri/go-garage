// Go-Garage - Bulk Maintenance Record Creation

"use strict";

(function () {
    /**
     * Initialize bulk maintenance form functionality.
     * Handles adding and removing record cards dynamically.
     */
    function initBulkMaintenance() {
        var container = document.getElementById("records-container");
        var addBtn = document.getElementById("add-record-btn");
        if (!container || !addBtn) {
            return;
        }

        function getRecordCards() {
            return container.querySelectorAll(".record-card");
        }

        /**
         * Update the record number in each card title and ensure
         * the first card does not show a remove button.
         */
        function updateRecordNumbers() {
            var cards = getRecordCards();
            cards.forEach(function (card, index) {
                var title = card.querySelector(".record-title");
                if (title) {
                    title.textContent = "Record #" + (index + 1);
                }
                var removeBtn = card.querySelector(".record-remove-btn");
                if (index === 0 && removeBtn) {
                    removeBtn.style.display = "none";
                } else if (index > 0 && removeBtn) {
                    removeBtn.style.display = "";
                }
            });
        }

        /**
         * Update all field IDs, label for-attributes, and aria-describedby
         * attributes within a card to use the given index.
         */
        function updateFieldIds(card, index) {
            var fields = card.querySelectorAll("input, select, textarea");
            fields.forEach(function (field) {
                var name = field.getAttribute("name");
                if (name && field.type !== "hidden") {
                    field.id = name + "_" + index;
                }
            });

            var labels = card.querySelectorAll("label");
            labels.forEach(function (label) {
                var forAttr = label.getAttribute("for");
                if (forAttr) {
                    var baseName = forAttr.replace(/_\d+$/, "");
                    label.setAttribute("for", baseName + "_" + index);
                }
            });

            fields.forEach(function (field) {
                var describedby = field.getAttribute("aria-describedby");
                if (describedby) {
                    var baseId = describedby.replace(/-\d+$/, "");
                    field.setAttribute("aria-describedby", baseId + "-" + index);
                }
            });

            var errors = card.querySelectorAll(".form-error");
            errors.forEach(function (errEl) {
                if (errEl.id) {
                    var baseId = errEl.id.replace(/-\d+$/, "");
                    errEl.id = baseId + "-" + index;
                }
            });
        }

        /**
         * Remove a record card and re-number the remaining cards.
         */
        function removeRecord(card) {
            card.remove();
            updateRecordNumbers();
        }

        /**
         * Clone the first record card, clear its values, and append it.
         */
        function addRecord() {
            var cards = getRecordCards();
            var template = cards[0];
            var newCard = template.cloneNode(true);
            var newIndex = cards.length;

            // Clear all field values.
            var inputs = newCard.querySelectorAll("input, textarea");
            inputs.forEach(function (field) {
                if (field.type !== "hidden") {
                    field.value = "";
                }
            });

            // Reset select elements to their first option.
            var selects = newCard.querySelectorAll("select");
            selects.forEach(function (sel) {
                sel.selectedIndex = 0;
            });

            // Hide custom service type group and remove required.
            var customGroup = newCard.querySelector(".custom-service-type-group");
            if (customGroup) {
                customGroup.classList.add("hidden");
                var customInput = customGroup.querySelector("input");
                if (customInput) {
                    customInput.removeAttribute("required");
                    customInput.value = "";
                }
            }

            // Remove any server-rendered error messages.
            var errors = newCard.querySelectorAll(".form-error");
            errors.forEach(function (el) {
                el.remove();
            });

            // Clear validation states.
            var invalidFields = newCard.querySelectorAll(".is-invalid, .is-valid");
            invalidFields.forEach(function (el) {
                el.classList.remove("is-invalid", "is-valid");
                el.removeAttribute("aria-invalid");
            });

            // Remove any JS validation error elements.
            var jsErrors = newCard.querySelectorAll(".js-validation-error");
            jsErrors.forEach(function (el) {
                el.remove();
            });

            // Update field IDs for the new index.
            updateFieldIds(newCard, newIndex);

            // Ensure remove button is present.
            var header = newCard.querySelector(".record-card-header");
            var existingRemoveBtn = newCard.querySelector(".record-remove-btn");
            if (!existingRemoveBtn && header) {
                var removeBtn = document.createElement("button");
                removeBtn.type = "button";
                removeBtn.className = "btn btn-secondary btn-sm record-remove-btn";
                removeBtn.textContent = "Remove";
                removeBtn.addEventListener("click", function () {
                    removeRecord(newCard);
                });
                header.appendChild(removeBtn);
            } else if (existingRemoveBtn) {
                existingRemoveBtn.style.display = "";
                var newRemoveBtn = existingRemoveBtn.cloneNode(true);
                existingRemoveBtn.parentNode.replaceChild(newRemoveBtn, existingRemoveBtn);
                newRemoveBtn.addEventListener("click", function () {
                    removeRecord(newCard);
                });
            }

            container.appendChild(newCard);
            updateRecordNumbers();

            // Focus the first select or text input in the new card.
            var firstInput = newCard.querySelector("select, input[type='text']");
            if (firstInput) {
                firstInput.focus();
            }
        }

        // Attach click handler to "Include More Records" button.
        addBtn.addEventListener("click", addRecord);

        // Attach click handlers to any existing remove buttons (server-rendered).
        var existingRemoveBtns = container.querySelectorAll(".record-remove-btn");
        existingRemoveBtns.forEach(function (btn) {
            btn.addEventListener("click", function () {
                removeRecord(btn.closest(".record-card"));
            });
        });
    }

    document.addEventListener("DOMContentLoaded", initBulkMaintenance);
})();
