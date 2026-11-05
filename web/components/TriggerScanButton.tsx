"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { triggerScan } from "@/lib/api";
import { BoltIcon } from "./Icons";

export default function TriggerScanButton() {
  const [running, setRunning] = useState(false);
  const router = useRouter();

  async function handleTrigger() {
    setRunning(true);
    try {
      await triggerScan();
      alert("Varredura climática concluída com sucesso! Os talhões foram reavaliados.");
      router.refresh();
    } catch {
      alert("Não foi possível executar a varredura no momento. Verifique a conexão com o backend.");
    } finally {
      setRunning(false);
    }
  }

  return (
    <button
      type="button"
      onClick={handleTrigger}
      disabled={running}
      className="flex items-center gap-2 border-2 border-ink bg-paper px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition hover:bg-ink hover:text-paper active:translate-x-0.5 active:translate-y-0.5 disabled:opacity-50"
    >
      <BoltIcon className="h-4 w-4 text-accent" />
      <span>{running ? "Processando varredura..." : "Forçar Varredura Agora"}</span>
    </button>
  );
}
