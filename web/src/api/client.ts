export interface User {
  id: number;
  email: string;
  name: string;
  createdAt: string;
}

export type FieldErrors = Partial<Record<"email" | "name" | "password", string>>;

/** An error response from the API, carrying the server's machine-readable code and per-field messages. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields: FieldErrors;

  constructor(status: number, code: string, message: string, fields: FieldErrors = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`/api/v1${path}`, {
      ...init,
      credentials: "same-origin",
      headers: { Accept: "application/json", "Content-Type": "application/json" },
    });
  } catch {
    throw new ApiError(
      0,
      "network",
      "Cannot reach the server. Check your connection and try again.",
    );
  }

  if (res.status === 204) return undefined as T;

  const body: unknown = await res.json().catch(() => null);
  if (!res.ok) {
    const e = (body as { error?: { code?: string; message?: string; fields?: FieldErrors } } | null)
      ?.error;
    throw new ApiError(
      res.status,
      e?.code ?? "unknown",
      e?.message ?? "Something went wrong.",
      e?.fields,
    );
  }
  return body as T;
}

const post = (path: string, data?: unknown) =>
  request<{ user: User }>(path, { method: "POST", body: data ? JSON.stringify(data) : undefined });

export const api = {
  me: () => request<{ user: User }>("/me").then((r) => r.user),
  login: (email: string, password: string) =>
    post("/auth/login", { email, password }).then((r) => r.user),
  register: (email: string, name: string, password: string) =>
    post("/auth/register", { email, name, password }).then((r) => r.user),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
};
