import { cache } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { API_URL } from "./config";
import type { AuthUser } from "./types";

export const SESSION_COOKIE = "session";

export async function getSessionToken(): Promise<string | null> {
  return (await cookies()).get(SESSION_COOKIE)?.value ?? null;
}

export async function setSessionCookie(token: string, expiresAt: Date): Promise<void> {
  (await cookies()).set(SESSION_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: expiresAt,
  });
}

export async function clearSessionCookie(): Promise<void> {
  (await cookies()).delete(SESSION_COOKIE);
}

/** The logged-in account, or null when there's no cookie or the API says the session is gone. */
export const getCurrentUser = cache(async (): Promise<AuthUser | null> => {
  const token = await getSessionToken();
  if (!token) return null;

  const res = await fetch(`${API_URL}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (res.status === 401) return null;
  if (!res.ok) throw new Error(`API /auth/me failed: ${res.status}`);
  return (await res.json()) as AuthUser;
});

/** For protected pages: sends stale or missing sessions through /sair, which clears the cookie. */
export async function requireUser(): Promise<AuthUser> {
  const user = await getCurrentUser();
  if (!user) redirect("/sair");
  return user;
}
