"use client";

import { Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";

export interface OrganizationRepository {
  id: string;
  name: string;
  visibility: "public" | "private";
}

interface RepositoryItemProps {
  repository: OrganizationRepository;
  onDelete: (repository: OrganizationRepository) => void;
}

export function RepositoryItem({ repository, onDelete }: RepositoryItemProps) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border/60 bg-card px-4 py-3 hover:bg-accent/50 transition-colors">
      <div className="flex items-center gap-3 flex-1">
        <div className="flex flex-col flex-1">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm">{repository.name}</span>
            <span
              className={`text-xs px-2 py-0.5 rounded-full ${
                repository.visibility === "public"
                  ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                  : "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300"
              }`}
            >
              {repository.visibility}
            </span>
          </div>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => onDelete(repository)}
        >
          <Trash2 className="size-4" />
        </Button>
      </div>
    </div>
  );
}
