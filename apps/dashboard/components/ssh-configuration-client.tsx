"use client";

import dynamic from "next/dynamic";

const SshConfigurationContentInner = dynamic(
  () =>
    import("@/components/ssh-configuration-content").then(
      (mod) => mod.SshConfigurationContent,
    ),
  { ssr: false },
);

export function SshConfigurationClient() {
  return <SshConfigurationContentInner />;
}
