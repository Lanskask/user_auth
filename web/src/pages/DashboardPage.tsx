import { useAuth } from "../auth/AuthContext";

const protections = [
  [
    "Passwords",
    "Stored only as bcrypt hashes. They are never returned by the API or written to logs.",
  ],
  [
    "Session cookie",
    "HttpOnly, SameSite=Lax, and Secure over HTTPS, so page scripts cannot read it.",
  ],
  [
    "Session tokens",
    "256 random bits. The database keeps only a SHA-256 hash, so a leaked table cannot be replayed.",
  ],
  [
    "Brute force",
    "Sign-in and registration are rate limited per client, and unknown emails cost the same time as wrong passwords.",
  ],
] as const;

export function DashboardPage() {
  const { state } = useAuth();
  if (state.status !== "authenticated") return null;
  const { user } = state;
  const since = new Date(user.createdAt).toLocaleDateString(undefined, { dateStyle: "long" });

  return (
    <div className="stack">
      <section className="card wide">
        <h1>Hello, {user.name}</h1>
        <p className="lede">You are signed in. This page is only served to a valid session.</p>
        <dl className="profile">
          <div>
            <dt>Email</dt>
            <dd>{user.email}</dd>
          </div>
          <div>
            <dt>Member since</dt>
            <dd>{since}</dd>
          </div>
          <div>
            <dt>Account ID</dt>
            <dd>{user.id}</dd>
          </div>
        </dl>
      </section>
      <section className="card wide" aria-labelledby="protect">
        <h2 id="protect">How this session is protected</h2>
        <dl className="facts">
          {protections.map(([term, text]) => (
            <div key={term}>
              <dt>{term}</dt>
              <dd>{text}</dd>
            </div>
          ))}
        </dl>
      </section>
    </div>
  );
}
