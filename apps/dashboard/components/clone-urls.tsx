"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface CloneUrlsProps {
  repositoryName: string;
  repositoryId: string;
}

export function CloneUrls({ repositoryName, repositoryId }: CloneUrlsProps) {
  const [protocol, setProtocol] = useState<"HTTPS" | "SSH">("HTTPS");
  const [copied, setCopied] = useState(false);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  const host = apiUrl.replace(/^https?:\/\//, "").split(":")[0];

  const url =
    protocol === "HTTPS"
      ? `${apiUrl}/git/${repositoryId}.git`
      : `git@${host}:${repositoryName}.git`;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy URL");
    }
  };

  return (
    <div className="flex w-full max-w-lg items-center gap-2">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="w-[80px]">
            {protocol}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={() => setProtocol("HTTPS")}>
            HTTPS
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setProtocol("SSH")}>
            SSH
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <div className="relative flex-1">
        <Input
          readOnly
          value={url}
          className="pr-20 font-mono text-sm"
          onClick={(e) => e.currentTarget.select()}
        />
        <Button
          size="icon"
          variant="ghost"
          className="absolute right-0 top-0 h-full w-10 text-muted-foreground hover:text-foreground"
          onClick={handleCopy}
          aria-label="Copy to clipboard"
        >
          {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
        </Button>
      </div>
    </div>
  );
}
