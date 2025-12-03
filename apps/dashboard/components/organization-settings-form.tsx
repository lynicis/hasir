"use client";

import { getOrganization } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { getMembers } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { OrganizationService } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { useQuery as useMembersQuery } from "@connectrpc/connect-query";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { Code, ConnectError } from "@connectrpc/connect";
import { useParams, useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { z } from "zod/v4";

import type { Permission } from "@/components/member-item";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import {
  reverseVisibilityMapper,
  visibilityMapper,
} from "@/lib/visibility-mapper";
import { DeleteOrganizationDialog } from "@/components/delete-organization-dialog";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { InviteUserDialog } from "@/components/invite-user-dialog";
import { useRefreshStore } from "@/stores/refresh-store";
import { Separator } from "@/components/ui/separator";
import { useSession } from "@/lib/session-provider";
import { Button } from "@/components/ui/button";
import { customRetry } from "@/lib/query-retry";
import { Input } from "@/components/ui/input";
import { isNotFoundError } from "@/lib/utils";
import { useClient } from "@/lib/use-client";

const memberRoleMapper = new Map<Role, Permission>([
  [Role.OWNER, "owner"],
  [Role.AUTHOR, "author"],
  [Role.READER, "reader"],
]);

const organizationSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, { error: "Please enter an organization name." })
    .max(100, { error: "Organization name must be at most 100 characters." }),
  visibility: z.enum(["public", "private"], {
    error: "Please select a visibility.",
  }),
});

type OrganizationFormValues = z.infer<typeof organizationSchema>;

