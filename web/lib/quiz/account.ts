export type AccountErrors = { email?: string; password?: string };

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MIN_PASSWORD = 8;
const MAX_PASSWORD_BYTES = 72;

export function validateAccount(email: string, password: string): AccountErrors {
  const errors: AccountErrors = {};
  if (!EMAIL_PATTERN.test(email.trim())) errors.email = "Digite um e-mail válido.";
  if (password.length < MIN_PASSWORD) {
    errors.password = "A senha precisa ter pelo menos 8 caracteres.";
  } else if (new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES) {
    errors.password = "A senha pode ter no máximo 72 caracteres.";
  }
  return errors;
}
