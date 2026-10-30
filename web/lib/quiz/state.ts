import type { QuizAnswers } from "./types";

export const STORAGE_KEY = "agrolang.quiz.v1";

export const initialAnswers: QuizAnswers = {
  work_style: null,
  areas_bucket: null,
  producers_bucket: null,
  name: "",
  phone: "",
  organization: "",
  crop_choice: "",
  crop_custom: "",
  crop_stage: null,
  city: "",
  state: "",
  latitude: null,
  longitude: null,
  soil_type: null,
  concerns: [],
  email: "",
};

export type QuizAction =
  | { type: "patch"; patch: Partial<QuizAnswers> }
  | { type: "replace"; answers: QuizAnswers }
  | { type: "reset" };

export function quizReducer(state: QuizAnswers, action: QuizAction): QuizAnswers {
  switch (action.type) {
    case "patch":
      return { ...state, ...action.patch };
    case "replace":
      return action.answers;
    case "reset":
      return initialAnswers;
  }
}

type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;

function defaultStorage(): StorageLike | null {
  try {
    return typeof window === "undefined" ? null : window.localStorage;
  } catch {
    return null;
  }
}

const NUMERIC_KEYS = new Set(["latitude", "longitude"]);

function sameKind(key: string, initial: unknown, value: unknown): boolean {
  if (Array.isArray(initial)) return Array.isArray(value);
  if (NUMERIC_KEYS.has(key)) return value === null || typeof value === "number";
  if (initial === null) return value === null || typeof value === "string";
  return typeof value === typeof initial;
}

/** Reads saved answers, ignoring anything unknown or of the wrong kind. Never throws. */
export function loadAnswers(storage: StorageLike | null = defaultStorage()): QuizAnswers {
  if (!storage) return initialAnswers;
  try {
    const raw = storage.getItem(STORAGE_KEY);
    if (!raw) return initialAnswers;
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed === null) return initialAnswers;

    const merged: Record<string, unknown> = { ...initialAnswers };
    for (const key of Object.keys(initialAnswers)) {
      const value = (parsed as Record<string, unknown>)[key];
      if (value !== undefined && sameKind(key, (initialAnswers as Record<string, unknown>)[key], value)) {
        merged[key] = value;
      }
    }
    return merged as QuizAnswers;
  } catch {
    return initialAnswers;
  }
}

export function saveAnswers(answers: QuizAnswers, storage: StorageLike | null = defaultStorage()): void {
  if (!storage) return;
  try {
    storage.setItem(STORAGE_KEY, JSON.stringify(answers));
  } catch {
    // storage is optional: the quiz keeps working from memory
  }
}

export function clearAnswers(storage: StorageLike | null = defaultStorage()): void {
  if (!storage) return;
  try {
    storage.removeItem(STORAGE_KEY);
  } catch {
    // nothing to clean up if storage is unavailable
  }
}
