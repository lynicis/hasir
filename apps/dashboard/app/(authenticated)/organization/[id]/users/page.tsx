"use client";

import { getMembers } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
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
import { customRetry } from "@/lib/query-retry";
import { isNotFoundError } from "@/lib/utils";

const memberRoleMapper = new Map<Role, Permission>([
  [Role.OWNER, "owner"],
  [Role.AUTHOR, "admin"],
  [Role.READER, "member"],
]);

export default function UsersPage() {
  const params = useParams();
  const organizationId = params.id as string;

  const { data: membersData, error: membersError } = useQuery(
    getMembers,
    { id: organizationId },
    { retry: customRetry }
  );

  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [deleteMemberDialog, setDeleteMemberDialog] = useState<{
    open: boolean;
    member: OrganizationMember | null;
  }>({ open: false, member: null });
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

  const fetchedMembers = useMemo<OrganizationMember[]>(() => {
    if (!membersData?.members) return [];

    return membersData.members.map((member) => ({
      id: member.id,
      email: member.email,
      name: member.email.split("@")[0] || member.email,
      permission: memberRoleMapper.get(member.role) as Permission,
    }));
  }, [membersData]);

  useEffect(() => {
    setMembers(fetchedMembers);
  }, [fetchedMembers]);

  useEffect(() => {
    if (membersError && !isNotFoundError(membersError)) {
      toast.error("Error occurred while fetching members");
    }
  }, [membersError]);

  function handlePermissionChange(memberId: string, newPermission: Permission) {
    setMembers((prev) =>
      prev.map((member) =>
        member.id === memberId
          ? { ...member, permission: newPermission }
          : member
      )
    );
    toast.success("Permission updated successfully");
  }

  function handleDeleteMember(member: OrganizationMember) {
    setDeleteMemberDialog({ open: true, member });
  }

  function confirmDeleteMember() {
    if (!deleteMemberDialog.member) return;

    setMembers((prev) =>
      prev.filter((m) => m.id !== deleteMemberDialog.member!.id)
    );
    toast.success(
      `${deleteMemberDialog.member.name} has been removed from the organization`
    );
    setDeleteMemberDialog({ open: false, member: null });
  }

  function handleInviteMember() {
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
        getInitials={getInitials}
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
