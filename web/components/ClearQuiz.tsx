"use client";

import { useEffect } from "react";
import { clearAnswers } from "@/lib/quiz/state";

/** The quiz answers are only needed until the account exists; drop them once the welcome page loads. */
export default function ClearQuiz() {
  useEffect(() => {
    clearAnswers();
  }, []);
  return null;
}
