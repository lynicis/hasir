"use client";

import dynamic from "next/dynamic";

const HeaderInner = dynamic(
  () => import("@/components/header").then((mod) => mod.Header),
  { ssr: false }
);

export function HeaderClient() {
  return <HeaderInner />;
}
