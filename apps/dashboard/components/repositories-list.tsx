"use client";

import { ChevronLeft, ChevronRight, Plus } from "lucide-react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";

import { RepositoryItem, type OrganizationRepository } from "./repository-item";

interface RepositoriesListProps {
  repositories: OrganizationRepository[];
  isLoading?: boolean;
  currentPage: number;
  totalPages: number;
  totalRepositories: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  onPageChange: (page: number) => void;
  onDelete: (repository: OrganizationRepository) => void;
  onCreate: () => void;
}

export function RepositoriesList({
  repositories,
  isLoading = false,
  currentPage,
  totalPages,
  totalRepositories,
  hasNextPage,
  hasPreviousPage,
  onPageChange,
  onDelete,
  onCreate,
}: RepositoriesListProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Repositories</CardTitle>
            <CardDescription>
              {isLoading
                ? "Loading repositories..."
                : totalRepositories > 0
                ? `${totalRepositories} ${
                    totalRepositories === 1 ? "repository" : "repositories"
                  }`
                : "Manage organization repositories"}
            </CardDescription>
          </div>
          <Button onClick={onCreate}>
            <Plus className="size-4 mr-2" />
            Create Repository
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center justify-between rounded-lg border border-border/60 bg-card px-4 py-3"
              >
                <Skeleton className="h-5 w-48" />
                <Skeleton className="h-8 w-16" />
              </div>
            ))}
          </div>
        ) : repositories.length === 0 ? (
          <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
            No repositories found
          </div>
        ) : (
          <>
            <div className="space-y-3">
              {repositories.map((repo) => (
                <RepositoryItem
                  key={repo.id}
                  repository={repo}
                  onDelete={onDelete}
                />
              ))}
            </div>
            {totalPages > 1 && (
              <div className="mt-6 flex items-center justify-between border-t border-border/60 pt-4">
                <div className="text-sm text-muted-foreground">
                  Page {currentPage} of {totalPages}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={!hasPreviousPage}
                  >
                    <ChevronLeft className="size-4" />
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={!hasNextPage}
                  >
                    Next
                    <ChevronRight className="size-4" />
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
