"use client";

import { getApiKeys } from "@buf/hasir_hasir.connectrpc_query-es/user/v1/user-UserService_connectquery";
import { getSshKeys } from "@buf/hasir_hasir.connectrpc_query-es/user/v1/user-UserService_connectquery";
import { UserService } from "@buf/hasir_hasir.bufbuild_es/user/v1/user_pb";
import { Check, Copy, Key, Plus, Trash2 } from "lucide-react";
import { Code, ConnectError } from "@connectrpc/connect";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@connectrpc/connect-query";
import { useForm } from "react-hook-form";
import { useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Pagination } from "@/components/ui/pagination";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { customRetry } from "@/lib/query-retry";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";

type NewApiKey = {
  id: string;
  name: string;
  fullKey: string;
};

const sshKeyFormSchema = z.object({
  name: z
    .string()
    .min(1, "Key name is required")
    .max(50, "Key name must be 50 characters or less"),
  publicKey: z
    .string()
    .min(1, "Public key is required")
    .refine(
      (key) => key.startsWith("ssh-") || key.startsWith("ecdsa-"),
      "Must be a valid SSH public key (should start with ssh- or ecdsa-)"
    ),
});

const apiKeyFormSchema = z.object({
  name: z
    .string()
    .min(1, "Key name is required")
    .max(50, "Key name must be 50 characters or less"),
});

type SshKeyFormData = z.infer<typeof sshKeyFormSchema>;
type ApiKeyFormData = z.infer<typeof apiKeyFormSchema>;

const PAGE_SIZE = 10;

