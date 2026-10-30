"use client";

import { CROP_OPTIONS, CROP_OTHER, STAGE_OPTIONS } from "@/lib/quiz/options";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { ChoiceList } from "./ChoiceList";
import { bigFieldClass, optionClass } from "./styles";

export default function CropStep({
  answers,
  onChange,
}: {
  answers: QuizAnswers;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  const choices = [...CROP_OPTIONS, CROP_OTHER] as string[];

  return (
    <>
      <div role="radiogroup" aria-label="Cultura" className="grid grid-cols-2 gap-3">
        {choices.map((choice) => {
          const selected = answers.crop_choice === choice;
          return (
            <button
              key={choice}
              type="button"
              role="radio"
              aria-checked={selected}
              onClick={() => onChange({ crop_choice: choice })}
              className={optionClass(selected)}
            >
              <span className="font-semibold">{choice === CROP_OTHER ? "Outra" : choice}</span>
            </button>
          );
        })}
      </div>

      {answers.crop_choice === CROP_OTHER && (
        <FieldShell label="Qual cultura?" htmlFor="crop_custom">
          <input
            id="crop_custom"
            className={bigFieldClass}
            value={answers.crop_custom}
            onChange={(e) => onChange({ crop_custom: e.target.value })}
            placeholder="Ex: Mandioca"
            autoFocus
          />
        </FieldShell>
      )}

      <p className="mt-2 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft">
        Em que fase está?
      </p>
      <ChoiceList
        options={STAGE_OPTIONS}
        value={answers.crop_stage}
        onChange={(crop_stage) => onChange({ crop_stage })}
      />
    </>
  );
}
