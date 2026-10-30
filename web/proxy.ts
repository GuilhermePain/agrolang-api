import { NextResponse, type NextRequest } from "next/server";

const SESSION_COOKIE = "session";
const PUBLIC_PATHS = ["/entrar", "/cadastro", "/sair"];
const GUEST_ONLY_PATHS = ["/entrar", "/cadastro"];

const matches = (pathname: string, paths: string[]) =>
  paths.some((p) => pathname === p || pathname.startsWith(`${p}/`));

// Optimistic check only: the real validation is requireUser() on the pages.
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasSession = request.cookies.has(SESSION_COOKIE);

  if (!hasSession && !matches(pathname, PUBLIC_PATHS)) {
    return NextResponse.redirect(new URL("/entrar", request.url));
  }
  if (hasSession && matches(pathname, GUEST_ONLY_PATHS)) {
    return NextResponse.redirect(new URL("/", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.(?:png|jpg|jpeg|svg|ico|webp)$).*)"],
};
