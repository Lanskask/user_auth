import type { FieldErrors } from "./client";

// These rules mirror internal/auth/service.go so users get instant feedback; the server remains the authority.
const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function validateEmail(email: string): string | undefined {
  const v = email.trim();
  return EMAIL.test(v) && v.length <= 254 ? undefined : "Enter a valid email address.";
}

export function validatePassword(password: string): string | undefined {
  // bcrypt only reads 72 bytes, so the limit is in bytes, not characters.
  const bytes = new TextEncoder().encode(password).length;
  return bytes >= 8 && bytes <= 72 ? undefined : "Use 8 to 72 characters.";
}

export function validateName(name: string): string | undefined {
  const n = [...name.trim()].length;
  return n >= 1 && n <= 100 ? undefined : "Enter your name (up to 100 characters).";
}

export function validateRegister(v: {
  email: string;
  name: string;
  password: string;
}): FieldErrors {
  const errors: FieldErrors = {};
  const email = validateEmail(v.email);
  const name = validateName(v.name);
  const password = validatePassword(v.password);
  if (email) errors.email = email;
  if (name) errors.name = name;
  if (password) errors.password = password;
  return errors;
}

export function validateLogin(v: { email: string; password: string }): FieldErrors {
  const errors: FieldErrors = {};
  const email = validateEmail(v.email);
  if (email) errors.email = email;
  if (!v.password) errors.password = "Enter your password.";
  return errors;
}
