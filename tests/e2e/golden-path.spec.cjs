const { test, expect } = require("@playwright/test");

test("register, log in, and add a vehicle", async ({ page }) => {
    const username = `browser-smoke-${Date.now()}`;
    const email = `${username}@example.com`;
    const password = "GarageTest123";

    await page.goto("/register");
    await page.getByLabel("Username").fill(username);
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password", { exact: true }).fill(password);
    await page.getByLabel("Confirm Password").fill(password);
    await page.getByRole("button", { name: "Create Account" }).click();

    await expect(page).toHaveURL(/\/login\?registered=true$/);
    await expect(page.getByText("Account created successfully. Please log in.")).toBeVisible();
    await page.getByLabel("Email or Username").fill(username);
    await page.getByLabel("Password", { exact: true }).fill(password);
    await page.getByRole("button", { name: "Log In" }).click();

    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole("heading", { name: `Welcome back, ${username}` })).toBeVisible();
    await page.getByRole("link", { name: "Add Your First Vehicle" }).click();
    await page.getByLabel("Make").fill("Toyota");
    await page.getByLabel("Model").fill("Camry");
    await page.getByLabel("Year").fill("2023");
    await page.getByLabel("VIN").fill("1HGCM82633A004352");
    await page.getByRole("button", { name: "Add Vehicle" }).click();

    await expect(page).toHaveURL(/\/vehicles\/[^?]+\?added=true$/);
    await expect(page.getByRole("heading", { name: "2023 Toyota Camry" })).toBeVisible();
});
