export function FieldShell({
  label,
  htmlFor,
  error,
  children,
}: {
  label: string;
  htmlFor: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label
        htmlFor={htmlFor}
        className="font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft"
      >
        {label}
      </label>
      {children}
      {error && <p className="font-mono text-xs text-risk-critical">{error}</p>}
    </div>
  );
}

export const fieldClass =
  "border-2 border-ink bg-paper px-3 py-2 font-mono text-sm text-ink outline-none transition focus:bg-card focus:shadow-[var(--shadow-hard)]";
