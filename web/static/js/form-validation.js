// Go-Garage - Client-Side Form Validation

"use strict";

(function () {
    // ========================================
    // Validation Rules
    // ========================================

    var EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    var PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;
    var DEFAULT_MIN_LENGTH = 3;
    var DEFAULT_MAX_LENGTH = 20;

    /**
     * Validate a single form field based on its HTML5 attributes and custom rules.
     * Checks required, email format, min/max length, number range, date,
     * password strength, and password confirmation.
     * @param {HTMLElement} field - The form field element to validate.
     * @returns {string} An error message if invalid, or empty string if valid.
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
        var minLength = parseInt(field.getAttribute("minlength"), DEFAULT_MIN_LENGTH);
        if (value.length < minLength) {
            return "Must be at least " + minLength + " characters.";
        }

        // MaxLength (browser usually enforces, but validate anyway)
        var maxLength = parseInt(field.getAttribute("maxlength"), DEFAULT_MAX_LENGTH);
        if (value.length > maxLength) {
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
            if (!PASSWORD_REGEX.test(value)) {
                return "Password does not meet the requirements.";
            }
        }

        // Password confirmation (fields with data-match-field)
        var matchId = field.getAttribute("data-match-field");
        if (matchId) {
            var matchField = field.form.querySelector("#" + matchId);
            if (matchField && value !== matchField.value.trim()) {
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
     * @param {HTMLElement} field - The form field element.
     * @returns {HTMLParagraphElement|null} The error element, or null if no form-group parent.
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
     * Show a validation error on a field by adding the is-invalid class,
     * setting aria-invalid, and displaying the error message.
     * @param {HTMLElement} field - The form field element.
     * @param {string} message - The error message to display.
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
     * Show validation success on a field by adding the is-valid class,
     * removing aria-invalid, and hiding the error message.
     * @param {HTMLElement} field - The form field element.
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
     * Clear all validation state from a field by removing is-invalid/is-valid
     * classes, aria-invalid, and hiding the error message.
     * @param {HTMLElement} field - The form field element.
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
     * Add an ID to the field's aria-describedby attribute list.
     * @param {HTMLElement} field - The form field element.
     * @param {string} id - The ID to add to the aria-describedby list.
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
     * Remove an ID from the field's aria-describedby attribute list.
     * @param {HTMLElement} field - The form field element.
     * @param {string} id - The ID to remove from the aria-describedby list.
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
     * Get all validatable fields from a form (inputs, selects, textareas
     * excluding hidden, submit, and disabled fields).
     * @param {HTMLFormElement} form - The form element to query.
     * @returns {NodeList} A list of validatable field elements.
     */
    function getValidatableFields(form) {
        return form.querySelectorAll(
            "input:not([type=hidden]):not([type=submit]):not(:disabled), " +
            "select:not(:disabled), " +
            "textarea:not(:disabled)"
        );
    }

    /**
     * Check whether a field has any validation constraints.
     * Returns true if the field has attributes or types that require validation.
     * @param {HTMLElement} field - The form field element.
     * @returns {boolean} True if the field has validation constraints.
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
     * Validate a single field and update the UI accordingly.
     * Shows an error or success state based on the validation result.
     * @param {HTMLElement} field - The form field element.
     * @returns {boolean} True if the field is valid, false otherwise.
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
     * Initialize client-side validation on all forms with the novalidate attribute.
     * Attaches blur, input, and submit event listeners to each validatable field.
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

                // Update validation on input for real-time feedback
                field.addEventListener("input", function () {
                    if (field.classList.contains("is-invalid")) {
                        validateAndShow(field);
                    }
                });
            });

            // Validate all fields on submit
            form.addEventListener("submit", function (e) {
                var currentFields = getValidatableFields(form);
                var isValid = true;
                var firstInvalid = null;

                currentFields.forEach(function (field) {
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
