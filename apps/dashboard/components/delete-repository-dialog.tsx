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

import { type OrganizationRepository } from "./repository-item";

interface DeleteRepositoryDialogProps {
  open: boolean;
  repository: OrganizationRepository | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function DeleteRepositoryDialog({
  open,
  repository,
  onOpenChange,
  onConfirm,
}: DeleteRepositoryDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete repository?</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the repository{" "}
            <strong>{repository?.name}</strong>? This action cannot be undone
            and all data will be permanently deleted.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Delete Repository
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
