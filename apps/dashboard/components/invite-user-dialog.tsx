"use client";

import { OrganizationService } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { Code, ConnectError } from "@connectrpc/connect";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { UserPlus } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod/v4";

import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";

const inviteUserSchema = z.object({
  email: z.email({ message: "Please enter a valid email address." }),
  role: z.enum(Role),
});

type InviteUserFormValues = z.infer<typeof inviteUserSchema>;

interface InviteUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
}

export function InviteUserDialog({
  open,
  onOpenChange,
  organizationId,
}: InviteUserDialogProps) {
  const organizationApiClient = useClient(OrganizationService);

  const {
    control,
    handleSubmit,
    formState: { isSubmitting },
    reset,
  } = useForm<InviteUserFormValues>({
    resolver: zodResolver(inviteUserSchema),
    defaultValues: {
      email: "",
      role: Role.READER,
    },
  });

  async function handleFormSubmit(values: InviteUserFormValues) {
    try {
      await organizationApiClient.inviteMember({
        id: organizationId,
        email: values.email,
        role: values.role,
      });

      toast.success(`Invitation sent to ${values.email}`);
      reset();
      onOpenChange(false);
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to invite members.");
          return;
        }

        if (error.code === Code.NotFound) {
          toast.error(
            "This user cannot be invited. They may not be registered, may have already received an invitation, or are already a member of this organization."
          );
          return;
        }

        if (error.code === Code.AlreadyExists) {
          toast.error("This user is already a member of the organization.");
          return;
        }

        if (error.code === Code.InvalidArgument) {
          const errorMessage =
            error.message || "Invalid email address. Please check your input.";
          toast.error(errorMessage);
          return;
        }
      }

      toast.error("Failed to send invitation. Please try again.");
    }
  }

  function handleOpenChange(open: boolean) {
    if (!isSubmitting) {
      if (!open) {
        reset();
      }
      onOpenChange(open);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite User</DialogTitle>
          <DialogDescription>
            Send an invitation to join this organization via email
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <FieldGroup className="space-y-4">
            <Controller
              control={control}
              name="email"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Email Address</FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    type="email"
                    placeholder="user@example.com"
                    disabled={isSubmitting}
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
              name="role"
              render={({ field, fieldState }) => {
                const selectValue =
                  typeof field.value === "number"
                    ? Role[field.value]?.toLowerCase() ?? ""
                    : field.value ?? "";

                return (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Role</FieldLabel>
                    <Select
                      value={selectValue}
                      onValueChange={(value) => {
                        const roleKey =
                          value.toUpperCase() as keyof typeof Role;
                        const roleValue = Role[roleKey];
                        field.onChange(roleValue);
                      }}
                      disabled={isSubmitting}
                    >
                      <SelectTrigger id={field.name} aria-label="Role">
                        <SelectValue placeholder="Select a role" />
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
                );
              }}
            />
            <div className="flex justify-end gap-3 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button type="submit" isLoading={isSubmitting}>
                <UserPlus className="size-4" />
                Send Invitation
              </Button>
            </div>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
