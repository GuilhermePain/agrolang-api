"use client";

import type { Option } from "@/lib/quiz/options";
import { optionClass } from "./styles";

function OptionBody({ label, hint, selected }: { label: string; hint?: string; selected: boolean }) {
  return (
    <>
      <span className="font-semibold">{label}</span>
      {hint && <span className={`mt-0.5 text-xs ${selected ? "text-paper/80" : "text-ink-soft"}`}>{hint}</span>}
    </>
  );
}

export function ChoiceList<T extends string>({
  options,
  value,
  onChange,
  label,
}: {
  options: Option<T>[];
  value: T | null;
  onChange: (value: T) => void;
  label?: string;
}) {
  return (
    <div role="radiogroup" aria-label={label} className="flex flex-col gap-3">
      {options.map((o) => {
        const selected = o.value === value;
        return (
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={selected}
            onClick={() => onChange(o.value)}
            className={optionClass(selected)}
          >
            <OptionBody label={o.label} hint={o.hint} selected={selected} />
          </button>
        );
      })}
    </div>
  );
}

export function MultiChoiceList<T extends string>({
  options,
  values,
  onChange,
  label,
}: {
  options: Option<T>[];
  values: T[];
  onChange: (values: T[]) => void;
  label?: string;
}) {
  function toggle(value: T) {
    onChange(values.includes(value) ? values.filter((v) => v !== value) : [...values, value]);
  }

  return (
    <div role="group" aria-label={label} className="flex flex-col gap-3">
      {options.map((o) => {
        const selected = values.includes(o.value);
        return (
          <button
            key={o.value}
            type="button"
            aria-pressed={selected}
            onClick={() => toggle(o.value)}
            className={optionClass(selected)}
          >
            <OptionBody label={o.label} hint={o.hint} selected={selected} />
          </button>
        );
      })}
    </div>
  );
}
