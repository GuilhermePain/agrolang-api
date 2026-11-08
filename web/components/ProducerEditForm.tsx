"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { updateProducerAction } from "@/app/actions/data";
import type { Producer } from "@/lib/types";
import { validateProducer, type ProducerFormValues, type ProducerFormErrors } from "@/lib/validation";
import { FieldShell, fieldClass } from "./FieldShell";

export default function ProducerEditForm({ producer }: { producer: Producer }) {
  const router = useRouter();
  const [values, setValues] = useState<ProducerFormValues>({
    name: producer.name,
    whatsapp_phone: producer.whatsapp_phone,
    city: producer.city,
    state: producer.state,
  });
  const [errors, setErrors] = useState<ProducerFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

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
      const result = await updateProducerAction(producer.id, values);
      if (result.ok) {
        router.push(`/producers/${producer.id}`);
      } else {
        setSubmitError(result.error);
      }
    } catch {
      setSubmitError("Não foi possível atualizar os dados do produtor.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 rise-in">
      <FieldShell label="Nome Completo" htmlFor="name" error={errors.name}>
        <input
          id="name"
          className={fieldClass}
          value={values.name}
          onChange={(e) => update("name", e.target.value)}
          placeholder="Ex: José da Silva"
        />
      </FieldShell>

      <FieldShell label="WhatsApp (com DDD)" htmlFor="whatsapp_phone" error={errors.whatsapp_phone}>
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

        <FieldShell label="Estado (UF)" htmlFor="state" error={errors.state}>
          <input
            id="state"
            className={fieldClass}
            value={values.state}
            onChange={(e) => update("state", e.target.value)}
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

      <div className="flex items-center gap-3 pt-2">
        <button
          type="submit"
          disabled={submitting}
          className="border-2 border-ink bg-accent px-5 py-2.5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
        >
          {submitting ? "Salvando..." : "Salvar Alterações"}
        </button>

        <Link
          href={`/producers/${producer.id}`}
          className="border border-line px-4 py-2 font-mono text-xs font-semibold uppercase text-ink hover:border-ink"
        >
          Cancelar
        </Link>
      </div>
    </form>
  );
}

