"use client";

import { UserPlus } from "lucide-react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

import {
  MemberItem,
  type OrganizationMember,
  type Permission,
} from "./member-item";

interface MembersListProps {
  members: OrganizationMember[];
  onPermissionChange: (memberId: string, newPermission: Permission) => void;
  onDelete: (member: OrganizationMember) => void;
  onInvite: () => void;
  canInvite: boolean;
  canEditPermissions: boolean;
  canRemoveMembers: boolean;
  getInitials: (name: string) => string;
  ownerCount: number;
}

export function MembersList({
  members,
  onPermissionChange,
  onDelete,
  onInvite,
  canInvite,
  canEditPermissions,
  canRemoveMembers,
  getInitials,
  ownerCount,
}: MembersListProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Members</CardTitle>
            <CardDescription>
              Manage organization members and their permissions
            </CardDescription>
          </div>
          <Button onClick={onInvite} disabled={!canInvite}>
            <UserPlus className="size-4 mr-2" />
            Invite Member
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {members.map((member) => (
            <MemberItem
              key={member.id}
              member={member}
              onPermissionChange={onPermissionChange}
              onDelete={onDelete}
              canEditPermissions={canEditPermissions}
              canRemove={canRemoveMembers}
              getInitials={getInitials}
              ownerCount={ownerCount}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
