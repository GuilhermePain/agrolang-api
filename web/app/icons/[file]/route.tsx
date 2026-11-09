import { ImageResponse } from "next/og";

// PNG icons for the web app manifest, drawn from the same sprout mark used in
// the nav (components/Icons.tsx). "maskable" keeps the mark inside the safe zone.
const SPECS: Record<string, { size: number; mark: number }> = {
  "192": { size: 192, mark: 0.62 },
  "512": { size: 512, mark: 0.62 },
  maskable: { size: 512, mark: 0.46 },
  apple: { size: 180, mark: 0.6 },
};

export function generateStaticParams() {
  return Object.keys(SPECS).map((file) => ({ file }));
}

export async function GET(_req: Request, { params }: { params: Promise<{ file: string }> }) {
  const { file } = await params;
  const spec = SPECS[file];
  if (!spec) return new Response("Not found", { status: 404 });

  const mark = Math.round(spec.size * spec.mark);
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "#8a4a2f",
        }}
      >
        <svg
          width={mark}
          height={mark}
          viewBox="0 0 24 24"
          fill="none"
          stroke="#ece3d3"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M7 20h10" />
          <path d="M10 20c5.5-2.5.8-6.4 3-10" />
          <path d="M9.5 9.4c1.1.8 1.8 2.2 2.3 3.7-2 .4-3.5.4-4.8-.3-1.2-.6-2.3-1.9-3-4.2 2.8-.5 4.4.1 5.5.8z" />
          <path d="M14.1 6a7 7 0 0 0-1.1 4c1.9-.1 3.3-.6 4.3-1.4 1-1 1.6-2.3 1.7-4.6-2.7.1-4.3 1-4.9 2z" />
        </svg>
      </div>
    ),
    { width: spec.size, height: spec.size },
  );
}
