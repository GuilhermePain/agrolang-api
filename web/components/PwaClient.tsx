"use client";

import { useEffect, useState, useSyncExternalStore } from "react";

type InstallEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

const DISMISS_KEY = "pwa-install-dismissed";

function isStandalone(): boolean {
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    (navigator as Navigator & { standalone?: boolean }).standalone === true
  );
}

function wasDismissed(): boolean {
  try {
    return localStorage.getItem(DISMISS_KEY) === "1";
  } catch {
    return false;
  }
}

const noopSubscribe = () => () => {};

// iOS Safari has no beforeinstallprompt; only there we show manual instructions.
function iosHintVisible(): boolean {
  const ua = navigator.userAgent;
  const ios = /iPad|iPhone|iPod/.test(ua) && /Safari/.test(ua) && !/CriOS|FxiOS|EdgiOS/.test(ua);
  return ios && !isStandalone() && !wasDismissed();
}

// Registers the service worker and shows a dismissible install hint:
// native prompt on Chromium/Android, manual instructions on iOS Safari.
export default function PwaClient() {
  const [installEvent, setInstallEvent] = useState<InstallEvent | null>(null);
  const [dismissed, setDismissed] = useState(false);
  const iosHint = useSyncExternalStore(noopSubscribe, iosHintVisible, () => false) && !dismissed;

  useEffect(() => {
    if ("serviceWorker" in navigator && process.env.NODE_ENV === "production") {
      navigator.serviceWorker.register("/sw.js", { scope: "/", updateViaCache: "none" }).catch(() => {});
    }

    if (isStandalone() || wasDismissed()) return;

    const onPrompt = (e: Event) => {
      e.preventDefault();
      setInstallEvent(e as InstallEvent);
    };
    const onInstalled = () => setInstallEvent(null);
    window.addEventListener("beforeinstallprompt", onPrompt);
    window.addEventListener("appinstalled", onInstalled);

    return () => {
      window.removeEventListener("beforeinstallprompt", onPrompt);
      window.removeEventListener("appinstalled", onInstalled);
    };
  }, []);

  function dismiss() {
    try {
      localStorage.setItem(DISMISS_KEY, "1");
    } catch {
      // private mode: the hint just comes back next visit
    }
    setInstallEvent(null);
    setDismissed(true);
  }

  async function install() {
    if (!installEvent) return;
    await installEvent.prompt();
    await installEvent.userChoice;
    setInstallEvent(null);
  }

  if (!installEvent && !iosHint) return null;

  return (
    <div
      role="dialog"
      aria-label="Instalar aplicativo"
      className="fixed inset-x-3 z-[1000] mx-auto flex max-w-md items-center gap-3 border-2 border-ink bg-card p-3 font-mono text-xs shadow-[var(--shadow-hard)]"
      style={{ bottom: "max(0.75rem, env(safe-area-inset-bottom))" }}
    >
      <p className="flex-1 leading-snug text-ink">
        {installEvent
          ? "Instale o AgroResiliente na tela inicial."
          : "Para instalar: toque em Compartilhar e depois em “Adicionar à Tela de Início”."}
      </p>
      {installEvent && (
        <button
          type="button"
          onClick={install}
          className="min-h-11 border-2 border-ink bg-accent px-3 font-semibold uppercase tracking-wide text-paper"
        >
          Instalar
        </button>
      )}
      <button
        type="button"
        onClick={dismiss}
        aria-label="Dispensar"
        className="min-h-11 min-w-11 border-2 border-ink px-2 font-semibold text-ink"
      >
        ✕
      </button>
    </div>
  );
}
