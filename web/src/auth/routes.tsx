import { Navigate, Outlet, useLocation } from "react-router";
import { useAuth } from "./AuthContext";

function Splash() {
  return (
    <p className="splash" role="status">
      Loading…
    </p>
  );
}

/** Renders child routes only for signed-in users; everyone else goes to /login and comes back afterwards. */
export function RequireAuth() {
  const { state } = useAuth();
  const location = useLocation();
  if (state.status === "loading") return <Splash />;
  if (state.status === "anonymous")
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  return <Outlet />;
}

/** Login and register are for visitors only; signed-in users are sent home. */
export function GuestOnly() {
  const { state } = useAuth();
  if (state.status === "loading") return <Splash />;
  if (state.status === "authenticated") return <Navigate to="/" replace />;
  return <Outlet />;
}
