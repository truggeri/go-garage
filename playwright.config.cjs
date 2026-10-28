const { randomBytes } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { defineConfig } = require("@playwright/test");

const port = process.env.E2E_PORT || "4173";
const baseURL = `http://127.0.0.1:${port}`;
const databaseDirectory = fs.mkdtempSync(path.join(os.tmpdir(), "go-garage-e2e-"));

process.once("exit", () => {
    fs.rmSync(databaseDirectory, { recursive: true, force: true });
});

module.exports = defineConfig({
    testDir: "./tests/e2e",
    testMatch: "**/*.spec.cjs",
    fullyParallel: false,
    workers: 1,
    reporter: "list",
    use: {
        baseURL,
        browserName: "chromium",
        trace: "retain-on-failure",
        screenshot: "only-on-failure"
    },
    webServer: {
        command: "go run ./cmd/server",
        url: `${baseURL}/health`,
        reuseExistingServer: false,
        timeout: 60_000,
        env: {
            ...process.env,
            SERVER_HOST: "127.0.0.1",
            APP_PORT: port,
            DB_PATH: path.join(databaseDirectory, "test.db"),
            JWT_SECRET: randomBytes(32).toString("hex"),
            CSRF_SECRET: randomBytes(32).toString("hex"),
            ENVIRONMENT: "development",
            LOG_LEVEL: "error"
        }
    }
});
