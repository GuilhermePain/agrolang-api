import { NextResponse } from "next/server";
import { clearSessionCookie } from "@/lib/session";

// Landing point for stale sessions: drops the cookie so /entrar doesn't bounce back.
export async function GET(request: Request) {
  await clearSessionCookie();
  return NextResponse.redirect(new URL("/entrar", request.url));
}
