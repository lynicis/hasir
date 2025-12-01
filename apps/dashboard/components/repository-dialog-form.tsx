"use client";

import { getOrganizations } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { RegistryService } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect } from "react";
import { toast } from "sonner";
import { z } from "zod/v4";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { visibilityMapper } from "@/lib/visibility-mapper";
import { useRefreshStore } from "@/stores/refresh-store";
import { Button } from "@/components/ui/button";
import { customRetry } from "@/lib/query-retry";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";

const repositorySchema = z.object({
  name: z
    .string()
    .min(1, { error: "Please enter a repository name." })
    .max(100, { error: "Repository name must be at most 100 characters." }),
  organizationId: z
    .string()
    .min(1, { error: "Please select an organization." }),
  visibility: z.enum(["public", "private"], {
    error: "Please select a visibility.",
  }),
});

export type RepositoryFormValues = z.infer<typeof repositorySchema>;

interface RepositoryDialogFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCancel: () => void;
}

export function RepositoryDialogForm({
  open,
  onOpenChange,
  onCancel,
}: RepositoryDialogFormProps) {
  const registryApiClient = useClient(RegistryService);
  const refreshRepositories = useRefreshStore(
    (state) => state.refreshRepositories
  );
  const organizationsRefreshKey = useRefreshStore(
    (state) => state.organizationsRefreshKey
  );

  const {
    data: organizationsData,
    isLoading: isLoadingOrganizations,
    refetch: refetchOrganizations,
  } = useQuery(
    getOrganizations,
    {
      pagination: {
        page: 1,
        pageLimit: 100,
      },
    },
    { retry: customRetry }
  );

  const organizations = organizationsData?.organizations ?? [];

  useEffect(() => {
    if (organizationsRefreshKey > 0) {
      refetchOrganizations();
    }
  }, [organizationsRefreshKey, refetchOrganizations]);

  const {
    control,
    handleSubmit,
    formState: { isSubmitting },
    reset,
  } = useForm<RepositoryFormValues>({
    resolver: zodResolver(repositorySchema),
    defaultValues: {
      name: "",
      organizationId: "",
      visibility: "public",
    },
  });

  async function handleFormSubmit(values: RepositoryFormValues) {
    try {
      await registryApiClient.createRepository({
        name: values.name,
        organizationId: values.organizationId,
        visibility:
          visibilityMapper.get(values.visibility) ?? Visibility.PRIVATE,
      });

      refreshRepositories();

      toast.success("Repository created successfully.");
      reset();
      onOpenChange(false);
    } catch {
      toast.error("Failed to create repository.");
    }
  }

  function handleCancel() {
    reset();
    onCancel();
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <DialogHeader>
            <DialogTitle>Create repository</DialogTitle>
            <DialogDescription>
              Create a new repository in your workspace.
            </DialogDescription>
          </DialogHeader>
          <FieldGroup className="mt-4">
            <Controller
              control={control}
              name="organizationId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Organization</FieldLabel>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    disabled={isSubmitting || isLoadingOrganizations}
                    aria-invalid={fieldState.invalid}
                  >
                    <SelectTrigger
                      id={field.name}
                      className="w-full"
                      aria-invalid={fieldState.invalid}
                    >
                      <SelectValue placeholder="Select an organization" />
                    </SelectTrigger>
                    <SelectContent>
                      {organizations.map((org) => (
                        <SelectItem key={org.id} value={org.id}>
                          {org.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              control={control}
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    placeholder="awesome-project"
                    disabled={isSubmitting}
                    autoFocus
                    aria-invalid={fieldState.invalid}
                  />
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              control={control}
              name="visibility"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Visibility</FieldLabel>
                  <RadioGroup
                    value={field.value}
                    onValueChange={field.onChange}
                    disabled={isSubmitting}
                    aria-invalid={fieldState.invalid}
                    className="mt-1"
                  >
                    <label className="flex items-center gap-2 text-sm">
                      <RadioGroupItem
                        value="public"
                        id={`${field.name}-public`}
                      />
                      <span>Public</span>
                    </label>
                    <label className="flex items-center gap-2 text-sm">
                      <RadioGroupItem
                        value="private"
                        id={`${field.name}-private`}
                      />
                      <span>Private</span>
                    </label>
                  </RadioGroup>
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          </FieldGroup>
          <DialogFooter className="mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              Create
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
