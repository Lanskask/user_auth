import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { ApiError, type FieldErrors } from "../api/client";
import { validateRegister } from "../api/validation";
import { useAuth } from "../auth/AuthContext";
import { Field } from "../components/Field";
import { useFocusFirstInvalid } from "./useFormFocus";

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [values, setValues] = useState({ name: "", email: "", password: "" });
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string>();
  const [pending, setPending] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const formRef = useFocusFirstInvalid(attempt);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(undefined);
    const found = validateRegister(values);
    setErrors(found);
    if (Object.keys(found).length > 0) {
      setAttempt((n) => n + 1);
      return;
    }

    setPending(true);
    try {
      await register(values.email.trim(), values.name.trim(), values.password);
      void navigate("/", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        // The server may know something the browser cannot, e.g. that the email is taken.
        setErrors(err.fields);
        setFormError(Object.keys(err.fields).length > 0 ? undefined : err.message);
        setAttempt((n) => n + 1);
      } else {
        setFormError("Something went wrong. Try again.");
      }
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="card">
      <h1>Create account</h1>
      <form ref={formRef} onSubmit={(e) => void onSubmit(e)} noValidate>
        {formError && (
          <p className="alert" role="alert">
            {formError}
          </p>
        )}
        <Field
          label="Name"
          autoComplete="name"
          value={values.name}
          error={errors.name}
          onChange={(e) => setValues({ ...values, name: e.target.value })}
        />
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
          autoComplete="new-password"
          hint="8 to 72 characters. A passphrase works well."
          value={values.password}
          error={errors.password}
          onChange={(e) => setValues({ ...values, password: e.target.value })}
        />
        <button type="submit" className="primary" disabled={pending}>
          {pending ? "Creating account…" : "Create account"}
        </button>
      </form>
      <p className="alt">
        Already registered? <Link to="/login">Sign in</Link>
      </p>
    </section>
  );
}
