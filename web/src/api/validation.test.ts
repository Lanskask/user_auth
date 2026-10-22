import { validateLogin, validatePassword, validateRegister } from "./validation";

describe("validateRegister", () => {
  it("accepts valid input", () => {
    expect(
      validateRegister({ email: "a@example.com", name: "Ann", password: "password1" }),
    ).toEqual({});
  });

  it("reports every invalid field", () => {
    const errors = validateRegister({ email: "nope", name: "  ", password: "short" });
    expect(Object.keys(errors).sort()).toEqual(["email", "name", "password"]);
  });
});

describe("validatePassword", () => {
  it("counts bytes, not characters", () => {
    expect(validatePassword("é".repeat(36))).toBeUndefined(); // 72 bytes
    expect(validatePassword("é".repeat(37))).toBeDefined(); // 74 bytes
  });
});

describe("validateLogin", () => {
  it("requires a password but not its length", () => {
    expect(validateLogin({ email: "a@example.com", password: "x" })).toEqual({});
    expect(validateLogin({ email: "a@example.com", password: "" }).password).toBeDefined();
  });
});
