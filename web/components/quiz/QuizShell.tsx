import type { ReactNode } from "react";

export default function QuizShell({
  index,
  total,
  title,
  hint,
  onBack,
  footer,
  children,
}: {
  index: number;
  /** null while the path (and so the step count) isn't known yet. */
  total: number | null;
  title: string;
  hint?: string;
  onBack?: () => void;
  footer?: ReactNode;
  children: ReactNode;
}) {
  const progress = total ? ((index + 1) / total) * 100 : 8;

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-1 flex-col">
      <header className="flex flex-col gap-3 px-5 pb-2 pt-5">
        <div className="flex items-center justify-between">
          {onBack ? (
            <button
              type="button"
              onClick={onBack}
              className="-ml-2 min-h-11 px-2 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
            >
              &larr; Voltar
            </button>
          ) : (
            <span />
          )}
          <span className="font-mono text-[11px] uppercase tracking-[0.2em] text-ink-soft">
            {total ? `Passo ${index + 1} de ${total}` : "Passo 1"}
          </span>
        </div>
        <div
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(progress)}
          className="h-2 border-2 border-ink bg-card"
        >
          <div className="h-full bg-accent transition-all" style={{ width: `${progress}%` }} />
        </div>
      </header>

      <main className="flex flex-1 flex-col gap-6 px-5 py-4">
        <div className="rise-in">
          <h1 className="font-display text-2xl italic leading-tight text-ink">{title}</h1>
          {hint && <p className="mt-1 text-sm text-ink-soft">{hint}</p>}
        </div>
        <div className="flex flex-col gap-4">{children}</div>
      </main>

      {footer && (
        <footer className="sticky bottom-0 border-t-2 border-ink bg-paper px-5 py-4">{footer}</footer>
      )}
    </div>
  );
}
