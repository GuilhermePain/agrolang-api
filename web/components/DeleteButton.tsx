"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

interface DeleteButtonProps {
  onDelete: () => Promise<unknown>;
  redirectTo: string;
  confirmMessage?: string;
  label?: string;
}

export default function DeleteButton({
  onDelete,
  redirectTo,
  confirmMessage = "Tem certeza que deseja excluir? Esta ação não pode ser desfeita.",
  label = "Excluir",
}: DeleteButtonProps) {
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  async function handleClick() {
    if (!window.confirm(confirmMessage)) return;

    setLoading(true);
    try {
      await onDelete();
      router.push(redirectTo);
      router.refresh();
    } catch {
      alert("Erro ao excluir. Tente novamente.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={loading}
      className="border border-risk-critical bg-paper px-3 py-1 font-mono text-xs font-semibold uppercase text-risk-critical transition hover:bg-risk-critical hover:text-paper disabled:opacity-50"
    >
      {loading ? "Excluindo..." : label}
    </button>
  );
}

