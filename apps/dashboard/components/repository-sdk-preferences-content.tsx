"use client";

import { getRecentCommit } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { RegistryService, SDK } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { useContext, useEffect, useMemo, useState } from "react";
import { useQuery } from "@connectrpc/connect-query";
import { AlertTriangle, Wrench } from "lucide-react";
import { useParams } from "next/navigation";
import { toast } from "sonner";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";
import { SdkInstallGuideDialog } from "@/components/sdk-install-guide-dialog";
import { RepositoryContext } from "@/lib/repository-context";
import { Skeleton } from "@/components/ui/skeleton";
import { SdkUrls } from "@/components/sdk-urls";
import { Switch } from "@/components/ui/switch";
import { customRetry } from "@/lib/query-retry";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { useClient } from "@/lib/use-client";

interface SdkOption {
  key: string;
  label: string;
  htmlId: string;
  sdkValue: SDK;
}

interface LanguageConfig {
  name: string;
  description: string;
  options: SdkOption[];
}

type SdkConfig = Record<
  string,
  {
    enabled: boolean;
    [key: string]: boolean;
  }
>;

const SDK_LANGUAGES: Record<string, LanguageConfig> = {
  go: {
    name: "Go",
    description: "Generate Go SDK with multiple protocol support options",
    options: [
      {
        key: "protocolBuffers",
        label: "Protocol Buffers",
        htmlId: "go-protocol-buffers",
        sdkValue: SDK.SDK_GO_PROTOBUF,
      },
      {
        key: "connectRpc",
        label: "Connect-RPC",
        htmlId: "go-connect-rpc",
        sdkValue: SDK.SDK_GO_CONNECTRPC,
      },
      {
        key: "grpc",
        label: "gRPC",
        htmlId: "go-grpc",
        sdkValue: SDK.SDK_GO_GRPC,
      },
    ],
  },
  javascript: {
    name: "JavaScript/Typescript",
    description:
      "Generate JavaScript SDK with various protocol implementations",
    options: [
      {
        key: "bufbuildEs",
        label: "@bufbuild/es",
        htmlId: "javascript-bufbuild-es",
        sdkValue: SDK.SDK_JS_BUFBUILD_ES,
      },
      {
        key: "protocolBuffersJs",
        label: "protocolbuffers/js",
        htmlId: "javascript-protobuffers-js",
        sdkValue: SDK.SDK_JS_PROTOBUF,
      },
      {
        key: "connectRpc",
        label: "@connectrpc/connectrpc",
        htmlId: "javascript-connectrpc",
        sdkValue: SDK.SDK_JS_CONNECTRPC,
      },
    ],
  },
  rust: {
    name: "Rust",
    description: "Generate Rust SDK using Protocol Buffers",
    options: [
      {
        key: "protocolBuffers",
        label: "Protocol Buffers",
        htmlId: "rust-protocol-buffers",
        sdkValue: SDK.SDK_RUST_PROTOBUF,
      },
    ],
  },
  java: {
    name: "Java",
    description: "Generate Java SDK using Protocol Buffers",
    options: [
      {
        key: "protocolBuffers",
        label: "Protocol Buffers",
        htmlId: "java-protocol-buffers",
        sdkValue: SDK.SDK_JAVA_PROTOBUF,
      },
    ],
  },
  csharp: {
    name: "C#",
    description: "Generate C# SDK using Protocol Buffers",
    options: [
      {
        key: "protocolBuffers",
        label: "Protocol Buffers",
        htmlId: "csharp-protocol-buffers",
        sdkValue: SDK.SDK_CSHARP_PROTOBUF,
      },
    ],
  },
};

const getInitialConfig = (): SdkConfig => {
  const config: SdkConfig = {};

  Object.entries(SDK_LANGUAGES).forEach(([langKey, langConfig]) => {
    const langObj = { enabled: false } as { enabled: boolean; [key: string]: boolean };
    langConfig.options.forEach((option) => {
      langObj[option.key] = false;
    });
    config[langKey] = langObj;
  });

  return config;
};

const mapSdkPreferencesToConfig = (
  sdkPreferences: Array<{ sdk: SDK; status: boolean }>
): SdkConfig => {
  const config = getInitialConfig();

  sdkPreferences.forEach((pref) => {
    Object.entries(SDK_LANGUAGES).forEach(([langKey, langConfig]) => {
      langConfig.options.forEach((option) => {
        if (option.sdkValue === pref.sdk && pref.status) {
          const lang = config[langKey];
          if (lang) {
            lang[option.key] = true;
            lang.enabled = true;
          }
        }
      });
    });
  });

  return config;
};

const configsAreEqual = (a: SdkConfig, b: SdkConfig): boolean => {
  const keys = Object.keys(a);
  if (keys.length !== Object.keys(b).length) return false;

  for (const key of keys) {
    const aObj = a[key];
    const bObj = b[key];
    if (!aObj || !bObj) return false;

    const objKeys = Object.keys(aObj);
    if (objKeys.length !== Object.keys(bObj).length) return false;

    for (const objKey of objKeys) {
      if (aObj[objKey] !== bObj[objKey]) return false;
    }
  }

  return true;
};

