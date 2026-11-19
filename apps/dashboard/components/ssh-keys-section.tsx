"use client";

import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@hasir/ui/components/field";
import { getSshKeys } from "@hasir/proto/gen/js/user/v1/user-UserService_connectquery";
import { UserService } from "@hasir/proto/gen/js/user/v1/user_pb";
import { Pagination } from "@hasir/ui/components/pagination";
import { Textarea } from "@hasir/ui/components/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@connectrpc/connect-query";
import { Button } from "@hasir/ui/components/button";
import { Input } from "@hasir/ui/components/input";
import { Key, Plus, Trash2 } from "lucide-react";
import { Code } from "@connectrpc/connect";
import { useForm } from "react-hook-form";
import { useState } from "react";
import { toast } from "sonner";
import { z } from "zod/v4";

import {
  keyNameSchema,
  sshPublicKeySchema,
} from "@/lib/validation-schemas";
import { handleConnectErrorToast } from "@/lib/connect-error";
import { customRetry } from "@/lib/query-retry";
import { useClient } from "@/lib/use-client";

const sshKeyFormSchema = z.object({
  name: keyNameSchema,
  publicKey: sshPublicKeySchema,
});

type SshKeyFormData = z.infer<typeof sshKeyFormSchema>;

const PAGE_SIZE = 10;

export function SshKeysSection() {
  const userApiClient = useClient(UserService);
  const [sshKeyPage, setSshKeyPage] = useState(1);

  const {
    data: sshKeysData,
    isLoading: isLoadingSshKeys,
    refetch: refetchSshKeys,
  } = useQuery(
    getSshKeys,
    { page: sshKeyPage, pageLimit: PAGE_SIZE },
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

  const sshKeys = sshKeysData?.keys || [];
  const totalSshPages = sshKeysData?.totalPage ?? 1;

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
      handleConnectErrorToast(error, {
        [Code.Unauthenticated]:
          "You are not authenticated. Please log in again.",
        [Code.PermissionDenied]:
          "You don't have permission to create SSH keys.",
        [Code.InvalidArgument]: (err) =>
          err.message || "Invalid request. Please check your input.",
        fallback: "Failed to add SSH key. Please try again.",
      });
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
      handleConnectErrorToast(error, {
        [Code.Unauthenticated]:
          "You are not authenticated. Please log in again.",
        [Code.PermissionDenied]:
          "You don't have permission to delete SSH keys.",
        [Code.NotFound]: "SSH key not found.",
        fallback: "Failed to delete SSH key. Please try again.",
      });
    }
  }

  return (
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
              className="group flex items-center justify-between gap-4 rounded-md border bg-card px-4 py-3 transition-colors"
            >
              <div className="min-w-0 flex-1 space-y-1">
                <div className="flex items-center gap-2">
                  <p className="font-medium">{key.name}</p>
                </div>
                <p className="font-mono text-xs text-muted-foreground">
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
  );
}
