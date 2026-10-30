"use server";

import { redirect } from "next/navigation";
import { ApiError, login, logout, signup } from "@/lib/api";
import { clearSessionCookie, setSessionCookie } from "@/lib/session";

export type AuthFormState =
  | { error?: string; fields?: Record<string, string>; emailTaken?: boolean }
  | undefined;

const NETWORK_ERROR = "Sem conexão com o servidor. Tente de novo.";

export async function signupAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const raw = formData.get("payload");
  if (typeof raw !== "string" || raw === "") {
    return { error: "Não foi possível enviar o cadastro. Tente de novo." };
  }

  let payload: unknown;
  try {
    payload = JSON.parse(raw);
  } catch {
    return { error: "Não foi possível enviar o cadastro. Tente de novo." };
  }

  let session;
  try {
    session = await signup(payload);
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 409) {
        return { emailTaken: true, error: "Este e-mail já tem cadastro.", fields: err.fields };
      }
      if (err.status === 400) {
        return { error: "Revise as respostas abaixo.", fields: err.fields };
      }
      return { error: "Não foi possível criar a conta agora. Tente de novo." };
    }
    return { error: NETWORK_ERROR };
  }

  await setSessionCookie(session.token, new Date(session.expires_at));
  redirect("/bem-vindo");
}

export async function loginAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const email = String(formData.get("email") ?? "");
  const password = String(formData.get("password") ?? "");
  if (email.trim() === "" || password === "") {
    return { error: "Digite o e-mail e a senha." };
  }

  let session;
  try {
    session = await login(email, password);
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) return { error: "E-mail ou senha incorretos." };
      return { error: "Não foi possível entrar agora. Tente de novo." };
    }
    return { error: NETWORK_ERROR };
  }

  await setSessionCookie(session.token, new Date(session.expires_at));
  redirect("/");
}

export async function logoutAction(): Promise<void> {
  try {
    await logout();
  } catch (err) {
    // an already-expired session is fine; anything else is not worth blocking the exit
    if (!(err instanceof ApiError)) throw err;
  }
  await clearSessionCookie();
  redirect("/entrar");
}
