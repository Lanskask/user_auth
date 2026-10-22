// Regenerates docs/screenshots/*.png from a running server (default http://localhost:8080).
// Usage: BASE_URL=http://localhost:8080 node scripts/screenshots.mjs
import { chromium } from "@playwright/test";

const base = process.env.BASE_URL ?? "http://localhost:8080";
const out = new URL("../../docs/screenshots/", import.meta.url).pathname;
const browser = await chromium.launch();

async function shot(name, { scheme, viewport, run }) {
  const context = await browser.newContext({ colorScheme: scheme, viewport, deviceScaleFactor: 2 });
  const page = await context.newPage();
  await run(page);
  await page.screenshot({ path: `${out}${name}.png` });
  await context.close();
}

const desktop = { width: 1100, height: 720 };
const email = `ada.${Math.random().toString(36).slice(2, 6)}@example.com`;

await shot("login-light", {
  scheme: "light",
  viewport: { width: 1100, height: 560 },
  run: async (page) => {
    await page.goto(`${base}/login`);
    await page.getByRole("button", { name: "Sign in" }).click(); // shows the validation state
  },
});

await shot("register-dark", {
  scheme: "dark",
  viewport: { width: 1100, height: 620 },
  run: async (page) => {
    await page.goto(`${base}/register`);
    await page.getByLabel("Name").fill("Ada Lovelace");
    await page.getByLabel("Email").fill("ada@example");
    await page.getByLabel("Password").fill("short");
    await page.getByRole("button", { name: "Create account" }).click();
  },
});

await shot("dashboard-light", {
  scheme: "light",
  viewport: desktop,
  run: async (page) => {
    await page.goto(`${base}/register`);
    await page.getByLabel("Name").fill("Ada Lovelace");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password").fill("correct horse battery");
    await page.getByRole("button", { name: "Create account" }).click();
    await page.getByRole("heading", { name: "Hello, Ada Lovelace" }).waitFor();
  },
});

await browser.close();
