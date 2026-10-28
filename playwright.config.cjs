const { randomBytes } = require("node:crypto");
const os = require("node:os");
const path = require("node:path");
const { defineConfig } = require("@playwright/test");

const port = process.env.E2E_PORT || "4173";
const baseURL = `http://127.0.0.1:${port}`;

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
            DB_PATH: path.join(os.tmpdir(), `go-garage-e2e-${randomBytes(16).toString("hex")}.db`),
            JWT_SECRET: randomBytes(32).toString("hex"),
            CSRF_SECRET: randomBytes(32).toString("hex"),
            ENVIRONMENT: "development",
            LOG_LEVEL: "error"
        }
    }
});
