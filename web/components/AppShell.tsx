"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { isFullscreenPath } from "@/lib/chrome";

// Guests can only reach the fullscreen routes (proxy.ts sends them to /entrar),
// so hiding the nav there also hides it from every guest.
export default function AppShell({ nav, children }: { nav: ReactNode; children: ReactNode }) {
  const fullscreen = isFullscreenPath(usePathname());

  if (fullscreen) {
    return <main className="flex min-h-0 flex-1 flex-col overflow-y-auto">{children}</main>;
  }

  return (
    <>
      {nav}
      {/* ESTÁTICA EM TELA CHEIA NO DESKTOP, COM ROLAGEM FLUIDA NO MOBILE */}
      <main className="flex min-h-0 flex-1 flex-col overflow-y-auto lg:overflow-hidden">{children}</main>
    </>
  );
}
