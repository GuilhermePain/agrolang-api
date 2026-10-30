import { fieldClass } from "../FieldShell";

export const bigFieldClass = `${fieldClass} min-h-12 text-base`;

export function optionClass(selected: boolean): string {
  return [
    "flex min-h-14 w-full flex-col justify-center border-2 px-4 py-3 text-left font-mono text-sm transition active:translate-y-px",
    selected
      ? "border-ink bg-ink text-paper shadow-[var(--shadow-hard)]"
      : "border-ink bg-card text-ink hover:shadow-[var(--shadow-hard)]",
  ].join(" ");
}

export const primaryButtonClass =
  "min-h-12 w-full border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50";
