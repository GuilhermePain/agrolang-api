"use client";

import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { bigFieldClass } from "./styles";

export default function IdentityStep({
  answers,
  technician,
  onChange,
}: {
  answers: QuizAnswers;
  technician: boolean;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  return (
    <>
      <FieldShell label="Seu nome" htmlFor="name">
        <input
          id="name"
          className={bigFieldClass}
          value={answers.name}
          onChange={(e) => onChange({ name: e.target.value })}
          autoComplete="name"
          placeholder="Ex: José da Silva"
        />
      </FieldShell>

      <FieldShell label="WhatsApp (com DDD)" htmlFor="phone">
        <input
          id="phone"
          className={bigFieldClass}
          value={answers.phone}
          onChange={(e) => onChange({ phone: e.target.value })}
          inputMode="numeric"
          autoComplete="tel-national"
          placeholder="19 99999-9999"
        />
      </FieldShell>

      {technician && (
        <FieldShell label="Organização" htmlFor="organization">
          <input
            id="organization"
            className={bigFieldClass}
            value={answers.organization}
            onChange={(e) => onChange({ organization: e.target.value })}
            autoComplete="organization"
            placeholder="Ex: Cooperativa Campo Verde"
          />
        </FieldShell>
      )}
    </>
  );
}
