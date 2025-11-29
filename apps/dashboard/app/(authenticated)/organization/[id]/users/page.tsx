"use client";

import { useState } from "react";
import { toast } from "sonner";

import {
  type OrganizationMember,
  type Permission,
} from "@/components/member-item";
import { DeleteMemberDialog } from "@/components/delete-member-dialog";
import { MembersList } from "@/components/members-list";

const mockMembers: OrganizationMember[] = [
  {
    id: "1",
    email: "john.doe@example.com",
    name: "John Doe",
    permission: "owner",
  },
  {
    id: "2",
    email: "jane.smith@example.com",
    name: "Jane Smith",
    permission: "admin",
  },
  {
    id: "3",
    email: "bob.johnson@example.com",
    name: "Bob Johnson",
    permission: "member",
  },
  {
    id: "4",
    email: "alice.brown@example.com",
    name: "Alice Brown",
    permission: "viewer",
  },
];

export default function UsersPage() {
  const [members, setMembers] = useState<OrganizationMember[]>(mockMembers);
  const [deleteMemberDialog, setDeleteMemberDialog] = useState<{
    open: boolean;
    member: OrganizationMember | null;
  }>({ open: false, member: null });

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
    toast.info("Invite member functionality coming soon");
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
    </div>
  );
}
