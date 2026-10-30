import type { Metadata } from "next";
import LoginForm from "@/components/LoginForm";

export const metadata: Metadata = { title: "Entrar · AgroResiliente" };

export default function LoginPage() {
  return (
    <div className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Estação de campo
        </p>
        <h1 className="font-display text-3xl italic text-ink">Entrar</h1>
      </div>

      <LoginForm />
    </div>
  );
}
