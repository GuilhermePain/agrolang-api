"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { createProducer, ApiError } from "@/lib/api";
import { validateProducer, type ProducerFormValues, type ProducerFormErrors } from "@/lib/validation";
import { FieldShell, fieldClass } from "./FieldShell";

const EMPTY: ProducerFormValues = { name: "", whatsapp_phone: "", city: "", state: "" };

export default function ProducerForm() {
  const [values, setValues] = useState<ProducerFormValues>(EMPTY);
  const [errors, setErrors] = useState<ProducerFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [createdId, setCreatedId] = useState<string | null>(null);

  function update<K extends keyof ProducerFormValues>(key: K, value: ProducerFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validation = validateProducer(values);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const created = await createProducer(values);
      setCreatedId(created.id);
    } catch (err) {
      setSubmitError(
        err instanceof ApiError ? err.message : "Não foi possível cadastrar o produtor."
      );
    } finally {
      setSubmitting(false);
    }
  }

  if (createdId) {
    return (
      <div className="rise-in border-2 border-risk-low bg-risk-low-bg px-5 py-6 font-mono text-sm text-risk-low">
        <p className="font-display text-lg italic text-ink">Produtor cadastrado.</p>
        <p className="mt-2 text-ink-soft">
          Agora cadastre uma propriedade para este produtor para começar o monitoramento.
        </p>
        <div className="mt-4 flex flex-wrap gap-3">
          <Link
            href={`/properties/new?producer_id=${createdId}`}
            className="border-2 border-ink bg-accent px-4 py-2 font-semibold uppercase tracking-wide text-paper hover:bg-ink"
          >
            Cadastrar propriedade
          </Link>
          <Link href="/" className="border-2 border-ink px-4 py-2 font-semibold uppercase tracking-wide text-ink">
            Voltar ao mapa
          </Link>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 rise-in">
      <FieldShell label="Nome" htmlFor="name" error={errors.name}>
        <input
          id="name"
          className={fieldClass}
          value={values.name}
          onChange={(e) => update("name", e.target.value)}
          placeholder="Ex: José da Silva"
        />
      </FieldShell>

      <FieldShell label="WhatsApp" htmlFor="whatsapp_phone" error={errors.whatsapp_phone}>
        <input
          id="whatsapp_phone"
          className={fieldClass}
          value={values.whatsapp_phone}
          onChange={(e) => update("whatsapp_phone", e.target.value)}
          placeholder="Ex: 5511999999999"
        />
      </FieldShell>

      <div className="grid grid-cols-2 gap-4">
        <FieldShell label="Cidade" htmlFor="city" error={errors.city}>
          <input
            id="city"
            className={fieldClass}
            value={values.city}
            onChange={(e) => update("city", e.target.value)}
            placeholder="Ex: Campinas"
          />
        </FieldShell>

        <FieldShell label="Estado" htmlFor="state" error={errors.state}>
          <input
            id="state"
            className={fieldClass}
            value={values.state}
            onChange={(e) => update("state", e.target.value.toUpperCase())}
            placeholder="Ex: SP"
            maxLength={2}
          />
        </FieldShell>
      </div>

      {submitError && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-3 py-2 font-mono text-sm text-risk-critical">
          {submitError}
        </p>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="w-fit border-2 border-ink bg-accent px-5 py-2.5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
      >
        {submitting ? "Cadastrando..." : "Cadastrar produtor"}
      </button>
    </form>
  );
}
