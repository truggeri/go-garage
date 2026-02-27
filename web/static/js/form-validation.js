// Go-Garage - Client-Side Form Validation

"use strict";

(function () {
    // ========================================
    // Validation Rules
    // ========================================

    var EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    var PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;

    /**
     * Validate a single form field based on its HTML5 attributes and custom rules.
     * Returns an error message string, or empty string if valid.
     */
    function validateField(field) {
        var value = field.value.trim();
        var tagName = field.tagName.toLowerCase();

        // Select elements: check if a value is selected
        if (tagName === "select") {
            if (field.hasAttribute("required") && !value) {
                return "Please select an option.";
            }
            return "";
        }

        // Required check
        if (field.hasAttribute("required") && !value) {
            return "This field is required.";
        }

        // Skip further checks if empty and not required
        if (!value) {
            return "";
        }

        // Email format
        if (field.type === "email" && !EMAIL_REGEX.test(value)) {
            return "Please enter a valid email address.";
        }

        // MinLength
        var minLength = field.getAttribute("minlength");
        if (minLength && value.length < parseInt(minLength, 10)) {
            return "Must be at least " + minLength + " characters.";
        }

        // MaxLength (browser usually enforces, but validate anyway)
        var maxLength = field.getAttribute("maxlength");
        if (maxLength && value.length > parseInt(maxLength, 10)) {
            return "Must be no more than " + maxLength + " characters.";
        }

        // Number validation
        if (field.type === "number") {
            var num = parseFloat(value);
            if (isNaN(num)) {
                return "Please enter a valid number.";
            }
            var min = field.getAttribute("min");
            if (min !== null && num < parseFloat(min)) {
                return "Value must be at least " + min + ".";
            }
            var max = field.getAttribute("max");
            if (max !== null && num > parseFloat(max)) {
                return "Value must be no more than " + max + ".";
            }
        }

        // Date validation
        if (field.type === "date" && isNaN(Date.parse(value))) {
            return "Please enter a valid date.";
        }

        // Password strength (fields with data-validate-password)
        if (field.getAttribute("data-validate-password") === "true") {
            if (!PASSWORD_REGEX.test(field.value)) {
                return "Password does not meet the requirements.";
            }
        }

        // Password confirmation (fields with data-match-field)
        var matchId = field.getAttribute("data-match-field");
        if (matchId) {
            var matchField = field.form.querySelector("#" + matchId);
            if (matchField && field.value !== matchField.value) {
                return "Passwords do not match.";
            }
        }

        return "";
    }

    // ========================================
    // UI Feedback
    // ========================================

    /**
     * Get or create the error message element for a field.
     * Appends to the end of the form-group to avoid conflicts with hint text.
     */
    function getErrorElement(field) {
        var group = field.closest(".form-group");
        if (!group) {
            return null;
        }
        var errorEl = group.querySelector(".form-error.js-validation-error");
        if (!errorEl) {
            errorEl = document.createElement("p");
            errorEl.className = "form-error js-validation-error";
            errorEl.setAttribute("aria-live", "polite");
            group.appendChild(errorEl);
        }
        return errorEl;
    }

    /**
     * Show validation error on a field.
     */
    function showError(field, message) {
        field.classList.add("is-invalid");
        field.classList.remove("is-valid");
        field.setAttribute("aria-invalid", "true");
        var errorEl = getErrorElement(field);
        if (errorEl) {
            errorEl.textContent = message;
            errorEl.style.display = "";
            var errorId = errorEl.id || field.id + "-validation-error";
            errorEl.id = errorId;
            addDescribedBy(field, errorId);
        }
    }

    /**
     * Show validation success on a field.
     */
    function showSuccess(field) {
        field.classList.remove("is-invalid");
        field.classList.add("is-valid");
        field.removeAttribute("aria-invalid");
        var errorEl = getErrorElement(field);
        if (errorEl) {
            errorEl.textContent = "";
            errorEl.style.display = "none";
            removeDescribedBy(field, errorEl.id);
        }
    }

    /**
     * Clear all validation state from a field.
     */
    function clearValidation(field) {
        field.classList.remove("is-invalid", "is-valid");
        field.removeAttribute("aria-invalid");
        var errorEl = field.closest(".form-group") &&
            field.closest(".form-group").querySelector(".js-validation-error");
        if (errorEl) {
            errorEl.textContent = "";
            errorEl.style.display = "none";
            removeDescribedBy(field, errorEl.id);
        }
    }

    /**
     * Add an ID to the field's aria-describedby list.
     */
    function addDescribedBy(field, id) {
        var current = field.getAttribute("aria-describedby") || "";
        var ids = current.split(/\s+/).filter(Boolean);
        if (ids.indexOf(id) === -1) {
            ids.push(id);
        }
        field.setAttribute("aria-describedby", ids.join(" "));
    }

    /**
     * Remove an ID from the field's aria-describedby list.
     */
    function removeDescribedBy(field, id) {
        var current = field.getAttribute("aria-describedby") || "";
        var ids = current.split(/\s+/).filter(function (v) {
            return v && v !== id;
        });
        if (ids.length > 0) {
            field.setAttribute("aria-describedby", ids.join(" "));
        } else {
            field.removeAttribute("aria-describedby");
        }
    }

    // ========================================
    // Form Initialization
    // ========================================

    /**
     * Get all validatable fields from a form.
     */
    function getValidatableFields(form) {
        return form.querySelectorAll(
            "input:not([type=hidden]):not([type=submit]):not(:disabled), " +
            "select:not(:disabled), " +
            "textarea:not(:disabled)"
        );
    }

    /**
     * Check whether a field should be validated.
     * Only validate fields that have validation constraints.
     */
    function hasConstraints(field) {
        return field.hasAttribute("required") ||
            field.hasAttribute("min") ||
            field.hasAttribute("max") ||
            field.hasAttribute("minlength") ||
            field.hasAttribute("maxlength") ||
            field.type === "email" ||
            field.type === "number" ||
            field.type === "date" ||
            field.getAttribute("data-validate-password") === "true" ||
            field.hasAttribute("data-match-field");
    }

    /**
     * Validate a single field and update UI. Returns true if valid.
     */
    function validateAndShow(field) {
        if (!hasConstraints(field)) {
            return true;
        }
        var error = validateField(field);
        if (error) {
            showError(field, error);
            return false;
        }
        showSuccess(field);
        return true;
    }

    /**
     * Initialize client-side validation on all forms with novalidate attribute.
     */
    function initFormValidation() {
        var forms = document.querySelectorAll("form[novalidate]");

        forms.forEach(function (form) {
            var fields = getValidatableFields(form);

            // Real-time validation on blur
            fields.forEach(function (field) {
                field.addEventListener("blur", function () {
                    // Only validate if user has interacted (field is not pristine)
                    if (field.value || field.classList.contains("is-invalid")) {
                        validateAndShow(field);
                    }
                });

                // Clear error on input to give immediate feedback
                field.addEventListener("input", function () {
                    if (field.classList.contains("is-invalid")) {
                        var error = validateField(field);
                        if (!error) {
                            showSuccess(field);
                        }
                    }
                });
            });

            // Validate all fields on submit
            form.addEventListener("submit", function (e) {
                var isValid = true;
                var firstInvalid = null;

                fields.forEach(function (field) {
                    if (!validateAndShow(field)) {
                        isValid = false;
                        if (!firstInvalid) {
                            firstInvalid = field;
                        }
                    }
                });

                if (!isValid) {
                    e.preventDefault();
                    if (firstInvalid) {
                        firstInvalid.focus();
                    }
                }
            });
        });
    }

    // ========================================
    // Initialize on DOM Ready
    // ========================================

    document.addEventListener("DOMContentLoaded", function () {
        initFormValidation();
    });
})();
