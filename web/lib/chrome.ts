// Routes that own the whole screen (login, quiz, welcome): no app nav, and the
// page scrolls normally instead of the desktop "fixed dashboard" layout.
export const FULLSCREEN_PATHS = ["/entrar", "/cadastro", "/sair", "/bem-vindo"];

export function isFullscreenPath(pathname: string | null): boolean {
  if (!pathname) return false;
  return FULLSCREEN_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`));
}
