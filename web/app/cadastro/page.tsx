import type { Metadata } from "next";
import { Suspense } from "react";
import QuizFlow from "@/components/quiz/QuizFlow";

export const metadata: Metadata = { title: "Cadastro · AgroResiliente" };

export default function SignupPage() {
  return (
    <Suspense fallback={null}>
      <QuizFlow />
    </Suspense>
  );
}
