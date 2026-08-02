"use client";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuGroup,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@hasir/ui/components/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@hasir/ui/components/avatar";
import { Button } from "@hasir/ui/components/button";
import { MoreVertical, Trash2 } from "lucide-react";

export type Permission = "owner" | "author" | "reader";

export interface OrganizationMember {
  id: string;
  email: string;
  name: string;
  permission: Permission;
  avatar?: string;
}

export const permissionLabels: Record<Permission, string> = {
  owner: "Owner",
  author: "Author",
  reader: "Reader",
};

interface MemberItemProps {
  member: OrganizationMember;
  onPermissionChange: (memberId: string, newPermission: Permission) => void;
  onDelete: (member: OrganizationMember) => void;
  canEditPermissions: boolean;
  canRemove: boolean;
  getInitials: (name: string) => string;
  ownerCount: number;
}

export function MemberItem({
  member,
  onPermissionChange,
  onDelete,
  canEditPermissions,
  canRemove,
  getInitials,
  ownerCount,
}: MemberItemProps) {
  const isOwner = member.permission === "owner";
  const isLastOwner = isOwner && ownerCount === 1;
  const canEditRole = canEditPermissions && !isLastOwner;

  return (
    <div className="flex items-center justify-between rounded-md border bg-card px-4 py-3 hover:bg-accent transition-colors">
      <div className="flex items-center gap-3">
        <Avatar>
          <AvatarImage src={member.avatar} alt={member.name} />
          <AvatarFallback>{getInitials(member.name)}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col">
          <span className="font-medium text-sm">{member.name}</span>
          <span className="text-xs text-muted-foreground">{member.email}</span>
        </div>
      </div>
      <div className="flex items-center gap-3">
        {!canEditRole ? (
          <Button variant="outline" size="sm" disabled>
            {permissionLabels[member.permission]}
          </Button>
        ) : (
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" size="sm" />}>
              {permissionLabels[member.permission]}
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuGroup><DropdownMenuLabel>Change Permission</DropdownMenuLabel></DropdownMenuGroup>
              <DropdownMenuSeparator />
              {Object.entries(permissionLabels).map(([value, label]) => (
                <DropdownMenuItem
                  key={value}
                  onClick={() =>
                    onPermissionChange(member.id, value as Permission)
                  }
                  className={member.permission === value ? "bg-accent" : ""}
                >
                  {label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
        <DropdownMenu>
          <DropdownMenuTrigger render={<Button variant="ghost" size="icon" />}>
            <MoreVertical className="size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              variant="destructive"
              onClick={() => onDelete(member)}
              disabled={member.permission === "owner" || !canRemove}
            >
              <Trash2 className="size-4 mr-2" />
              Remove from organization
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}
