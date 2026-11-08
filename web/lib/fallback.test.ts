import { describe, expect, it } from "vitest";
import { redirect } from "next/navigation";
import { orFallback } from "./fallback";

function redirectError(): unknown {
  try {
    redirect("/sair");
  } catch (err) {
    return err;
  }
}

describe("orFallback", () => {
  it("returns the resolved value", async () => {
    await expect(orFallback(Promise.resolve([1]), [])).resolves.toEqual([1]);
  });

  it("returns the fallback on ordinary errors", async () => {
    await expect(orFallback(Promise.reject(new Error("boom")), [])).resolves.toEqual([]);
  });

  it("rethrows session redirects", async () => {
    const err = redirectError();
    await expect(orFallback(Promise.reject(err), [])).rejects.toBe(err);
  });
});
