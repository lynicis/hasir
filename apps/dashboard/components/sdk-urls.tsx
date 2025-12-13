"use client";

import { getOrganizationId } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { Check, ChevronDown, Copy, Package } from "lucide-react";
import { useQuery } from "@connectrpc/connect-query";
import { useState } from "react";
import { toast } from "sonner";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { customRetry } from "@/lib/query-retry";
import { Input } from "@/components/ui/input";

interface SdkUrlsProps {
  repositoryId: string;
}

type SdkType =
  | "go-protobuf"
  | "go-connectrpc"
  | "go-grpc"
  | "js-bufbuild-es"
  | "js-protobuf"
  | "js-connectrpc";

const SDK_OPTIONS: { value: SdkType; label: string; group: string }[] = [
  { value: "go-protobuf", label: "Go / Protocol Buffers", group: "Go" },
  { value: "go-connectrpc", label: "Go / Connect-RPC", group: "Go" },
  { value: "go-grpc", label: "Go / gRPC", group: "Go" },
  { value: "js-bufbuild-es", label: "JS / @bufbuild/es", group: "JavaScript" },
  { value: "js-protobuf", label: "JS / protocolbuffers", group: "JavaScript" },
  { value: "js-connectrpc", label: "JS / @connectrpc", group: "JavaScript" },
];

export function SdkUrls({ repositoryId }: SdkUrlsProps) {
  const [protocol, setProtocol] = useState<"HTTPS" | "SSH">("HTTPS");
  const [sdkType, setSdkType] = useState<SdkType>("go-protobuf");
  const [copied, setCopied] = useState(false);

  const { data: organizationData, isLoading: isLoadingOrganization } = useQuery(
    getOrganizationId,
    { repositoryId },
    { retry: customRetry }
  );

  const organizationId = organizationData?.id;

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const urlObj = new URL(apiUrl);
  const host = urlObj.hostname;
  const port = urlObj.port ? `:${urlObj.port}` : "";

  const sdkPath = organizationId
    ? `/sdk/${organizationId}/${repositoryId}/${sdkType}/`
    : "";

  const url =
    protocol === "HTTPS"
      ? `${apiUrl}${sdkPath}`
      : `ssh://git@${host}${port}${sdkPath}`;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      toast.success("SDK URL copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy URL");
    }
  };

  const selectedSdk = SDK_OPTIONS.find((opt) => opt.value === sdkType);
  const goOptions = SDK_OPTIONS.filter((opt) => opt.group === "Go");
  const jsOptions = SDK_OPTIONS.filter((opt) => opt.group === "JavaScript");

  if (isLoadingOrganization || !organizationId) {
    return (
      <div className="flex w-full flex-col gap-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <Package className="size-4" />
          <span>SDK Download URL</span>
        </div>
        <div className="h-10 w-full animate-pulse rounded-md bg-muted" />
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-3">
      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <Package className="size-4" />
        <span>SDK Download URL</span>
      </div>

      <div className="flex w-full items-center gap-2">
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

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="min-w-[160px] justify-between">
              <span className="truncate">{selectedSdk?.label}</span>
              <ChevronDown className="ml-2 size-4 shrink-0 opacity-50" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-[200px]">
            <DropdownMenuLabel>Go SDKs</DropdownMenuLabel>
            {goOptions.map((option) => (
              <DropdownMenuItem
                key={option.value}
                onClick={() => setSdkType(option.value)}
              >
                {option.label}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuLabel>JavaScript SDKs</DropdownMenuLabel>
            {jsOptions.map((option) => (
              <DropdownMenuItem
                key={option.value}
                onClick={() => setSdkType(option.value)}
              >
                {option.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

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
          aria-label="Copy SDK URL to clipboard"
        >
          {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
        </Button>
      </div>
    </div>
  );
}
