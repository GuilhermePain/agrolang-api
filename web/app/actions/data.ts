"use server";

import { ApiError, createProducer, createProperty } from "@/lib/api";
import type { NewProducer, NewProperty, Producer, Property } from "@/lib/types";

export type ActionResult<T> = { ok: true; data: T } | { ok: false; error: string };

export async function createProducerAction(values: NewProducer): Promise<ActionResult<Producer>> {
  try {
    return { ok: true, data: await createProducer(values) };
  } catch (err) {
    if (err instanceof ApiError) return { ok: false, error: err.message };
    throw err;
  }
}

export async function createPropertyAction(values: NewProperty): Promise<ActionResult<Property>> {
  try {
    return { ok: true, data: await createProperty(values) };
  } catch (err) {
    if (err instanceof ApiError) return { ok: false, error: err.message };
    throw err;
  }
}
