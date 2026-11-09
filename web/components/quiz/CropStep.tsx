"use client";

import { CROP_OPTIONS, CROP_OTHER, STAGE_OPTIONS } from "@/lib/quiz/options";
import { cropList } from "@/lib/quiz/steps";
import type { CropStage, QuizAnswers } from "@/lib/quiz/types";
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
  const names = cropList(answers);

  function toggle(choice: string) {
    const picked = answers.crops.some((c) => c.choice === choice);
    onChange({
      crops: picked
        ? answers.crops.filter((c) => c.choice !== choice)
        : [...answers.crops, { choice, stage: null }],
    });
  }

  function setStage(choice: string, stage: CropStage) {
    onChange({ crops: answers.crops.map((c) => (c.choice === choice ? { ...c, stage } : c)) });
  }

  return (
    <>
      <div role="group" aria-label="Culturas" className="grid grid-cols-2 gap-3">
        {choices.map((choice) => {
          const selected = answers.crops.some((c) => c.choice === choice);
          return (
            <button
              key={choice}
              type="button"
              aria-pressed={selected}
              onClick={() => toggle(choice)}
              className={optionClass(selected)}
            >
              <span className="font-semibold">{choice === CROP_OTHER ? "Outra" : choice}</span>
            </button>
          );
        })}
      </div>

      {answers.crops.some((c) => c.choice === CROP_OTHER) && (
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

      {answers.crops.map((entry, i) => {
        const name = names[i].crop || "outra cultura";
        return (
          <div key={entry.choice} className="flex flex-col gap-3">
            <p className="mt-2 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft">
              Em que fase está: {name}?
            </p>
            <ChoiceList
              label={`Fase de ${name}`}
              options={STAGE_OPTIONS}
              value={entry.stage}
              onChange={(stage) => setStage(entry.choice, stage)}
            />
          </div>
        );
      })}
    </>
  );
}
