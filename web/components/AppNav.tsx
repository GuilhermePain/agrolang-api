import Link from "next/link";
import { logoutAction } from "@/app/actions/auth";
import { SproutIcon } from "./Icons";

const linkClass =
  "flex min-h-11 items-center justify-center rounded px-2 text-ink transition hover:bg-ink hover:text-paper sm:px-2.5";

// BARRA SUPERIOR DE NAVEGAÇÃO COMPACTA (FIXA NO DESKTOP). Only rendered for
// logged-in routes; see AppShell.
export default function AppNav() {
  return (
    <nav className="shrink-0 border-b-2 border-ink bg-paper-deep px-4 py-1 sm:px-6">
      <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-x-4">
        <Link href="/" className="group flex min-h-11 items-center gap-2">
          <SproutIcon className="h-5 w-5 text-accent transition group-hover:scale-110" />
          <span className="font-display text-lg font-bold italic tracking-tight text-ink sm:text-xl">
            AgroResiliente
          </span>
        </Link>

        <form action={logoutAction} className="sm:order-last">
          <button
            type="submit"
            className="min-h-11 border-2 border-ink px-3 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink transition hover:bg-ink hover:text-paper"
          >
            Sair
          </button>
        </form>

        <div className="grid w-full grid-cols-4 gap-1 pb-1 font-mono text-[11px] font-semibold uppercase tracking-wide sm:flex sm:w-auto sm:flex-1 sm:justify-end sm:gap-3 sm:pb-0 sm:text-xs">
          <Link href="/" className={linkClass}>
            <span className="sm:hidden">Mapa</span>
            <span className="hidden sm:inline">Mapa de Risco</span>
          </Link>
          <Link href="/producers" className={linkClass}>
            Produtores
          </Link>
          <Link href="/notifications" className={linkClass}>
            Disparos
          </Link>
          <Link
            href="/properties/new"
            className="flex min-h-11 items-center justify-center rounded border border-ink bg-accent px-2 text-paper transition hover:bg-ink sm:px-3"
          >
            + Talhão
          </Link>
        </div>
      </div>
    </nav>
  );
}
