"use client";

import { useActionState } from "react";
import Link from "next/link";
import { loginAction } from "@/app/actions/auth";
import { FieldShell, fieldClass } from "./FieldShell";

const bigField = `${fieldClass} min-h-12 text-base`;

export default function LoginForm() {
  const [state, formAction, pending] = useActionState(loginAction, undefined);

  return (
    <form action={formAction} className="flex flex-col gap-5 rise-in">
      <FieldShell label="E-mail" htmlFor="email">
        <input
          id="email"
          name="email"
          type="email"
          inputMode="email"
          autoComplete="email"
          className={bigField}
          placeholder="voce@exemplo.com"
          required
        />
      </FieldShell>

      <FieldShell label="Senha" htmlFor="password">
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          className={bigField}
          required
        />
      </FieldShell>

      {state?.error && (
        <p
          role="alert"
          className="border-2 border-risk-critical bg-risk-critical-bg px-3 py-2 font-mono text-sm text-risk-critical"
        >
          {state.error}
        </p>
      )}

      <button
        type="submit"
        disabled={pending}
        className="min-h-12 border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
      >
        {pending ? "Entrando..." : "Entrar"}
      </button>

      <p className="font-mono text-xs text-ink-soft">
        Ainda não tem conta?{" "}
        <Link href="/cadastro" className="text-accent underline">
          Fazer o cadastro
        </Link>
      </p>
    </form>
  );
}