export function SshApiKeyPanel() {
  const userApiClient = useClient(UserService);

  const [sshKeyPage, setSshKeyPage] = useState(1);
  const [apiKeyPage, setApiKeyPage] = useState(1);

  const {
    data: sshKeysData,
    isLoading: isLoadingSshKeys,
    refetch: refetchSshKeys,
  } = useQuery(
    getSshKeys,
    { page: sshKeyPage, pageLimit: PAGE_SIZE },
    { retry: customRetry }
  );

  const {
    data: apiKeysData,
    isLoading: isLoadingApiKeys,
    refetch: refetchApiKeys,
  } = useQuery(
    getApiKeys,
    { page: apiKeyPage, pageLimit: PAGE_SIZE },
    { retry: customRetry }
  );

  const sshKeyForm = useForm<SshKeyFormData>({
    resolver: zodResolver(sshKeyFormSchema),
    mode: "onChange",
    defaultValues: {
      name: "",
      publicKey: "",
    },
  });

  const apiKeyForm = useForm<ApiKeyFormData>({
    resolver: zodResolver(apiKeyFormSchema),
    mode: "onChange",
    defaultValues: {
      name: "",
    },
  });

  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<NewApiKey | null>(
    null
  );

  const sshKeys = sshKeysData?.keys || [];
  const apiKeys = apiKeysData?.keys || [];

  const totalSshPages = sshKeysData?.totalPage ?? 1;
  const totalApiPages = apiKeysData?.totalPage ?? 1;

  async function handleAddSshKey(data: SshKeyFormData) {
    try {
      await userApiClient.createSshKey({
        name: data.name,
        publicKey: data.publicKey,
      });

      setSshKeyPage(1);
      await refetchSshKeys();

      sshKeyForm.reset();

      toast.success("SSH key added successfully", {
        description: "Your SSH key has been saved.",
      });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("You are not authenticated. Please log in again.");
          return;
        }

        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to create SSH keys.");
          return;
        }

        if (error.code === Code.InvalidArgument) {
          toast.error(
            error.message || "Invalid request. Please check your input."
          );
          return;
        }
      }

      toast.error("Failed to add SSH key. Please try again.");
    }
  }

  async function handleGenerateApiKey(data: ApiKeyFormData) {
    try {
      const response = await userApiClient.createApiKey({
        name: data.name,
      });

      if (!response.key) {
        throw new Error("API key was not returned from server");
      }

      setApiKeyPage(1);
      const updatedData = await refetchApiKeys();

      const newKey = updatedData.data?.keys?.find((k) => k.name === data.name);
      if (newKey) {
        setNewlyCreatedKey({
          id: newKey.id,
          name: newKey.name,
          fullKey: response.key,
        });
      }

      apiKeyForm.reset();

      toast.success("API key generated", {
        description: `Your new API key: ${response.key}. Copy it now - you won't be able to see it again!`,
        duration: 10000,
      });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("You are not authenticated. Please log in again.");
          return;
        }

        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to create API keys.");
          return;
        }

        if (error.code === Code.InvalidArgument) {
          toast.error(
            error.message || "Invalid request. Please check your input."
          );
          return;
        }
      }

      toast.error("Failed to generate API key. Please try again.");
    }
  }

  async function handleDeleteSshKey(keyId: string) {
    try {
      await userApiClient.revokeSshKey({
        id: keyId,
      });

      if (sshKeys.length === 1 && sshKeyPage > 1) {
        setSshKeyPage(sshKeyPage - 1);
      }

      await refetchSshKeys();

      toast.success("SSH key deleted", {
        description: "The SSH key has been removed.",
      });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("You are not authenticated. Please log in again.");
          return;
        }

        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete SSH keys.");
          return;
        }

        if (error.code === Code.NotFound) {
          toast.error("SSH key not found.");
          return;
        }
      }

      toast.error("Failed to delete SSH key. Please try again.");
    }
  }

  async function handleCopyApiKey(key: { id: string; name: string }) {
    if (newlyCreatedKey && newlyCreatedKey.id === key.id) {
      navigator.clipboard.writeText(newlyCreatedKey.fullKey);
      setCopiedId(key.id);
      toast.success("Copied to clipboard", {
        description: `${key.name} has been copied.`,
      });
      setNewlyCreatedKey(null);

      setTimeout(() => {
        setCopiedId(null);
      }, 2000);
    }
  }

  async function handleDeleteApiKey(keyId: string) {
    try {
      await userApiClient.revokeApiKey({
        id: keyId,
      });

      if (apiKeys.length === 1 && apiKeyPage > 1) {
        setApiKeyPage(apiKeyPage - 1);
      }

      await refetchApiKeys();

      if (newlyCreatedKey?.id === keyId) {
        setNewlyCreatedKey(null);
      }

      toast.success("API key deleted", {
        description: "The API key has been removed.",
      });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("You are not authenticated. Please log in again.");
          return;
        }

        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete API keys.");
          return;
        }

        if (error.code === Code.NotFound) {
          toast.error("API key not found.");
          return;
        }
      }

      toast.error("Failed to delete API key. Please try again.");
    }
  }

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
        <section aria-labelledby="ssh-key-heading" className="space-y-4">
          <div className="space-y-1">
            <h3
              id="ssh-key-heading"
              className="text-base font-semibold leading-none"
            >
              SSH public keys
            </h3>
            <p className="text-sm text-muted-foreground">
              Add SSH public keys for secure authentication. Keys are securely
              stored on the backend.
            </p>
          </div>

          <form
            onSubmit={sshKeyForm.handleSubmit(handleAddSshKey)}
            className="space-y-4"
          >
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="ssh-key-name">Key name</FieldLabel>
                <Input
                  id="ssh-key-name"
                  placeholder="e.g., Work Laptop, Personal MacBook"
                  maxLength={50}
                  disabled={sshKeyForm.formState.isSubmitting}
                  {...sshKeyForm.register("name")}
                />
                {sshKeyForm.formState.errors.name && (
                  <p className="text-sm text-destructive">
                    {sshKeyForm.formState.errors.name.message}
                  </p>
                )}
                <FieldDescription>
                  A friendly name to identify this SSH key.
                </FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor="ssh-public-key">Public key</FieldLabel>
                <Textarea
                  id="ssh-public-key"
                  placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@example.com"
                  rows={4}
                  spellCheck={false}
                  className="font-mono text-xs"
                  disabled={sshKeyForm.formState.isSubmitting}
                  {...sshKeyForm.register("publicKey")}
                />
                {sshKeyForm.formState.errors.publicKey && (
                  <p className="text-sm text-destructive">
                    {sshKeyForm.formState.errors.publicKey.message}
                  </p>
                )}
                <FieldDescription>
                  Paste your SSH <strong>public</strong> key (never your private
                  key).
                </FieldDescription>
              </Field>
              <Field>
                <Button
                  type="submit"
                  disabled={
                    !sshKeyForm.formState.isValid ||
                    sshKeyForm.formState.isSubmitting
                  }
                  isLoading={sshKeyForm.formState.isSubmitting}
                >
                  <Plus />
                  Add SSH key
                </Button>
              </Field>
            </FieldGroup>
          </form>

          {isLoadingSshKeys ? (
            <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 px-6 py-12 text-center">
              <div className="rounded-full bg-muted p-3">
                <Key className="size-6 text-muted-foreground animate-pulse" />
              </div>
              <h4 className="mt-4 font-medium">Loading SSH keys...</h4>
            </div>
          ) : sshKeys.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 px-6 py-12 text-center">
              <div className="rounded-full bg-muted p-3">
                <Key className="size-6 text-muted-foreground" />
              </div>
              <h4 className="mt-4 font-medium">No SSH keys yet</h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Add your first SSH public key to enable secure authentication.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {sshKeys.map((key) => (
                <div
                  key={key.id}
                  className="group flex items-center justify-between gap-4 rounded-lg border bg-card px-4 py-3 shadow-sm transition-all hover:shadow-md"
                >
                  <div className="min-w-0 flex-1 space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="font-medium">{key.name}</p>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Key ID: {key.id}
                    </p>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => handleDeleteSshKey(key.id)}
                      className="opacity-0 transition-opacity group-hover:opacity-100"
                    >
                      <Trash2 />
                      <span className="sr-only">Delete SSH key</span>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {!isLoadingSshKeys && sshKeys.length > 0 && (
            <Pagination
              currentPage={sshKeyPage}
              totalPages={totalSshPages}
              onPageChange={setSshKeyPage}
              disabled={isLoadingSshKeys}
            />
          )}
        </section>

        <Separator />

        <section aria-labelledby="api-key-heading" className="space-y-4">
          <div className="space-y-1">
            <h3
              id="api-key-heading"
              className="text-base font-semibold leading-none"
            >
              API keys
            </h3>
            <p className="text-sm text-muted-foreground">
              Generate personal API keys for programmatic access. Keys are
              securely stored on the backend.
            </p>
          </div>

          <form
            onSubmit={apiKeyForm.handleSubmit(handleGenerateApiKey)}
            className="space-y-4"
          >
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="api-key-name">Key name</FieldLabel>
                <Input
                  id="api-key-name"
                  placeholder="e.g., Production API, Development Key"
                  maxLength={50}
                  disabled={apiKeyForm.formState.isSubmitting}
                  {...apiKeyForm.register("name")}
                />
                {apiKeyForm.formState.errors.name && (
                  <p className="text-sm text-destructive">
                    {apiKeyForm.formState.errors.name.message}
                  </p>
                )}
                <FieldDescription>
                  A friendly name to identify this API key.
                </FieldDescription>
              </Field>
              <Field>
                <Button
                  type="submit"
                  variant="outline"
                  size="sm"
                  disabled={
                    !apiKeyForm.formState.isValid ||
                    apiKeyForm.formState.isSubmitting
                  }
                  isLoading={apiKeyForm.formState.isSubmitting}
                >
                  <Plus />
                  Generate key
                </Button>
              </Field>
            </FieldGroup>
          </form>

          {isLoadingApiKeys ? (
            <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 px-6 py-12 text-center">
              <div className="rounded-full bg-muted p-3">
                <Key className="size-6 text-muted-foreground animate-pulse" />
              </div>
              <h4 className="mt-4 font-medium">Loading API keys...</h4>
            </div>
          ) : apiKeys.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 px-6 py-12 text-center">
              <div className="rounded-full bg-muted p-3">
                <Key className="size-6 text-muted-foreground" />
              </div>
              <h4 className="mt-4 font-medium">No API keys yet</h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Generate your first API key to get started with programmatic
                access.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {apiKeys.map((key) => {
                const isNewlyCreated = newlyCreatedKey?.id === key.id;

                return (
                  <div
                    key={key.id}
                    className="group flex items-center justify-between gap-4 rounded-lg border bg-card px-4 py-3 shadow-sm transition-all hover:shadow-md"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{key.name}</p>
                        {isNewlyCreated && (
                          <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900 dark:text-green-300">
                            New
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {isNewlyCreated ? (
                        <>
                          <code className="rounded-md bg-muted px-3 py-1.5 font-mono text-xs font-medium">
                            {newlyCreatedKey.fullKey}
                          </code>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => handleCopyApiKey(key)}
                            className="opacity-0 transition-opacity group-hover:opacity-100"
                          >
                            {copiedId === key.id ? (
                              <Check className="text-green-600" />
                            ) : (
                              <Copy />
                            )}
                            <span className="sr-only">Copy API key</span>
                          </Button>
                        </>
                      ) : (
                        <p className="text-xs italic text-muted-foreground">
                          Key is hidden for security
                        </p>
                      )}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => handleDeleteApiKey(key.id)}
                        className="opacity-0 transition-opacity group-hover:opacity-100"
                      >
                        <Trash2 />
                        <span className="sr-only">Delete API key</span>
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {!isLoadingApiKeys && apiKeys.length > 0 && (
            <Pagination
              currentPage={apiKeyPage}
              totalPages={totalApiPages}
              onPageChange={setApiKeyPage}
              disabled={isLoadingApiKeys}
            />
          )}
        </section>
      </CardContent>
    </Card>
  );
}
