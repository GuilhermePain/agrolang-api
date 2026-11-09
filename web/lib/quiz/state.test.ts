import { describe, expect, it } from "vitest";
import {
  STORAGE_KEY,
  clearAnswers,
  initialAnswers,
  loadAnswers,
  quizReducer,
  saveAnswers,
} from "./state";

function fakeStorage(initial: Record<string, string> = {}) {
  const data = { ...initial };
  return {
    data,
    getItem: (key: string) => data[key] ?? null,
    setItem: (key: string, value: string) => {
      data[key] = value;
    },
    removeItem: (key: string) => {
      delete data[key];
    },
  };
}

describe("quizReducer", () => {
  it("merges a patch into the answers", () => {
    const next = quizReducer(initialAnswers, { type: "patch", patch: { name: "José" } });
    expect(next.name).toBe("José");
    expect(next.city).toBe("");
  });

  it("replaces and resets the answers", () => {
    const filled = { ...initialAnswers, name: "José" };
    expect(quizReducer(initialAnswers, { type: "replace", answers: filled })).toBe(filled);
    expect(quizReducer(filled, { type: "reset" })).toEqual(initialAnswers);
  });
});

describe("persistence", () => {
  it("round-trips answers through storage", () => {
    const storage = fakeStorage();
    const answers = { ...initialAnswers, name: "José", concerns: ["frost" as const] };
    saveAnswers(answers, storage);
    expect(loadAnswers(storage)).toEqual(answers);
  });

  it("returns the initial answers when storage is empty, null or corrupted", () => {
    expect(loadAnswers(fakeStorage())).toEqual(initialAnswers);
    expect(loadAnswers(null)).toEqual(initialAnswers);
    expect(loadAnswers(fakeStorage({ [STORAGE_KEY]: "{not json" }))).toEqual(initialAnswers);
  });

  it("drops unknown keys and values of the wrong kind", () => {
    const storage = fakeStorage({
      [STORAGE_KEY]: JSON.stringify({ name: "José", password: "oops", concerns: "frost", latitude: "x", city: 5 }),
    });
    const loaded = loadAnswers(storage);
    expect(loaded.name).toBe("José");
    expect(loaded.concerns).toEqual([]);
    expect(loaded.city).toBe("");
    expect(loaded.latitude).toBeNull();
    expect("password" in loaded).toBe(false);
  });

  it("keeps stored crops that are well-formed and drops the rest", () => {
    const storage = fakeStorage({
      [STORAGE_KEY]: JSON.stringify({
        crops: [
          { choice: "Milho", stage: "flowering" },
          { choice: "Café", stage: null },
          { choice: "Soja", stage: "bloom" },
          { choice: 5, stage: null },
          "Feijão",
          null,
        ],
      }),
    });
    expect(loadAnswers(storage).crops).toEqual([
      { choice: "Milho", stage: "flowering" },
      { choice: "Café", stage: null },
    ]);
  });

  it("clears the stored answers", () => {
    const storage = fakeStorage({ [STORAGE_KEY]: "{}" });
    clearAnswers(storage);
    expect(storage.data[STORAGE_KEY]).toBeUndefined();
  });

  it("does not throw when storage fails", () => {
    const broken = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
      removeItem: () => {
        throw new Error("blocked");
      },
    };
    expect(loadAnswers(broken)).toEqual(initialAnswers);
    expect(() => saveAnswers(initialAnswers, broken)).not.toThrow();
    expect(() => clearAnswers(broken)).not.toThrow();
  });
});
