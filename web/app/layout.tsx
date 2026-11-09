import type { Metadata, Viewport } from "next";
import { Fraunces, IBM_Plex_Mono } from "next/font/google";
import AppNav from "@/components/AppNav";
import AppShell from "@/components/AppShell";
import PwaClient from "@/components/PwaClient";
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
  applicationName: "AgroResiliente",
  icons: { apple: "/icons/apple" },
  appleWebApp: { capable: true, title: "AgroResiliente", statusBarStyle: "default" },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: "#ddd0b8",
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
      <body className="h-full min-h-dvh lg:h-dvh lg:overflow-hidden flex flex-col bg-paper text-ink">
        <AppShell nav={<AppNav />}>{children}</AppShell>
        <PwaClient />
      </body>
    </html>
  );
}
