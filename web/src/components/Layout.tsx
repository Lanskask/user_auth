import { Link, Outlet } from "react-router";
import { useAuth } from "../auth/AuthContext";

export function Layout() {
  const { state, logout } = useAuth();

  return (
    <>
      <a className="skip" href="#main">
        Skip to content
      </a>
      <header className="site-header">
        <Link to="/" className="brand">
          user_auth
        </Link>
        <nav aria-label="Account">
          {state.status === "authenticated" && (
            <>
              <span className="who">{state.user.name}</span>
              <button type="button" className="link" onClick={() => void logout()}>
                Sign out
              </button>
            </>
          )}
          {state.status === "anonymous" && (
            <>
              <Link to="/login">Sign in</Link>
              <Link to="/register">Create account</Link>
            </>
          )}
        </nav>
      </header>
      <main id="main" tabIndex={-1}>
        <Outlet />
      </main>
    </>
  );
}
