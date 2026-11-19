"use client";

import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@hasir/ui/components/field";
import { getApiKeys } from "@hasir/proto/gen/js/user/v1/user-UserService_connectquery";
import { UserService } from "@hasir/proto/gen/js/user/v1/user_pb";
import { Check, Copy, Key, Plus, Trash2 } from "lucide-react";
import { Pagination } from "@hasir/ui/components/pagination";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@connectrpc/connect-query";
import { Button } from "@hasir/ui/components/button";
import { Input } from "@hasir/ui/components/input";
import { Code } from "@connectrpc/connect";
import { useForm } from "react-hook-form";
import { useState } from "react";
import { toast } from "sonner";
import { z } from "zod/v4";

import { handleConnectErrorToast } from "@/lib/connect-error";
import { keyNameSchema } from "@/lib/validation-schemas";
import { customRetry } from "@/lib/query-retry";
import { useClient } from "@/lib/use-client";

type NewApiKey = {
  id?: string;
  name: string;
  fullKey: string;
};

const apiKeyFormSchema = z.object({
  name: keyNameSchema,
});

type ApiKeyFormData = z.infer<typeof apiKeyFormSchema>;

const PAGE_SIZE = 10;

export function ApiKeysSection() {
  const userApiClient = useClient(UserService);
  const [apiKeyPage, setApiKeyPage] = useState(1);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<NewApiKey | null>(
    null
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

  const apiKeyForm = useForm<ApiKeyFormData>({
    resolver: zodResolver(apiKeyFormSchema),
    mode: "onChange",
    defaultValues: {
      name: "",
    },
  });

  const apiKeys = apiKeysData?.keys || [];
  const totalApiPages = apiKeysData?.totalPage ?? 1;
  const isNewlyCreatedKeyInList = Boolean(
    newlyCreatedKey?.id && apiKeys.some((k) => k.id === newlyCreatedKey.id)
  );

  async function handleGenerateApiKey(data: ApiKeyFormData) {
    let createdKey: NewApiKey;
    try {
      const response = await userApiClient.createApiKey({
        name: data.name,
      });

      if (!response.key) {
        throw new Error("API key was not returned from server");
      }

      createdKey = {
        name: data.name,
        fullKey: response.key,
      };
      setNewlyCreatedKey(createdKey);
      apiKeyForm.reset();
    } catch (error) {
      handleConnectErrorToast(error, {
        [Code.Unauthenticated]:
          "You are not authenticated. Please log in again.",
        [Code.PermissionDenied]:
          "You don't have permission to create API keys.",
        [Code.InvalidArgument]: (err) =>
          err.message || "Invalid request. Please check your input.",
        fallback: "Failed to generate API key. Please try again.",
      });
      return;
    }

    setApiKeyPage(1);
    try {
      const updatedData = await refetchApiKeys();
      const newKey = updatedData.data?.keys?.find((k) => k.name === data.name);
      if (newKey) {
        setNewlyCreatedKey({
          id: newKey.id,
          name: newKey.name,
          fullKey: createdKey.fullKey,
        });
      }
    } catch {
      toast.error("API key created, but failed to refresh the list.", {
        description: "Copy your new API key now before leaving this page.",
      });
    }
  }

  async function handleCopyApiKey(key: { id?: string; name: string }) {
    if (
      newlyCreatedKey &&
      (newlyCreatedKey.id
        ? newlyCreatedKey.id === key.id
        : newlyCreatedKey.name === key.name)
    ) {
      await navigator.clipboard.writeText(newlyCreatedKey.fullKey);
      const copyTargetId = key.id ?? newlyCreatedKey.name;
      setCopiedId(copyTargetId);
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
      handleConnectErrorToast(error, {
        [Code.Unauthenticated]:
          "You are not authenticated. Please log in again.",
        [Code.PermissionDenied]:
          "You don't have permission to delete API keys.",
        [Code.NotFound]: "API key not found.",
        fallback: "Failed to delete API key. Please try again.",
      });
    }
  }

  return (
    <section aria-labelledby="api-key-heading" className="space-y-4">
      <div className="space-y-1">
        <h3
          id="api-key-heading"
          className="text-base font-semibold leading-none"
        >
          API keys
        </h3>
        <p className="text-sm text-muted-foreground">
          Generate personal API keys for programmatic access. Keys are securely
          stored on the backend.
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

      {newlyCreatedKey && !isNewlyCreatedKeyInList && (
        <div className="group flex items-center justify-between gap-4 rounded-md border bg-card px-4 py-3 transition-colors">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="font-medium">{newlyCreatedKey.name}</p>
              <span className="rounded-md bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900 dark:text-green-300">
                New
              </span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <code className="rounded-md bg-muted px-3 py-1.5 font-mono text-xs font-medium">
              {newlyCreatedKey.fullKey}
            </code>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={() => handleCopyApiKey(newlyCreatedKey)}
            >
              {copiedId === (newlyCreatedKey.id ?? newlyCreatedKey.name) ? (
                <Check className="text-green-600" />
              ) : (
                <Copy />
              )}
              <span className="sr-only">Copy API key</span>
            </Button>
          </div>
        </div>
      )}

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
            Generate your first API key to get started with programmatic access.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {apiKeys.map((key) => {
            const isNewlyCreated = newlyCreatedKey?.id === key.id;

            return (
              <div
                key={key.id}
                className="group flex items-center justify-between gap-4 rounded-md border bg-card px-4 py-3 transition-colors"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{key.name}</p>
                    {isNewlyCreated && (
                      <span className="rounded-md bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900 dark:text-green-300">
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
  );
}
