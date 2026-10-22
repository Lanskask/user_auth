import { useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { ApiError, type FieldErrors } from "../api/client";
import { validateLogin } from "../api/validation";
import { useAuth } from "../auth/AuthContext";
import { Field } from "../components/Field";
import { useFocusFirstInvalid } from "./useFormFocus";

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from ?? "/";

  const [values, setValues] = useState({ email: "", password: "" });
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string>();
  const [pending, setPending] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const formRef = useFocusFirstInvalid(attempt);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(undefined);
    const found = validateLogin(values);
    setErrors(found);
    if (Object.keys(found).length > 0) {
      setAttempt((n) => n + 1);
      return;
    }

    setPending(true);
    try {
      await login(values.email.trim(), values.password);
      void navigate(from, { replace: true });
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="card">
      <h1>Sign in</h1>
      <form ref={formRef} onSubmit={(e) => void onSubmit(e)} noValidate>
        {formError && (
          <p className="alert" role="alert">
            {formError}
          </p>
        )}
        <Field
          label="Email"
          type="email"
          autoComplete="email"
          value={values.email}
          error={errors.email}
          onChange={(e) => setValues({ ...values, email: e.target.value })}
        />
        <Field
          label="Password"
          type="password"
          autoComplete="current-password"
          value={values.password}
          error={errors.password}
          onChange={(e) => setValues({ ...values, password: e.target.value })}
        />
        <button type="submit" className="primary" disabled={pending}>
          {pending ? "Signing in…" : "Sign in"}
        </button>
      </form>
      <p className="alt">
        New here? <Link to="/register">Create an account</Link>
      </p>
    </section>
  );
}
