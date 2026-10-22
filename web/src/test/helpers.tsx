import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { App } from "../App";
import { AuthProvider } from "../auth/AuthContext";

export interface MockResponse {
  status: number;
  body?: unknown;
}

/** Replaces fetch with a router keyed by "METHOD /path"; unmatched calls fail the test loudly. */
export function mockApi(routes: Record<string, MockResponse | (() => MockResponse)>) {
  const calls: { key: string; body: unknown }[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const key = `${init?.method ?? "GET"} ${url}`;
      calls.push({ key, body: init?.body ? JSON.parse(String(init.body)) : undefined });
      const route = routes[key];
      if (!route) throw new Error(`unexpected request: ${key}`);
      const r = typeof route === "function" ? route() : route;
      return new Response(r.status === 204 ? null : JSON.stringify(r.body ?? {}), {
        status: r.status,
      });
    }),
  );
  return calls;
}

export const ann = {
  id: 1,
  email: "ann@example.com",
  name: "Ann",
  createdAt: "2026-01-02T03:04:05Z",
};

export function renderApp(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AuthProvider>
        <App />
      </AuthProvider>
    </MemoryRouter>,
  );
}
