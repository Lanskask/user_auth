import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api, ApiError, type User } from "../api/client";

type AuthState =
  { status: "loading" } | { status: "anonymous" } | { status: "authenticated"; user: User };

interface AuthContextValue {
  state: AuthState;
  login(email: string, password: string): Promise<void>;
  register(email: string, name: string, password: string): Promise<void>;
  logout(): Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: "loading" });

  // Ask the server whether the session cookie is still valid; the cookie itself is HttpOnly and unreadable here.
  useEffect(() => {
    let cancelled = false;
    api
      .me()
      .then((user) => !cancelled && setState({ status: "authenticated", user }))
      .catch((e: unknown) => {
        if (cancelled) return;
        if (e instanceof ApiError && e.status !== 401) console.error(e);
        setState({ status: "anonymous" });
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const user = await api.login(email, password);
    setState({ status: "authenticated", user });
  }, []);

  const register = useCallback(async (email: string, name: string, password: string) => {
    await api.register(email, name, password);
    const user = await api.login(email, password);
    setState({ status: "authenticated", user });
  }, []);

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } finally {
      setState({ status: "anonymous" });
    }
  }, []);

  const value = useMemo(
    () => ({ state, login, register, logout }),
    [state, login, register, logout],
  );
  return <AuthContext value={value}>{children}</AuthContext>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}
