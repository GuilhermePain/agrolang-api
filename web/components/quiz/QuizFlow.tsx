"use client";

import { useEffect, useReducer, useState, type ReactNode } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  AREAS_OPTIONS,
  CONCERN_OPTIONS,
  PRODUCERS_OPTIONS,
  SOIL_OPTIONS,
  STYLE_OPTIONS,
} from "@/lib/quiz/options";
import { initialAnswers, loadAnswers, quizReducer, saveAnswers } from "@/lib/quiz/state";
import { firstIncompleteIndex, stepsFor, validateStep, type StepId } from "@/lib/quiz/steps";
import type { QuizAnswers } from "@/lib/quiz/types";
import AccountStep from "./AccountStep";
import { ChoiceList, MultiChoiceList } from "./ChoiceList";
import CropStep from "./CropStep";
import IdentityStep from "./IdentityStep";
import LocationStep from "./LocationStep";
import QuizShell from "./QuizShell";
import { primaryButtonClass } from "./styles";

const TITLES: Record<StepId, { title: string; hint?: string }> = {
  style: { title: "Qual dessas combina com você?" },
  areas: { title: "Quantas áreas de plantio você tem?" },
  identity: { title: "Como podemos te chamar?", hint: "Os avisos chegam pelo WhatsApp." },
  crop: { title: "O que você planta?" },
  location: { title: "Onde fica a sua terra?", hint: "Busque a cidade e ajuste o pino sobre a sua área." },
  soil: { title: "Como é a terra?" },
  concerns: {
    title: "O que mais te preocupa no clima?",
    hint: "Pode marcar mais de um. Se pular, avisamos de tudo.",
  },
  producers: { title: "Quantos produtores você atende?" },
  account: { title: "Crie o seu acesso", hint: "Com e-mail e senha você entra de novo quando quiser." },
};

function parseStep(raw: string | null): number {
  const n = Number(raw ?? "1");
  return Number.isInteger(n) && n >= 1 ? n - 1 : 0;
}

export default function QuizFlow() {
  const router = useRouter();
  const params = useSearchParams();
  const [answers, dispatch] = useReducer(quizReducer, initialAnswers);
  const [hydrated, setHydrated] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Restore saved answers once in the browser; storage is never read during render.
  useEffect(() => {
    dispatch({ type: "replace", answers: loadAnswers() });
    // eslint-disable-next-line react-hooks/set-state-in-effect -- restoring browser-only storage after mount
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (hydrated) saveAnswers(answers);
  }, [answers, hydrated]);

  const steps = stepsFor(answers.work_style);
  const index = Math.min(parseStep(params.get("passo")), steps.length - 1);
  const furthest = firstIncompleteIndex(steps, answers);

  // A URL can't skip ahead of unanswered steps (e.g. a reload or a pasted link).
  useEffect(() => {
    if (hydrated && index > furthest) router.replace(`/cadastro?passo=${furthest + 1}`);
  }, [hydrated, index, furthest, router]);

  if (!hydrated || index > furthest) {
    return (
      <QuizShell index={0} total={null} title="Carregando...">
        {null}
      </QuizShell>
    );
  }

  const stepId = steps[index];
  const { title, hint } = TITLES[stepId];
  const patch = (p: Partial<QuizAnswers>) => dispatch({ type: "patch", patch: p });

  function goTo(i: number) {
    setError(null);
    router.push(`/cadastro?passo=${i + 1}`);
  }

  function next() {
    const message = validateStep(stepId, answers);
    if (message) {
      setError(message);
      return;
    }
    goTo(index + 1);
  }

  // Single-choice steps advance on tap; the Continue button stays for returning visitors.
  function pick(p: Partial<QuizAnswers>) {
    patch(p);
    setError(null);
    window.setTimeout(() => goTo(index + 1), 180);
  }

  function renderStep(): ReactNode {
    switch (stepId) {
      case "style":
        return (
          <ChoiceList
            options={STYLE_OPTIONS}
            value={answers.work_style}
            onChange={(work_style) => pick({ work_style })}
          />
        );
      case "areas":
        return (
          <ChoiceList
            options={AREAS_OPTIONS}
            value={answers.areas_bucket}
            onChange={(areas_bucket) => pick({ areas_bucket })}
          />
        );
      case "producers":
        return (
          <ChoiceList
            options={PRODUCERS_OPTIONS}
            value={answers.producers_bucket}
            onChange={(producers_bucket) => pick({ producers_bucket })}
          />
        );
      case "soil":
        return (
          <ChoiceList
            options={SOIL_OPTIONS}
            value={answers.soil_type}
            onChange={(soil_type) => pick({ soil_type })}
          />
        );
      case "identity":
        return <IdentityStep answers={answers} technician={answers.work_style === "technician"} onChange={patch} />;
      case "crop":
        return <CropStep answers={answers} onChange={patch} />;
      case "location":
        return <LocationStep answers={answers} onChange={patch} />;
      case "concerns":
        return (
          <MultiChoiceList
            options={CONCERN_OPTIONS}
            values={answers.concerns}
            onChange={(concerns) => patch({ concerns })}
          />
        );
      case "account":
        return (
          <AccountStep
            answers={answers}
            onEmailChange={(email) => patch({ email })}
            onFixStep={(step) => {
              const target = steps.indexOf(step);
              if (target !== -1) goTo(target);
            }}
          />
        );
    }
  }

  const footer =
    stepId === "account" ? null : (
      <div className="flex flex-col gap-2">
        {error && (
          <p role="alert" className="font-mono text-sm text-risk-critical">
            {error}
          </p>
        )}
        <button type="button" onClick={next} className={primaryButtonClass}>
          Continuar
        </button>
        {index === 0 && (
          <p className="text-center font-mono text-xs text-ink-soft">
            Já tem conta?{" "}
            <Link href="/entrar" className="text-accent underline">
              Entrar
            </Link>
          </p>
        )}
      </div>
    );

  return (
    <QuizShell
      index={index}
      total={answers.work_style ? steps.length : null}
      title={title}
      hint={hint}
      onBack={index > 0 ? () => goTo(index - 1) : undefined}
      footer={footer}
    >
      {renderStep()}
    </QuizShell>
  );
}
