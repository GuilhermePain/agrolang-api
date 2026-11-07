"use client";

import { useActionState, useState, type FormEvent } from "react";
import Link from "next/link";
import { signupAction } from "@/app/actions/auth";
import { validateAccount } from "@/lib/quiz/account";
import { buildSignupPayload } from "@/lib/quiz/payload";
import { stepForField, type StepId } from "@/lib/quiz/steps";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { bigFieldClass, primaryButtonClass } from "./styles";

export default function AccountStep({
  answers,
  onEmailChange,
  onFixStep,
}: {
  answers: QuizAnswers;
  onEmailChange: (email: string) => void;
  onFixStep: (step: StepId) => void;
}) {
  const [state, formAction, pending] = useActionState(signupAction, undefined);
  const [password, setPassword] = useState("");
  const [showErrors, setShowErrors] = useState(false);

  const accountErrors = validateAccount(answers.email, password);
  let payload = "";
  try {
    payload = JSON.stringify(buildSignupPayload(answers, password));
  } catch {
    payload = ""; // earlier steps are incomplete; the guard in QuizFlow normally prevents this
  }
  const invalid = Object.keys(accountErrors).length > 0 || payload === "";

  const fieldMessages = state?.fields ? Object.entries(state.fields) : [];
  const firstBadStep =
    fieldMessages.map(([field]) => stepForField(field)).find((s): s is StepId => s !== null && s !== "account") ??
    null;

  function handleSubmit(e: FormEvent) {
    if (invalid) {
      e.preventDefault();
      setShowErrors(true);
    }
  }

  return (
    <form action={formAction} onSubmit={handleSubmit} className="flex flex-col gap-5">
      <input type="hidden" name="payload" value={payload} />

      <FieldShell label="E-mail" htmlFor="email" error={showErrors ? accountErrors.email : undefined}>
        <input
          id="email"
          type="email"
          inputMode="email"
          autoComplete="email"
          className={bigFieldClass}
          value={answers.email}
          onChange={(e) => onEmailChange(e.target.value)}
          placeholder="voce@exemplo.com"
        />
      </FieldShell>

      <FieldShell
        label="Senha (mínimo 8 caracteres)"
        htmlFor="password"
        error={showErrors ? accountErrors.password : undefined}
      >
        <input
          id="password"
          type="password"
          autoComplete="new-password"
          className={bigFieldClass}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </FieldShell>

      {state?.error && (
        <div
          role="alert"
          className="flex flex-col gap-2 border-2 border-risk-critical bg-risk-critical-bg px-3 py-3 font-mono text-sm text-risk-critical"
        >
          <p>{state.error}</p>
          {/* a taken e-mail already says it in the paragraph above; don't repeat it as a field message */}
          {!state.emailTaken && fieldMessages.length > 0 && (
            <ul className="list-disc pl-5 text-xs">
              {fieldMessages.map(([field, message]) => (
                <li key={field}>{message}</li>
              ))}
            </ul>
          )}
          {state.emailTaken && (
            <Link href="/entrar" className="font-semibold underline">
              Entrar com este e-mail
            </Link>
          )}
          {firstBadStep && (
            <button type="button" onClick={() => onFixStep(firstBadStep)} className="w-fit font-semibold underline">
              Corrigir respostas
            </button>
          )}
        </div>
      )}

      <button type="submit" disabled={pending} className={primaryButtonClass}>
        {pending ? "Criando sua conta..." : "Criar minha conta"}
      </button>
    </form>
  );
}
