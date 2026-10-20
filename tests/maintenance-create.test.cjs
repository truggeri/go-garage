const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");

function setup(validVehicle = true, validRecords = true) {
    const handlers = {};
    const buttons = [{ disabled: false }];
    const status = { textContent: "" };
    const cards = [0, 1, 2].map(index => ({
        index, removed: false,
        reportValidity: () => validRecords,
        remove() { this.removed = true; }
    }));
    const vehicle = { value: "vehicle-2", reportValidity: () => validVehicle };
    const submit = { addEventListener: (name, callback) => { handlers.click = callback; } };
    const container = {
        setAttribute() {},
        querySelector: () => status,
        querySelectorAll: selector => selector === "button" ? buttons : cards.filter(card => !card.removed),
        addEventListener: (name, callback) => { handlers[name] = callback; }
    };
    const requests = [];
    const document = {
        addEventListener: (name, callback) => callback(),
        getElementById: id => ({
            "bulk-maintenance-form": container,
            vehicle_id: vehicle,
            "add-record-btn": buttons[0],
            "records-container": { dispatchEvent() {} },
            "submit-records-btn": submit
        })[id]
    };
    vm.runInNewContext(fs.readFileSync(__dirname + "/../web/static/js/maintenance-create.js", "utf8"), {
        document, Event,
        htmx: { trigger: form => requests.push(form.index) },
        window: { setTimeout: callback => callback() }
    });
    return { handlers, cards, requests, vehicle, buttons, status };
}

test("selects the current vehicle URL and serializes requests, suppressing intermediate redirects", () => {
    const { handlers, cards, requests, vehicle, buttons } = setup();
    handlers.click();
    handlers.click();
    assert.deepEqual(requests, [0]);
    assert.equal(buttons[0].disabled, true);
    const config = { detail: {} };
    handlers["htmx:configRequest"](config);
    assert.equal(config.detail.path, "/api/v1/vehicles/vehicle-2/maintenance");
    let prevented = false;
    handlers["htmx:beforeOnLoad"]({ detail: { xhr: { status: 201 } }, preventDefault() { prevented = true; } });
    assert.equal(prevented, true);
    handlers["htmx:afterRequest"]({ detail: { elt: cards[0], xhr: { status: 201 } } });
    assert.equal(cards[0].removed, true);
    assert.deepEqual(requests, [0, 1]);
    handlers["htmx:afterRequest"]({ detail: { elt: {}, xhr: { status: 201 } } });
    assert.deepEqual(requests, [0, 1], "ignore the duplicate ancestor notification");
    handlers["htmx:afterRequest"]({ detail: { elt: cards[1], xhr: { status: 201 } } });
    assert.deepEqual(requests, [0, 1, 2]);
    prevented = false;
    handlers["htmx:beforeOnLoad"]({ detail: { xhr: { status: 201 } }, preventDefault() { prevented = true; } });
    assert.equal(prevented, false);
    handlers["htmx:afterRequest"]({ detail: { elt: cards[2], xhr: { status: 201 } } });
    assert.equal(vehicle.disabled, false);
});

test("stops on API or network failure and retries only unsaved records", () => {
    for (const statusCode of [400, 403, 500, 0]) {
        const { handlers, cards, requests, status, buttons } = setup();
        handlers.click();
        handlers["htmx:afterRequest"]({ detail: { elt: cards[0], xhr: { status: 201 } } });
        handlers["htmx:afterRequest"]({ detail: { elt: cards[1], xhr: { status: statusCode } } });
        assert.deepEqual(requests, [0, 1]);
        assert.match(status.textContent, /1 record\(s\) saved.*Only unsaved records/);
        assert.equal(buttons[0].disabled, false);
        handlers.click();
        assert.deepEqual(requests, [0, 1, 1]);
    }
});

test("validates the dropdown and all cards before saving any records", () => {
    for (const [vehicleValid, recordsValid] of [[false, true], [true, false]]) {
        const { handlers, requests } = setup(vehicleValid, recordsValid);
        handlers.click();
        assert.deepEqual(requests, []);
    }
});
