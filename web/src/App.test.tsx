import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach } from "vitest";
import { ann, mockApi, renderApp } from "./test/helpers";

const unauthenticated = {
  status: 401,
  body: { error: { code: "unauthenticated", message: "sign in" } },
};

afterEach(() => vi.unstubAllGlobals());

describe("routing", () => {
  it("sends anonymous visitors from / to the sign-in page", async () => {
    mockApi({ "GET /api/v1/me": unauthenticated });
    renderApp("/");
    expect(await screen.findByRole("heading", { name: "Sign in" })).toBeInTheDocument();
  });

  it("shows the dashboard for a valid session and signs out", async () => {
    const user = userEvent.setup();
    mockApi({
      "GET /api/v1/me": { status: 200, body: { user: ann } },
      "POST /api/v1/auth/logout": { status: 204 },
    });
    renderApp("/");
    expect(await screen.findByRole("heading", { name: "Hello, Ann" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Sign out" }));
    expect(await screen.findByRole("heading", { name: "Sign in" })).toBeInTheDocument();
  });

  it("redirects signed-in users away from /login", async () => {
    mockApi({ "GET /api/v1/me": { status: 200, body: { user: ann } } });
    renderApp("/login");
    expect(await screen.findByRole("heading", { name: "Hello, Ann" })).toBeInTheDocument();
  });

  it("renders a 404 page for unknown routes", async () => {
    mockApi({ "GET /api/v1/me": unauthenticated });
    renderApp("/nope");
    expect(await screen.findByRole("heading", { name: "Page not found" })).toBeInTheDocument();
  });
});

describe("login form", () => {
  it("shows field errors and focuses the first invalid field without calling the API", async () => {
    const user = userEvent.setup();
    const calls = mockApi({ "GET /api/v1/me": unauthenticated });
    renderApp("/login");
    await screen.findByRole("heading", { name: "Sign in" });

    await user.click(screen.getByRole("button", { name: "Sign in" }));

    const email = screen.getByLabelText("Email");
    expect(email).toHaveAttribute("aria-invalid", "true");
    expect(email).toHaveAccessibleDescription("Enter a valid email address.");
    await waitFor(() => expect(email).toHaveFocus());
    expect(calls.map((c) => c.key)).toEqual(["GET /api/v1/me"]);
  });

  it("announces a server rejection", async () => {
    const user = userEvent.setup();
    mockApi({
      "GET /api/v1/me": unauthenticated,
      "POST /api/v1/auth/login": {
        status: 401,
        body: { error: { code: "invalid_credentials", message: "invalid email or password" } },
      },
    });
    renderApp("/login");
    await user.type(await screen.findByLabelText("Email"), "ann@example.com");
    await user.type(screen.getByLabelText("Password"), "wrong-password");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("invalid email or password");
  });

  it("signs in and lands on the dashboard", async () => {
    const user = userEvent.setup();
    const calls = mockApi({
      "GET /api/v1/me": unauthenticated,
      "POST /api/v1/auth/login": { status: 200, body: { user: ann } },
    });
    renderApp("/login");
    await user.type(await screen.findByLabelText("Email"), " ann@example.com ");
    await user.type(screen.getByLabelText("Password"), "password1");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("heading", { name: "Hello, Ann" })).toBeInTheDocument();
    expect(calls.at(-1)?.body).toEqual({ email: "ann@example.com", password: "password1" });
  });

  it("toggles password visibility", async () => {
    const user = userEvent.setup();
    mockApi({ "GET /api/v1/me": unauthenticated });
    renderApp("/login");
    const password = await screen.findByLabelText("Password");
    expect(password).toHaveAttribute("type", "password");
    await user.click(screen.getByRole("button", { name: "Show" }));
    expect(password).toHaveAttribute("type", "text");
  });
});

describe("register form", () => {
  it("maps a duplicate-email response onto the email field", async () => {
    const user = userEvent.setup();
    mockApi({
      "GET /api/v1/me": unauthenticated,
      "POST /api/v1/auth/register": {
        status: 409,
        body: {
          error: {
            code: "email_taken",
            message: "exists",
            fields: { email: "already registered" },
          },
        },
      },
    });
    renderApp("/register");
    await user.type(await screen.findByLabelText("Name"), "Ann");
    await user.type(screen.getByLabelText("Email"), "ann@example.com");
    await user.type(screen.getByLabelText("Password"), "password1");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(await screen.findByText("already registered")).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toHaveAttribute("aria-invalid", "true");
  });

  it("registers, signs in and lands on the dashboard", async () => {
    const user = userEvent.setup();
    const calls = mockApi({
      "GET /api/v1/me": unauthenticated,
      "POST /api/v1/auth/register": { status: 201, body: { user: ann } },
      "POST /api/v1/auth/login": { status: 200, body: { user: ann } },
    });
    renderApp("/register");
    await user.type(await screen.findByLabelText("Name"), "Ann");
    await user.type(screen.getByLabelText("Email"), "ann@example.com");
    await user.type(screen.getByLabelText("Password"), "password1");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(await screen.findByRole("heading", { name: "Hello, Ann" })).toBeInTheDocument();
    expect(calls.map((c) => c.key).slice(1)).toEqual([
      "POST /api/v1/auth/register",
      "POST /api/v1/auth/login",
    ]);
  });
});
