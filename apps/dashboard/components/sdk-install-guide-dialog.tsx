"use client";

import { BookOpen, Copy, Check } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

interface SdkInstallGuideDialogProps {
  organizationId: string;
  repositoryId: string;
  commitHash: string;
}

export function SdkInstallGuideDialog({ organizationId, repositoryId, commitHash }: SdkInstallGuideDialogProps) {
  const [copiedText, setCopiedText] = useState<string | null>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  let host = "localhost";
  try {
    const urlObj = new URL(apiUrl);
    host = urlObj.hostname;
  } catch { }

  const goImportPath = `${host}/sdk/${organizationId}/${repositoryId}/${commitHash}/go-connectrpc`;
  const httpJsUrl = `git+${apiUrl}/sdk/${organizationId}/${repositoryId}/${commitHash}/js-connectrpc/`;
  const sshJsUrl = `git+ssh://git@${host}:2222/sdk/${organizationId}/${repositoryId}/${commitHash}/js-connectrpc/`;

  const jsHttpCommand = `npm install ${httpJsUrl}`;
  const jsSshCommand = `npm install ${sshJsUrl}`;

  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedText(id);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopiedText(null), 2000);
    } catch {
      toast.error("Failed to copy");
    }
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <button
          className="inline-flex items-center justify-center rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
          aria-label="View SDK installation guide"
        >
          <BookOpen className="size-4" />
        </button>
      </DialogTrigger>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <BookOpen className="size-5 text-primary" />
            <span>SDK Installation Guide</span>
          </DialogTitle>
          <DialogDescription>
            Follow these instructions to install the generated SDK in your project.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="go" className="w-full mt-2">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="go">Go SDK</TabsTrigger>
            <TabsTrigger value="js">JavaScript / TypeScript</TabsTrigger>
          </TabsList>

          <TabsContent value="go" className="space-y-4 pt-3">
            <div className="space-y-2">
              <h4 className="text-sm font-semibold">1. Configure Go Environment</h4>
              <p className="text-sm text-muted-foreground">
                Because this registry runs locally or securely without a public certificate, you must tell Go to allow insecure connections and skip checksum verification:
              </p>
              <div className="relative">
                <pre className="p-3 bg-muted rounded-md text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all w-full max-w-full">
                  {`go env -w GOINSECURE=${host}\ngo env -w GONOSUMDB=${host}`}
                </pre>
                <button
                  onClick={() => copyToClipboard(`go env -w GOINSECURE=${host}\ngo env -w GONOSUMDB=${host}`, "go-env")}
                  className="absolute right-2 top-2 p-1.5 rounded-md hover:bg-background text-muted-foreground transition-colors"
                  aria-label="Copy Go env commands"
                >
                  {copiedText === "go-env" ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">2. Get the package</h4>
              <p className="text-sm text-muted-foreground">
                Fetch the package into your project using the vanity import path:
              </p>
              <div className="relative">
                <pre className="p-3 bg-muted rounded-md text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all w-full max-w-full">
                  {`go get ${goImportPath}`}
                </pre>
                <button
                  onClick={() => copyToClipboard(`go get ${goImportPath}`, "go-get")}
                  className="absolute right-2 top-2 p-1.5 rounded-md hover:bg-background text-muted-foreground transition-colors"
                  aria-label="Copy go get command"
                >
                  {copiedText === "go-get" ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">3. Go Module Replacement</h4>
              <p className="text-sm text-muted-foreground">
                If your project uses a different module name or you want to alias the import path, add a replace directive to your <code className="bg-muted px-1 py-0.5 rounded">go.mod</code> file:
              </p>
              <div className="relative">
                <pre className="p-3 bg-muted rounded-md text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all w-full max-w-full">
                  {`replace ${goImportPath} => hasir v0.0.0-${commitHash}`}
                </pre>
                <button
                  onClick={() => copyToClipboard(`replace ${goImportPath} => hasir v0.0.0-${commitHash}`, "go-replace")}
                  className="absolute right-2 top-2 p-1.5 rounded-md hover:bg-background text-muted-foreground transition-colors"
                  aria-label="Copy Go replace command"
                >
                  {copiedText === "go-replace" ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
                </button>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="js" className="space-y-4 pt-3">
            <p className="text-sm text-muted-foreground">
              You can install JavaScript/TypeScript SDKs directly from the registry using Git over HTTP or SSH:
            </p>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">Install via HTTP</h4>
              <div className="relative">
                <pre className="p-3 bg-muted rounded-md text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all w-full max-w-full">
                  {jsHttpCommand}
                </pre>
                <button
                  onClick={() => copyToClipboard(jsHttpCommand, "js-http")}
                  className="absolute right-2 top-2 p-1.5 rounded-md hover:bg-background text-muted-foreground transition-colors"
                  aria-label="Copy NPM HTTP install command"
                >
                  {copiedText === "js-http" ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">Install via SSH (Recommended for Private Repos)</h4>
              <div className="relative">
                <pre className="p-3 bg-muted rounded-md text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all w-full max-w-full">
                  {jsSshCommand}
                </pre>
                <button
                  onClick={() => copyToClipboard(jsSshCommand, "js-ssh")}
                  className="absolute right-2 top-2 p-1.5 rounded-md hover:bg-background text-muted-foreground transition-colors"
                  aria-label="Copy NPM SSH install command"
                >
                  {copiedText === "js-ssh" ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
                </button>
              </div>
            </div>

            <div className="text-xs text-muted-foreground">
              Note: You can replace <code className="bg-muted px-1 py-0.5 rounded">npm install</code> with <code className="bg-muted px-1 py-0.5 rounded">yarn add</code>, <code className="bg-muted px-1 py-0.5 rounded">pnpm add</code>, or <code className="bg-muted px-1 py-0.5 rounded">bun add</code> depending on your package manager.
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
