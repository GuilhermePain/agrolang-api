"use client";

import dynamic from "next/dynamic";
import type { Property, Alert } from "@/lib/api";

const PropertyMap = dynamic(() => import("./PropertyMap"), { ssr: false });

export default function PropertyMapClient({
  items,
}: {
  items: { property: Property; alerts: Alert[] }[];
}) {
  return <PropertyMap items={items} />;
}