export default function RepositorySdkPreferencesContent() {
  const params = useParams();
  const repositoryId = params.repositoryId as string;
  const registryApiClient = useClient(RegistryService);
  const context = useContext(RepositoryContext);

  if (!context) {
    throw new Error("SdkPreferencesPage must be used within RepositoryLayout");
  }

  const { repository, isLoading, error } = context;
  const isManagedByBuf = repository?.managedByBuf ?? false;

  const { data: recentCommit } = useQuery(
    getRecentCommit,
    { repositoryId },
    { retry: customRetry }
  );

  const commitHash = recentCommit?.id ?? "latest";

  const serverConfig = useMemo(() => {
    if (repository?.sdkPreferences !== undefined) {
      return mapSdkPreferencesToConfig(repository.sdkPreferences);
    }
    return getInitialConfig();
  }, [repository]);

  const [config, setConfig] = useState<SdkConfig>(serverConfig);
  const [prevServerConfig, setPrevServerConfig] = useState(serverConfig);
  const [isSaving, setIsSaving] = useState(false);

  if (!isSaving && serverConfig !== prevServerConfig) {
    setPrevServerConfig(serverConfig);
    setConfig(serverConfig);
  }

  const hasChanges = useMemo(() => !configsAreEqual(config, serverConfig), [config, serverConfig]);

  useEffect(() => {
    if (error && !repository) {
      toast.error("Failed to load repository data");
    }
  }, [error, repository]);

  const handleLanguageToggle = (language: string) => {
    if (isManagedByBuf) return;
    setConfig((prev) => {
      const current = prev[language] || { enabled: false };
      return {
        ...prev,
        [language]: {
          ...current,
          enabled: !current.enabled,
          ...(current.enabled &&
            Object.keys(current).reduce((acc, key) => {
              if (key !== "enabled") acc[key] = false;
              return acc;
            }, {} as Record<string, boolean>)),
        },
      };
    });
  };

  const handleSubOptionToggle = (language: string, option: string) => {
    if (isManagedByBuf) return;
    setConfig((prev) => {
      const current = prev[language] || { enabled: false };
      const currentValue = (current as Record<string, boolean>)[option] || false;
      const newValue = !currentValue;

      return {
        ...prev,
        [language]: {
          ...current,
          [option]: newValue,
          enabled: newValue ? true : current.enabled,
        },
      };
    });
  };

  const handleSave = async () => {
    try {
      setIsSaving(true);

      const sdkPreferences: Array<{ sdk: SDK; status: boolean }> = [];
      Object.entries(SDK_LANGUAGES).forEach(([langKey, langConfig]) => {
        langConfig.options.forEach((option) => {
          const isEnabled = (config[langKey]?.[option.key] as boolean) ?? false;
          sdkPreferences.push({ sdk: option.sdkValue, status: isEnabled });
        });
      });

      await registryApiClient.updateSdkPreferences({
        id: repositoryId,
        sdkPreferences,
      });

      setConfig(serverConfig);
      toast.success("SDK preferences saved successfully");
    } catch {
      toast.error("Failed to save SDK preferences");
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    setConfig(serverConfig);
  };

  const isSubmitting = isLoading || isSaving;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <div className="flex items-center gap-2">
            <Wrench className="size-5 text-muted-foreground" />
            <h2 className="text-xl font-semibold">
              SDK Generation Preferences
            </h2>
          </div>
          <p className="text-muted-foreground mt-1 text-sm">
            Configure which SDKs to generate for this repository
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2].map((i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-6 w-32" />
                <Skeleton className="mt-2 h-4 w-full" />
              </CardHeader>
              <CardContent className="space-y-3">
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2">
          <Wrench className="size-5 text-muted-foreground" />
          <h2 className="text-xl font-semibold">SDK Generation Preferences</h2>
        </div>
        <p className="text-muted-foreground mt-1 text-sm">
          Configure which SDKs to generate for this repository
        </p>
      </div>

      {isManagedByBuf && (
        <Alert role="alert">
          <AlertTriangle className="size-4" />
          <AlertTitle>Managed by Buf</AlertTitle>
          <AlertDescription>
            SDK generation preferences are configured externally and cannot be
            modified here. Use the Buf Schema Registry to configure SDK
            generation for this repository.
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {Object.entries(SDK_LANGUAGES).map(([langKey, langConfig]) => (
          <Card key={langKey}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>{langConfig.name}</span>
                <Switch
                  checked={config[langKey]?.enabled || false}
                  onCheckedChange={() => handleLanguageToggle(langKey)}
                  disabled={isManagedByBuf || isSubmitting}
                />
              </CardTitle>
              <CardDescription>{langConfig.description}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                {langConfig.options.map((option) => (
                  <div
                    key={option.key}
                    className="flex items-center justify-between"
                  >
                    <Label
                      htmlFor={option.htmlId}
                      className={`text-sm ${!config[langKey]?.enabled ? "text-muted-foreground" : ""
                        }`}
                    >
                      {option.label}
                    </Label>
                    <Switch
                      id={option.htmlId}
                      checked={
                        (config[langKey]?.[option.key] as boolean) || false
                      }
                      onCheckedChange={() =>
                        handleSubOptionToggle(langKey, option.key)
                      }
                      disabled={
                        isManagedByBuf ||
                        !config[langKey]?.enabled ||
                        isSubmitting
                      }
                    />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {!isManagedByBuf && (
        <div className="flex items-center gap-3">
          <Button
            type="button"
            onClick={handleSave}
            disabled={!hasChanges || isSaving}
            isLoading={isSaving}
          >
            Save Configuration
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={handleReset}
            disabled={!hasChanges || isSaving}
          >
            Reset to Defaults
          </Button>
        </div>
      )}

      {repository && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Download SDK</span>
              <SdkInstallGuideDialog
                organizationId={repository.organizationId}
                repositoryId={repositoryId}
                commitHash={commitHash}
              />
            </CardTitle>
            <CardDescription>
              Copy the SDK URL to install the generated SDK package
            </CardDescription>
          </CardHeader>
          <CardContent>
            <SdkUrls
              organizationId={repository.organizationId}
              repositoryId={repositoryId}
              commitHash={commitHash}
            />
          </CardContent>
        </Card>
      )}
    </div>
  );
}
