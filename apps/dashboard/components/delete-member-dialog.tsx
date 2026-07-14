"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

import { type OrganizationMember } from "./member-item";

interface DeleteMemberDialogProps {
  open: boolean;
  member: OrganizationMember | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function DeleteMemberDialog({
  open,
  member,
  onOpenChange,
  onConfirm,
}: DeleteMemberDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove member from organization?</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to remove <strong>{member?.name}</strong> (
            {member?.email}) from this organization? This action cannot be
            undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Remove Member
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