export function OrganizationSettingsForm() {
  const params = useParams();
  const router = useRouter();
  const organizationId = params.id as string;
  const organizationApiClient = useClient(OrganizationService);
  const { session } = useSession();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const {
    data,
    isLoading: isLoadingOrganization,
    error: organizationError,
    refetch,
  } = useQuery(
    getOrganization,
    {
      id: organizationId,
    },
    { retry: customRetry }
  );

  const organizationsRefreshKey = useRefreshStore(
    (state) => state.organizationsRefreshKey
  );

  const organization = data?.organization;

  const { data: membersData } = useMembersQuery(
    getMembers,
    { id: organizationId },
    { retry: customRetry }
  );

  const members =
    membersData?.members ??
    (data as { members?: { email: string; role: Role }[] } | undefined)
      ?.members;

  const currentMemberPermission = useMemo<Permission | null>(() => {
    if (!members || members.length === 0 || !session?.user?.email) {
      return null;
    }

    const currentMember = members.find(
      (member) => member.email === session.user?.email
    );

    if (!currentMember) {
      return null;
    }

    return memberRoleMapper.get(currentMember.role) as Permission;
  }, [members, session]);

  const isOwner = currentMemberPermission === "owner";

  const {
    control,
    handleSubmit,
    formState: { isSubmitting },
    reset,
  } = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationSchema),
    defaultValues: {
      name: organization?.name || "",
      visibility: "public",
    },
  });

  useEffect(() => {
    if (organization) {
      reset({
        name: organization.name,
        visibility:
          reverseVisibilityMapper.get(organization.visibility) || "private",
      });
    }
  }, [organization, reset]);

  useEffect(() => {
    if (organizationError && !isNotFoundError(organizationError)) {
      toast.error("Error occurred while fetching organization");
    }
  }, [organizationError]);

  useEffect(() => {
    if (organizationsRefreshKey > 0) {
      refetch();
    }
  }, [organizationsRefreshKey, refetch]);

  async function handleFormSubmit(values: OrganizationFormValues) {
    try {
      await organizationApiClient.updateOrganization({
        id: organizationId,
        name: values.name,
        visibility:
          visibilityMapper.get(values.visibility) || Visibility.PUBLIC,
      });

      toast.success("Organization updated successfully.");
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to update this organization.");
          return;
        }

        if (error.code === Code.NotFound) {
          toast.error("Organization not found.");
          return;
        }

        if (error.code === Code.InvalidArgument) {
          const errorMessage =
            error.message ||
            "Invalid organization data. Please check your input.";
          toast.error(errorMessage);
          return;
        }
      }

      toast.error("Failed to update organization. Please try again.");
    }
  }

  async function handleDeleteOrganization() {
    if (!organization) return;

    setIsDeleting(true);
    try {
      await organizationApiClient.deleteOrganization({
        id: organizationId,
      });

      toast.success("Organization deleted successfully.");
      router.push("/dashboard");
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete this organization.");
          setIsDeleteDialogOpen(false);
          return;
        }

        if (error.code === Code.NotFound) {
          toast.error("Organization not found.");
          setIsDeleteDialogOpen(false);
          return;
        }
      }

      toast.error("Failed to delete organization. Please try again.");
      setIsDeleteDialogOpen(false);
    } finally {
      setIsDeleting(false);
    }
  }

  if (isLoadingOrganization) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>General Settings</CardTitle>
          <CardDescription>
            Update your organization information
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="h-4 w-32 bg-muted animate-pulse rounded" />
              <div className="h-10 w-full bg-muted animate-pulse rounded" />
            </div>
            <div className="space-y-2">
              <div className="h-4 w-24 bg-muted animate-pulse rounded" />
              <div className="h-20 w-full bg-muted animate-pulse rounded" />
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!isOwner && currentMemberPermission !== null) {
    router.replace(`/organization/${organizationId}/users`);
    return null;
  }

  if (!organization) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>General Settings</CardTitle>
          <CardDescription>
            Update your organization information
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Organization not found.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>General Settings</CardTitle>
          <CardDescription>
            Update your organization information
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(handleFormSubmit)}>
            <FieldGroup className="space-y-6">
              <Controller
                control={control}
                name="name"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>
                      Organization Name
                    </FieldLabel>
                    <Input
                      {...field}
                      id={field.name}
                      type="text"
                      placeholder="My Organization"
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
                name="visibility"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Visibility</FieldLabel>
                    <RadioGroup
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={isSubmitting}
                      aria-invalid={fieldState.invalid}
                      className="mt-2"
                    >
                      <label className="flex items-center gap-2 cursor-pointer">
                        <RadioGroupItem
                          value="public"
                          id={`${field.name}-public`}
                        />
                        <span className="text-sm">Public</span>
                      </label>
                      <label className="flex items-center gap-2 cursor-pointer">
                        <RadioGroupItem
                          value="private"
                          id={`${field.name}-private`}
                        />
                        <span className="text-sm">Private</span>
                      </label>
                    </RadioGroup>
                    {fieldState.error && (
                      <FieldError errors={[fieldState.error]} />
                    )}
                  </Field>
                )}
              />
              <Separator />
              <Field>
                <Button
                  type="submit"
                  isLoading={isSubmitting}
                  disabled={isSubmitting}
                >
                  Save Changes
                </Button>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Members</CardTitle>
          <CardDescription>
            Invite users to join your organization
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4">
            <div>
              <h3 className="text-sm font-medium mb-1">Invite User</h3>
              <p className="text-sm text-muted-foreground mb-4">
                Send an invitation email to add new members to your
                organization.
              </p>
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsInviteDialogOpen(true)}
                disabled={!isOwner}
              >
                Invite User
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="mt-6 border-destructive/50">
        <CardHeader>
          <CardTitle className="text-xl text-destructive">
            Danger Zone
          </CardTitle>
          <CardDescription>
            Irreversible and destructive actions
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4">
            <div>
              <h3 className="text-sm font-medium mb-1">Delete Organization</h3>
              <p className="text-sm text-muted-foreground mb-4">
                Once you delete an organization, there is no going back. Please
                be certain.
              </p>
              <Button
                type="button"
                variant="destructive"
                onClick={() => setIsDeleteDialogOpen(true)}
                disabled={isDeleting}
              >
                Delete Organization
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <InviteUserDialog
        open={isInviteDialogOpen}
        onOpenChange={setIsInviteDialogOpen}
        organizationId={organizationId}
      />

      <DeleteOrganizationDialog
        open={isDeleteDialogOpen}
        organizationName={organization?.name ?? null}
        onOpenChange={(open) => {
          if (!isDeleting) {
            setIsDeleteDialogOpen(open);
          }
        }}
        onConfirm={handleDeleteOrganization}
        isDeleting={isDeleting}
      />
    </>
  );
}
