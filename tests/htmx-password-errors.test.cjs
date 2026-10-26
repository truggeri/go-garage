const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");

test("shared error rendering never reflects submitted passwords or detail values", () => {
    for (const file of ["htmx-conventions.js", "htmx-conventions-minified.js"]) {
        const handlers = {};
        const error = { textContent: "" };
        const general = { textContent: "" };
        const control = {
            value: "SubmittedPassword123",
            setAttribute() {},
            removeAttribute() {},
            focus() {}
        };
        const form = {
            closest: () => form,
            querySelectorAll: selector => selector === "[data-field]" ? [error] :
                selector === '[aria-invalid="true"]' ? [control] : [general],
            querySelector: selector => selector.startsWith("[data-field=") ? error :
                selector.startsWith("[name=") ? control : general
        };
        vm.runInNewContext(fs.readFileSync(__dirname + "/../web/static/js/" + file, "utf8"), {
            document: {
                readyState: "complete",
                addEventListener: (name, callback) => { handlers[name] = callback; }
            },
            window: {},
            CSS: { escape: value => value }
        });
        handlers["htmx:responseError"]({
            detail: {
                elt: form,
                parameters: { new_password: control.value },
                xhr: { responseText: JSON.stringify({
                    success: false,
                    error: {
                        message: "Validation failed",
                        details: [{ field: "new_password", message: "Password is too weak", value: control.value }]
                    }
                }) }
            }
        });
        assert.equal(error.textContent, "Password is too weak");
        assert.equal(general.textContent, "");
        assert.equal(control.value, "SubmittedPassword123");
        handlers["htmx:sendError"]({ detail: { elt: form, parameters: { new_password: control.value } } });
        assert.equal(error.textContent, "");
        assert.match(general.textContent, /Unable to reach the server/);
        assert.ok(!general.textContent.includes(control.value));
    }
});
