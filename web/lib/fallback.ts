import { unstable_rethrow } from "next/navigation";

/**
 * Resolves to `value` when an optional load fails, but lets Next's control-flow
 * errors (the redirect("/sair") from lib/api on a 401, notFound, …) through.
 */
export async function orFallback<T>(promise: Promise<T>, value: T): Promise<T> {
  try {
    return await promise;
  } catch (err) {
    unstable_rethrow(err);
    return value;
  }
}
