import type { Metadata } from "next";

import { SshConfigurationClient } from "@/components/ssh-configuration-client";

export const metadata: Metadata = {
  title: "SSH Configuration Guide - Hasir",
  description:
    "Learn how to configure SSH keys for secure access to Hasir repositories.",
};

export default function SshConfigurationPage() {
  return <SshConfigurationClient />;
}
