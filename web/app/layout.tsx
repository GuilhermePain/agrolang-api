import type { Metadata } from "next";
import Link from "next/link";
import { Fraunces, IBM_Plex_Mono } from "next/font/google";
import { SproutIcon } from "@/components/Icons";
import "./globals.css";

const displayFont = Fraunces({
  variable: "--font-display",
  subsets: ["latin"],
  axes: ["opsz", "SOFT", "WONK"],
  style: ["normal", "italic"],
});

const monoFont = IBM_Plex_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "AgroResiliente",
  description: "Estação de inteligência e adaptação climática para propriedades rurais",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="pt-BR"
      className={`${displayFont.variable} ${monoFont.variable} h-full antialiased`}
    >
      <body className="h-full min-h-screen lg:h-screen lg:overflow-hidden flex flex-col bg-paper text-ink">
        {/* BARRA SUPERIOR DE NAVEGAÇÃO COMPACTA (FIXA NO DESKTOP) */}
        <nav className="shrink-0 border-b-2 border-ink bg-paper-deep px-4 py-2 sm:px-6">
          <div className="mx-auto flex max-w-7xl items-center justify-between gap-4">
            <Link href="/" className="flex items-center gap-2 group">
              <SproutIcon className="h-5 w-5 text-accent transition group-hover:scale-110" />
              <span className="font-display text-lg font-bold italic tracking-tight text-ink sm:text-xl">
                AgroResiliente
              </span>
            </Link>

            <div className="flex items-center gap-2 sm:gap-3 font-mono text-xs font-semibold uppercase tracking-wide">
              <Link
                href="/"
                className="rounded px-2.5 py-1 text-ink transition hover:bg-ink hover:text-paper"
              >
                Mapa de Risco
              </Link>
              <Link
                href="/producers"
                className="rounded px-2.5 py-1 text-ink transition hover:bg-ink hover:text-paper"
              >
                Produtores
              </Link>
              <Link
                href="/notifications"
                className="rounded px-2.5 py-1 text-ink transition hover:bg-ink hover:text-paper"
              >
                Disparos
              </Link>
              <Link
                href="/properties/new"
                className="rounded border border-ink bg-accent px-3 py-1 text-paper transition hover:bg-ink"
              >
                + Talhão
              </Link>
            </div>
          </div>
        </nav>

        {/* ÁREA PRINCIPAL: ESTÁTICA EM TELA CHEIA NO DESKTOP, COM ROLAGEM FLUIDA NO MOBILE */}
        <main className="flex flex-1 flex-col overflow-y-auto lg:overflow-hidden min-h-0">
          {children}
        </main>
      </body>
    </html>
  );
}
