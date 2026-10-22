import { Link } from "react-router";

export function NotFoundPage() {
  return (
    <section className="card">
      <h1>Page not found</h1>
      <p className="alt">
        That address does not exist. <Link to="/">Go to the start page</Link>
      </p>
    </section>
  );
}
