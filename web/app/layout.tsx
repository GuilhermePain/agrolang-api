import type { Metadata } from "next";
import { Fraunces, IBM_Plex_Mono } from "next/font/google";
import AppNav from "@/components/AppNav";
import AppShell from "@/components/AppShell";
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
        <AppShell nav={<AppNav />}>{children}</AppShell>
      </body>
    </html>
  );
}
