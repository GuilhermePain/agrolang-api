import { describe, expect, it } from "vitest";
import { isFullscreenPath } from "./chrome";

describe("isFullscreenPath", () => {
  it("matches the login, quiz, logout and welcome routes", () => {
    for (const p of ["/entrar", "/cadastro", "/sair", "/bem-vindo", "/cadastro/x"]) {
      expect(isFullscreenPath(p)).toBe(true);
    }
  });

  it("keeps the nav on app routes and lookalike prefixes", () => {
    for (const p of ["/", "/producers", "/properties/1", "/notifications", "/entrarx"]) {
      expect(isFullscreenPath(p)).toBe(false);
    }
  });

  it("treats a missing pathname as an app route", () => {
    expect(isFullscreenPath(null)).toBe(false);
  });
});
