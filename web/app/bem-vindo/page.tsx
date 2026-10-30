import type { Metadata } from "next";
import Link from "next/link";
import ClearQuiz from "@/components/ClearQuiz";
import { listProducers } from "@/lib/api";
import { CONCERN_LABELS, welcomeMessage } from "@/lib/personalize";
import { requireUser } from "@/lib/session";

export const metadata: Metadata = { title: "Tudo pronto · AgroResiliente" };

export default async function WelcomePage() {
  const user = await requireUser();
  const producers = user.profile === "producer" ? await listProducers() : [];
  const city = producers[0]?.city;
  const firstName = user.name.trim().split(/\s+/)[0];
  const concerns = user.concerns.map((c) => CONCERN_LABELS[c]).filter(Boolean);

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-1 flex-col justify-center gap-6 px-5 py-8">
      <ClearQuiz />

      <div className="rise-in">
        <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">Cadastro concluído</p>
        <h1 className="font-display text-3xl italic leading-tight text-ink">Tudo pronto, {firstName}.</h1>
        <p className="mt-3 text-base text-ink">{welcomeMessage(user, city)}</p>
      </div>

      {concerns.length > 0 && (
        <ul className="flex flex-wrap gap-2 rise-in" style={{ animationDelay: "80ms" }}>
          {concerns.map((label) => (
            <li
              key={label}
              className="border-2 border-ink bg-card px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink"
            >
              {label}
            </li>
          ))}
        </ul>
      )}

      <div className="flex flex-col gap-3 rise-in" style={{ animationDelay: "140ms" }}>
        {user.profile === "technician" && (
          <Link
            href="/producers/new"
            className="flex min-h-12 items-center justify-center border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink"
          >
            Cadastrar o primeiro produtor
          </Link>
        )}
        <Link
          href="/"
          className={`flex min-h-12 items-center justify-center border-2 border-ink px-5 font-mono text-sm font-semibold uppercase tracking-wide transition ${
            user.profile === "technician" ? "text-ink hover:bg-ink hover:text-paper" : "bg-accent text-paper hover:bg-ink"
          }`}
        >
          Ir para o painel
        </Link>
      </div>
    </div>
  );
}
