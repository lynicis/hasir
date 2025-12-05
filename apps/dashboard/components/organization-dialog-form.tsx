"use client";

import { OrganizationService } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { Controller, useFieldArray, useForm } from "react-hook-form";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { ArrowLeftIcon, PlusIcon, XIcon } from "lucide-react";
import { Code, ConnectError } from "@connectrpc/connect";
import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
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
import { useRegistryStore } from "@/stores/registry-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";
import { cn } from "@/lib/utils";

const organizationSchema = z.object({
  name: z
    .string()
    .min(1, { error: "Please enter an organization name." })
    .max(100, { error: "Organization name must be at most 100 characters." }),
  visibility: z.enum(["public", "private"], {
    error: "Please select a visibility.",
  }),
  invites: z.array(
    z.object({
      email: z.email({ error: "Please enter a valid email address." }),
      role: z.enum(["reader", "author", "owner"] as const, {
        error: "Please select a role.",
      }),
    })
  ),
});

export type OrganizationFormValues = z.infer<typeof organizationSchema>;

interface OrganizationDialogFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCancel: () => void;
}

export function OrganizationDialogForm({
  open,
  onOpenChange,
  onCancel,
}: OrganizationDialogFormProps) {
  const organizationApiClient = useClient(OrganizationService);
  const invalidateOrganizations = useRegistryStore(
    (state) => state.invalidateOrganizations
  );
  const [step, setStep] = useState<1 | 2>(1);

  const {
    control,
    handleSubmit,
    formState: { isSubmitting },
    reset,
    getValues,
    setError,
  } = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationSchema),
    defaultValues: {
      name: "",
      visibility: "public",
      invites: [],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: "invites",
  });

  async function handleFormSubmit(values: OrganizationFormValues) {
    try {
      const members = values.invites.map(({ email, role }) => {
        const mappedRole =
          role === "author"
            ? Role.AUTHOR
            : role === "owner"
            ? Role.OWNER
            : Role.READER;

        return {
          email,
          role: mappedRole,
        };
      });

      await organizationApiClient.createOrganization({
        name: values.name,
        visibility:
          visibilityMapper.get(values.visibility) || Visibility.PUBLIC,
        members,
      });

      invalidateOrganizations();

      toast.success("Organization created successfully.");
      handleClose();
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          toast.error(
            "Some invitations could not be sent. Invitees may not be registered, may have already been invited, or are already members of this organization."
          );
          return;
        }
      }

      toast.error("Failed to create organization.");
    }
  }

  function handleClose() {
    reset();
    setStep(1);
    onOpenChange(false);
  }

  function handleCancel() {
    reset();
    setStep(1);
    onCancel();
  }

  function handleNext(e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    const { name } = getValues();
    if (!name.trim()) {
      setError("name", { message: "Please enter an organization name." });
      return;
    }
    setStep(2);
  }

  function handleBack() {
    setStep(1);
  }

  function addInvite() {
    append({ email: "", role: "reader" });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <DialogHeader>
            <div className="flex items-center gap-2">
              {step === 2 && (
                <button
                  type="button"
                  onClick={handleBack}
                  className="-ml-1 rounded-md p-1.5 transition-colors hover:bg-accent"
                  aria-label="Go back"
                >
                  <ArrowLeftIcon className="size-4" />
                </button>
              )}
              <DialogTitle>
                {step === 1 ? "Create organization" : "Invite your friends"}
              </DialogTitle>
            </div>
            <DialogDescription>
              {step === 1
                ? "Set up your new organization."
                : "Invite team members to collaborate with you."}
            </DialogDescription>
          </DialogHeader>

          <div className="mt-4 flex items-center gap-2">
            <div
              className={cn(
                "h-1.5 flex-1 rounded-full transition-colors",
                step >= 1 ? "bg-primary" : "bg-muted"
              )}
            />
            <div
              className={cn(
                "h-1.5 flex-1 rounded-full transition-colors",
                step >= 2 ? "bg-primary" : "bg-muted"
              )}
            />
          </div>

          {step === 1 && (
            <FieldGroup className="mt-4">
              <Controller
                control={control}
                name="name"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                    <Input
                      {...field}
                      id={field.name}
                      placeholder="My Organization"
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
                        <div className="flex flex-col">
                          <span>Public</span>
                          <span className="text-xs text-muted-foreground">
                            Anyone can see this organization
                          </span>
                        </div>
                      </label>
                      <label className="flex items-center gap-2 text-sm">
                        <RadioGroupItem
                          value="private"
                          id={`${field.name}-private`}
                        />
                        <div className="flex flex-col">
                          <span>Private</span>
                          <span className="text-xs text-muted-foreground">
                            Only members can see this organization
                          </span>
                        </div>
                      </label>
                    </RadioGroup>
                    {fieldState.error && (
                      <FieldError errors={[fieldState.error]} />
                    )}
                  </Field>
                )}
              />
            </FieldGroup>
          )}

          {step === 2 && (
            <div className="mt-4 space-y-4">
              <div className="space-y-3">
                {fields.length === 0 ? (
                  <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border py-8 text-center">
                    <p className="text-sm text-muted-foreground">
                      No invites yet
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Add email addresses to invite team members
                    </p>
                  </div>
                ) : (
                  fields.map((field, index) => (
                    <div key={field.id} className="space-y-2">
                      <div className="flex items-start gap-2">
                        <Controller
                          control={control}
                          name={`invites.${index}.email`}
                          render={({ field: inputField, fieldState }) => (
                            <Field
                              data-invalid={fieldState.invalid}
                              className="flex-1"
                            >
                              <Input
                                {...inputField}
                                type="email"
                                placeholder="friend@example.com"
                                disabled={isSubmitting}
                                autoFocus={index === fields.length - 1}
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
                          name={`invites.${index}.role`}
                          render={({ field: selectField, fieldState }) => (
                            <Field
                              data-invalid={fieldState.invalid}
                              className="w-32"
                            >
                              <Select
                                value={selectField.value}
                                onValueChange={selectField.onChange}
                                disabled={isSubmitting}
                              >
                                <SelectTrigger aria-label="Role">
                                  <SelectValue placeholder="Role" />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="reader">Reader</SelectItem>
                                  <SelectItem value="author">Author</SelectItem>
                                  <SelectItem value="owner">Owner</SelectItem>
                                </SelectContent>
                              </Select>
                              {fieldState.error && (
                                <FieldError errors={[fieldState.error]} />
                              )}
                            </Field>
                          )}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => remove(index)}
                          disabled={isSubmitting}
                          className="shrink-0"
                        >
                          <XIcon className="size-4" />
                        </Button>
                      </div>
                    </div>
                  ))
                )}
              </div>

              <Button
                type="button"
                variant="outline"
                onClick={addInvite}
                disabled={isSubmitting}
                className="w-full"
              >
                <PlusIcon className="size-4" />
                Add email
              </Button>
            </div>
          )}

          <DialogFooter className="mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            {step === 1 ? (
              <Button
                type="button"
                onClick={handleNext}
                disabled={isSubmitting}
              >
                Next
              </Button>
            ) : (
              <Button type="submit" disabled={isSubmitting}>
                {fields.length > 0 ? "Create & Invite" : "Create"}
              </Button>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
