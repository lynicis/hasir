"use client";

import { getMembers } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { OrganizationService } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";

import {
  type OrganizationMember,
  type Permission,
} from "@/components/member-item";
import { DeleteMemberDialog } from "@/components/delete-member-dialog";
import { InviteUserDialog } from "@/components/invite-user-dialog";
import { MembersList } from "@/components/members-list";
import { useSession } from "@/lib/session-provider";
import { customRetry } from "@/lib/query-retry";
import { isNotFoundError } from "@/lib/utils";
import { useClient } from "@/lib/use-client";

const memberRoleMapper = new Map<Role, Permission>([
  [Role.OWNER, "owner"],
  [Role.AUTHOR, "author"],
  [Role.READER, "reader"],
]);

const permissionRoleMapper = new Map<Permission, Role>([
  ["owner", Role.OWNER],
  ["author", Role.AUTHOR],
  ["reader", Role.READER],
]);

export default function OrganizationUsersContent() {
  const params = useParams();
  const organizationId = params.id as string;
  const { session } = useSession();
  const organizationApiClient = useClient(OrganizationService);

  const {
    data: membersData,
    error: membersError,
    refetch: refetchMembers,
  } = useQuery(
    getMembers,
    { id: organizationId },
    { retry: customRetry },
  );

  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [deleteMemberDialog, setDeleteMemberDialog] = useState<{
    open: boolean;
    member: OrganizationMember | null;
  }>({ open: false, member: null });
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

  const currentMemberPermission = useMemo<Permission | null>(() => {
    if (!membersData?.members || !session?.user?.email) {
      return null;
    }

    const currentMember = membersData.members.find(
      (member) => member.email === session.user?.email
    );

    if (!currentMember) {
      return null;
    }

    return memberRoleMapper.get(currentMember.role) as Permission;
  }, [membersData, session]);

  const canInviteMembers =
    currentMemberPermission === "owner" || currentMemberPermission === "author";
  const canRemoveMembers = currentMemberPermission === "owner";
  const canEditPermissions = currentMemberPermission === "owner";

  const fetchedMembers = useMemo<OrganizationMember[]>(() => {
    if (!membersData?.members) return [];

    return membersData.members.map((member) => ({
      id: member.id,
      email: member.email,
      name: member.username,
      permission: memberRoleMapper.get(member.role) as Permission,
    }));
  }, [membersData]);

  const ownerCount = useMemo(() => {
    return members.filter((member) => member.permission === "owner").length;
  }, [members]);

  useEffect(() => {
    setMembers(fetchedMembers);
  }, [fetchedMembers]);

  useEffect(() => {
    if (membersError && !isNotFoundError(membersError)) {
      toast.error("Error occurred while fetching members");
    }
  }, [membersError]);

  async function handlePermissionChange(
    memberId: string,
    newPermission: Permission
  ) {
    if (!canEditPermissions) {
      toast.error("You don't have permission to change member roles.");
      return;
    }

    const newRole = permissionRoleMapper.get(newPermission);
    if (!newRole) {
      toast.error("Invalid permission selected.");
      return;
    }

    const previousMembers = members;
    setMembers((prev) =>
      prev.map((member) =>
        member.id === memberId
          ? { ...member, permission: newPermission }
          : member
      )
    );

    try {
      await organizationApiClient.updateMemberRole({
        organizationId: organizationId,
        memberId: memberId,
        role: newRole,
      });

      toast.success("Permission updated successfully");
      await refetchMembers();
    } catch (error) {
      setMembers(previousMembers);

      if (error instanceof ConnectError) {
        switch (error.code) {
          case Code.PermissionDenied:
            toast.error("You don't have permission to change member roles.");
            break;
          case Code.NotFound:
            toast.error("Member not found.");
            break;
          case Code.InvalidArgument:
            toast.error("Invalid role selected.");
            break;
          default:
            toast.error("Failed to update permission. Please try again.");
        }
      } else {
        toast.error("Failed to update permission. Please try again.");
      }
    }
  }

  function handleDeleteMember(member: OrganizationMember) {
    if (!canRemoveMembers) {
      toast.error("You don't have permission to remove members.");
      return;
    }

    setDeleteMemberDialog({ open: true, member });
  }

  async function confirmDeleteMember() {
    if (!deleteMemberDialog.member) return;

    const memberToDelete = deleteMemberDialog.member;
    const previousMembers = members;

    setMembers((prev) => prev.filter((member) => member.id !== memberToDelete.id));
    setDeleteMemberDialog({ open: false, member: null });

    try {
      await organizationApiClient.deleteMember({
        organizationId: organizationId,
        memberId: memberToDelete.id,
      });

      toast.success(
        `${memberToDelete.name} has been removed from the organization`
      );
      await refetchMembers();
    } catch (error) {
      setMembers(previousMembers);

      if (error instanceof ConnectError) {
        switch (error.code) {
          case Code.PermissionDenied:
            toast.error("You don't have permission to remove members.");
            break;
          case Code.NotFound:
            toast.error("Member not found.");
            break;
          default:
            toast.error("Failed to remove member. Please try again.");
        }
      } else {
        toast.error("Failed to remove member. Please try again.");
      }
    }
  }

  function handleInviteMember() {
    if (!canInviteMembers) {
      toast.error("You don't have permission to invite members.");
      return;
    }

    setIsInviteDialogOpen(true);
  }

  function getInitials(name: string): string {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  }

  return (
    <div className="h-full space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Users</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage organization members and their permissions
        </p>
      </div>
      <MembersList
        members={members}
        onPermissionChange={handlePermissionChange}
        onDelete={handleDeleteMember}
        onInvite={handleInviteMember}
        canInvite={canInviteMembers}
        canEditPermissions={canEditPermissions}
        canRemoveMembers={canRemoveMembers}
        getInitials={getInitials}
        ownerCount={ownerCount}
      />
      <DeleteMemberDialog
        open={deleteMemberDialog.open}
        member={deleteMemberDialog.member}
        onOpenChange={(open) =>
          setDeleteMemberDialog({ open, member: deleteMemberDialog.member })
        }
        onConfirm={confirmDeleteMember}
      />
      <InviteUserDialog
        open={isInviteDialogOpen}
        onOpenChange={setIsInviteDialogOpen}
        organizationId={organizationId}
      />
    </div>
  );
}
