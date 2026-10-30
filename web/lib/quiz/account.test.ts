import { describe, expect, it } from "vitest";
import { validateAccount } from "./account";

describe("validateAccount", () => {
  it("accepts a valid email and password", () => {
    expect(validateAccount("jose@exemplo.com", "segredo123")).toEqual({});
  });

  it("rejects a malformed email", () => {
    expect(validateAccount("jose", "segredo123").email).toBe("Digite um e-mail válido.");
    expect(validateAccount("jose@exemplo", "segredo123").email).toBe("Digite um e-mail válido.");
  });

  it("rejects passwords that are too short or too long", () => {
    expect(validateAccount("jose@exemplo.com", "curta").password).toBe(
      "A senha precisa ter pelo menos 8 caracteres."
    );
    expect(validateAccount("jose@exemplo.com", "a".repeat(73)).password).toBe(
      "A senha pode ter no máximo 72 caracteres."
    );
  });
});
