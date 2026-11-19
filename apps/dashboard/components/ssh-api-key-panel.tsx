"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hasir/ui/components/card";
import { Separator } from "@hasir/ui/components/separator";
import { Key } from "lucide-react";

import { SshKeysSection } from "@/components/ssh-keys-section";
import { ApiKeysSection } from "@/components/api-keys-section";

export function SshApiKeyPanel() {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Key className="size-5 text-muted-foreground" />
          <CardTitle className="text-2xl">SSH &amp; API Access</CardTitle>
        </div>
        <CardDescription>
          Manage your SSH public key and generate personal API keys for secure
          HTTP access.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-8">
        <SshKeysSection />
        <Separator />
        <ApiKeysSection />
      </CardContent>
    </Card>
  );
}
