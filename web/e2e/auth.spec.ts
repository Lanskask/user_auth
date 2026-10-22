import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page } from "@playwright/test";

const PASSWORD = "correct horse battery";

/** Each test gets its own account so tests stay independent and can run in parallel against one database. */
function newUser() {
  const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  return { name: "Ada Lovelace", email: `ada-${id}@example.com` };
}

async function register(page: Page, user: { name: string; email: string }) {
  await page.goto("/register");
  await page.getByLabel("Name").fill(user.name);
  await page.getByLabel("Email").fill(user.email);
  await page.getByLabel("Password").fill(PASSWORD);
  await page.getByRole("button", { name: "Create account" }).click();
  await expect(page.getByRole("heading", { name: `Hello, ${user.name}` })).toBeVisible();
}

test.describe("access control", () => {
  test("anonymous visitors are sent to sign-in", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();
  });

  test("unknown routes show a 404 page", async ({ page }) => {
    await page.goto("/definitely/not/here");
    await expect(page.getByRole("heading", { name: "Page not found" })).toBeVisible();
  });
});

test.describe("registration and session", () => {
  test("register, survive a reload, sign out, then be locked out", async ({ page }) => {
    const user = newUser();
    await register(page, user);
    await expect(page.getByText(user.email)).toBeVisible();

    await page.reload();
    await expect(page.getByRole("heading", { name: `Hello, ${user.name}` })).toBeVisible();

    await page.getByRole("button", { name: "Sign out" }).click();
    await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();

    await page.goto("/");
    await expect(page).toHaveURL(/\/login$/);
  });

  test("the session cookie is HttpOnly and invisible to page scripts", async ({
    page,
    context,
  }) => {
    await register(page, newUser());

    const cookie = (await context.cookies()).find((c) => c.name === "session");
    expect(cookie).toMatchObject({ httpOnly: true, sameSite: "Lax", path: "/" });
    expect(await page.evaluate(() => document.cookie)).toBe("");
  });

  test("a duplicate email is reported on the email field", async ({ page, browser }) => {
    const user = newUser();
    await register(page, user);

    const other = await browser.newContext();
    const second = await other.newPage();
    await second.goto("/register");
    await second.getByLabel("Name").fill("Someone Else");
    await second.getByLabel("Email").fill(user.email.toUpperCase());
    await second.getByLabel("Password").fill(PASSWORD);
    await second.getByRole("button", { name: "Create account" }).click();

    await expect(second.getByLabel("Email")).toHaveAttribute("aria-invalid", "true");
    await expect(second.getByText("already registered", { exact: true })).toBeVisible();
    await other.close();
  });
});

test.describe("sign-in", () => {
  test("rejects a wrong password with a generic message, accepts the right one", async ({
    page,
  }) => {
    const user = newUser();
    await register(page, user);
    await page.getByRole("button", { name: "Sign out" }).click();

    await page.getByLabel("Email").fill(user.email);
    await page.getByLabel("Password").fill("not the password");
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page.getByRole("alert")).toHaveText("invalid email or password");

    await page.getByLabel("Password").fill(PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page.getByRole("heading", { name: `Hello, ${user.name}` })).toBeVisible();
  });

  test("returns to the page that required sign-in", async ({ page }) => {
    const user = newUser();
    await register(page, user);
    await page.getByRole("button", { name: "Sign out" }).click();

    await page.goto("/");
    await page.getByLabel("Email").fill(user.email);
    await page.getByLabel("Password").fill(PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page).toHaveURL(/\/$/);
  });
});

test.describe("forms", () => {
  test("client validation shows errors and moves focus to the first invalid field", async ({
    page,
  }) => {
    await page.goto("/register");
    await page.getByRole("button", { name: "Create account" }).click();

    await expect(page.getByLabel("Name")).toBeFocused();
    await expect(page.getByLabel("Name")).toHaveAttribute("aria-invalid", "true");
    await expect(page.getByLabel("Email")).toHaveAccessibleDescription(
      "Enter a valid email address.",
    );
    await expect(page.getByLabel("Password")).toHaveAccessibleDescription(
      "Use 8 to 72 characters.",
    );
  });

  test("the password can be revealed and hidden again", async ({ page }) => {
    await page.goto("/login");
    const password = page.getByLabel("Password");
    await password.fill("secret-value");
    await expect(password).toHaveAttribute("type", "password");
    await page.getByRole("button", { name: "Show" }).click();
    await expect(password).toHaveAttribute("type", "text");
    await page.getByRole("button", { name: "Hide" }).click();
    await expect(password).toHaveAttribute("type", "password");
  });

  test("works with the keyboard alone", async ({ page }) => {
    await page.goto("/login");
    await page.getByLabel("Email").focus();
    await page.keyboard.type("nobody@example.com");
    await page.keyboard.press("Tab");
    await page.keyboard.type("whatever-password");
    await page.keyboard.press("Enter");
    await expect(page.getByRole("alert")).toBeVisible();
  });
});

test.describe("quality", () => {
  for (const path of ["/login", "/register"]) {
    test(`${path} has no detectable accessibility violations`, async ({ page }) => {
      await page.goto(path);
      await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
      const results = await new AxeBuilder({ page })
        .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
        .analyze();
      expect(results.violations).toEqual([]);
    });
  }

  test("the dashboard has no accessibility violations", async ({ page }) => {
    await register(page, newUser());
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test("no horizontal scrolling on a 320px phone", async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 640 });
    await register(page, newUser());
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - window.innerWidth,
    );
    expect(overflow).toBeLessThanOrEqual(0);
  });

  test("API responses are not cached and security headers are present", async ({ request }) => {
    const res = await request.get("/api/v1/me");
    expect(res.status()).toBe(401);
    expect(res.headers()["cache-control"]).toBe("no-store");
    expect(res.headers()["x-content-type-options"]).toBe("nosniff");
    expect(res.headers()["content-security-policy"]).toContain("frame-ancestors 'none'");
  });

  test("a cross-origin POST is refused", async ({ request }) => {
    const res = await request.post("/api/v1/auth/logout", {
      headers: { Origin: "https://evil.example" },
    });
    expect(res.status()).toBe(403);
  });
});
