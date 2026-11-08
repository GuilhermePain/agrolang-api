"use server";

import { revalidatePath } from "next/cache";
import {
  ApiError,
  createProducer,
  createProperty,
  deleteProducer,
  deleteProperty,
  resolveAlert,
  triggerScan,
  updateProducer,
  updateProperty,
} from "@/lib/api";
import type { DeleteResult, NewProducer, NewProperty, Producer, Property } from "@/lib/types";

export type ActionResult<T> = { ok: true; data: T } | { ok: false; error: string };

// ApiErrors become a message for the form; anything else (session redirects
// from lib/api, network failures) is rethrown so Next can handle it.
async function run<T>(call: () => Promise<T>, revalidate: boolean): Promise<ActionResult<T>> {
  try {
    const data = await call();
    // Every page reads per-user data with no-store, so invalidating the whole
    // tree is cheap and keeps already-visited pages (dashboard, lists) fresh.
    if (revalidate) revalidatePath("/", "layout");
    return { ok: true, data };
  } catch (err) {
    if (err instanceof ApiError) return { ok: false, error: err.message };
    throw err;
  }
}

export async function createProducerAction(values: NewProducer): Promise<ActionResult<Producer>> {
  return run(() => createProducer(values), true);
}

export async function updateProducerAction(
  id: string,
  values: NewProducer
): Promise<ActionResult<Producer>> {
  return run(() => updateProducer(id, values), true);
}

export async function deleteProducerAction(id: string): Promise<ActionResult<DeleteResult>> {
  return run(() => deleteProducer(id), true);
}

export async function createPropertyAction(values: NewProperty): Promise<ActionResult<Property>> {
  return run(() => createProperty(values), true);
}

export async function updatePropertyAction(
  id: string,
  values: NewProperty
): Promise<ActionResult<Property>> {
  return run(() => updateProperty(id, values), true);
}

export async function deletePropertyAction(id: string): Promise<ActionResult<DeleteResult>> {
  return run(() => deleteProperty(id), true);
}

export async function resolveAlertAction(id: string): Promise<ActionResult<DeleteResult>> {
  return run(() => resolveAlert(id), true);
}

export async function triggerScanAction(): Promise<ActionResult<{ status: string }>> {
  return run(() => triggerScan(), true);
}
