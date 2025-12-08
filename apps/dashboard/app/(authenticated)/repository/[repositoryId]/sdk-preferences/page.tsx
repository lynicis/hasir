"use client";

import { RegistryService, SDK } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { useContext, useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Wrench } from "lucide-react";
import { toast } from "sonner";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { useClient } from "@/lib/use-client";

import { RepositoryContext } from "../layout";

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
    description: "Generate JavaScript SDK with various protocol implementations",
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
};

const getInitialConfig = (): SdkConfig => {
  const config: SdkConfig = {};

  Object.entries(SDK_LANGUAGES).forEach(([langKey, langConfig]) => {
    config[langKey] = { enabled: false };
    langConfig.options.forEach((option) => {
      config[langKey][option.key] = false;
    });
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
          config[langKey][option.key] = true;
          config[langKey].enabled = true;
        }
      });
    });
  });

  return config;
};

export default function SdkPreferencesPage() {
  const params = useParams();
  const repositoryId = params.repositoryId as string;
  const context = useContext(RepositoryContext);
  const registryApiClient = useClient(RegistryService);

  if (!context) {
    throw new Error(
      "SdkPreferencesPage must be used within RepositoryLayout"
    );
  }

  const { repository, isLoading, error } = context;

  const serverConfig = useMemo(() => {
    if (repository?.sdkPreferences !== undefined) {
      return mapSdkPreferencesToConfig(repository.sdkPreferences);
    }
    return getInitialConfig();
  }, [repository]);

  const [config, setConfig] = useState<SdkConfig>(serverConfig);
  const [prevServerConfig, setPrevServerConfig] = useState(serverConfig);
  const [isSaving, setIsSaving] = useState(false);

  if (serverConfig !== prevServerConfig) {
    setPrevServerConfig(serverConfig);
    setConfig(serverConfig);
  }

  useEffect(() => {
    if (error && !repository) {
      toast.error("Failed to load repository data");
    }
  }, [error, repository]);

  const handleLanguageToggle = (language: string) => {
    setConfig((prev) => ({
      ...prev,
      [language]: {
        ...prev[language],
        enabled: !prev[language].enabled,
        ...(prev[language].enabled &&
          Object.keys(prev[language]).reduce(
            (acc, key) => {
              if (key !== "enabled") acc[key] = false;
              return acc;
            },
            {} as Record<string, boolean>
          )),
      },
    }));
  };

  const handleSubOptionToggle = (language: string, option: string) => {
    setConfig((prev) => {
      const currentValue = (prev[language] as Record<string, boolean>)[option];
      const newValue = !currentValue;

      return {
        ...prev,
        [language]: {
          ...prev[language],
          [option]: newValue,
          enabled: newValue ? true : prev[language].enabled,
        },
      };
    });
  };

  const getSdkPreferences = () => {
    const enabledSdkValues = new Set<SDK>();
    Object.entries(SDK_LANGUAGES).forEach(([langKey, langConfig]) => {
      langConfig.options.forEach((option) => {
        if (config[langKey]?.[option.key]) {
          enabledSdkValues.add(option.sdkValue);
        }
      });
    });

    const preferences: Array<{ sdk: SDK; status: boolean }> = [];

    Object.values(SDK)
      .filter((val): val is SDK => typeof val === "number" && val !== SDK.SDK_UNSPECIFIED)
      .forEach((sdkVal) => {
        preferences.push({
          sdk: sdkVal,
          status: enabledSdkValues.has(sdkVal) ?? false,
        });
      });

    return preferences;
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const sdkPreferences = getSdkPreferences();

      await registryApiClient.updateSdkPreferences({
        id: repositoryId,
        sdkPreferences,
      });

      toast.success("SDK preferences saved successfully");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to save SDK preferences");
    } finally {
        setIsSaving(false);
    }
  };

  const handleReset = () => {
    if (repository?.sdkPreferences) {
      const mappedConfig = mapSdkPreferencesToConfig(repository.sdkPreferences);
      setConfig(mappedConfig);
    } else {
      setConfig(getInitialConfig());
    }
    toast.info("Preferences reset to saved values");
  };

  if (isLoading) {
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

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {Object.entries(SDK_LANGUAGES).map(([langKey, langConfig]) => (
          <Card key={langKey}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>{langConfig.name}</span>
                <Switch
                  checked={config[langKey]?.enabled || false}
                  onCheckedChange={() => handleLanguageToggle(langKey)}
                  disabled={isLoading || isSaving}
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
                      className={`text-sm ${
                        !config[langKey]?.enabled ? "text-muted-foreground" : ""
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
                        !config[langKey]?.enabled ||
                        isLoading ||
                        isSaving
                      }
                    />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <Button
          onClick={handleSave}
          isLoading={isSaving}
          disabled={isSaving}
        >
          Save Configuration
        </Button>
        <Button
          variant="outline"
          onClick={handleReset}
          disabled={isSaving}
        >
          Reset to Defaults
        </Button>
      </div>

      <div className="text-muted-foreground rounded-lg border border-border bg-muted/50 p-4 text-sm">
        <p className="font-medium">Available SDK Targets:</p>
        <ul className="ml-4 mt-2 list-disc space-y-1">
          {Object.entries(SDK_LANGUAGES).map(([langKey, langConfig]) => (
            <li key={langKey}>
              <strong>{langConfig.name}:</strong>{" "}
              {langConfig.options.map((opt) => opt.label).join(", ")}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
