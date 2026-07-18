"use client";

import { getOrganizations } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { getRepositories } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo, useState } from "react";
import { Settings } from "lucide-react";
import { toast } from "sonner";
import Link from "next/link";

import { OrganizationDialogForm } from "@/components/organization-dialog-form";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { reverseVisibilityMapper } from "@/lib/visibility-mapper";
import { useRegistryStore } from "@/stores/registry-store";
import { Pagination } from "@/components/ui/pagination";
import { Skeleton } from "@/components/ui/skeleton";
import { customRetry } from "@/lib/query-retry";
import { isNotFoundError } from "@/lib/utils";

const DEFAULT_PAGINATION = {
  page: 1,
  pageLimit: 5,
};

export function Dashboard() {
  const [activeOrgId, setActiveOrgId] = useState<string | "all">("all");
  const [isCreateOrgDialogOpen, setIsCreateOrgDialogOpen] = useState(false);
  const [orgPage, setOrgPage] = useState(DEFAULT_PAGINATION.page);
  const [repoPage, setRepoPage] = useState(1);

  const organizationsVersion = useRegistryStore(
    (state) => state.organizationsVersion
  );
  const repositoriesVersion = useRegistryStore(
    (state) => state.repositoriesVersion
  );

  const organizationsQueryParams = useMemo(
    () => ({
      pagination: {
        page: orgPage,
        pageLimit: DEFAULT_PAGINATION.pageLimit,
      },
    }),
    [orgPage]
  );

  const repositoriesQueryParams = useMemo(
    () =>
      activeOrgId === "all"
        ? {
            pagination: {
              page: repoPage,
              pageLimit: DEFAULT_PAGINATION.pageLimit,
            },
          }
        : {
            pagination: {
              page: repoPage,
              pageLimit: DEFAULT_PAGINATION.pageLimit,
            },
            organizationId: activeOrgId,
          },
    [activeOrgId, repoPage]
  );

  const {
    data: organizations,
    isLoading: isLoadingOrganizations,
    error: organizationsError,
    refetch: refetchOrganizations,
  } = useQuery(getOrganizations, organizationsQueryParams, {
    retry: customRetry,
  });

  const {
    data: repositoriesData,
    isLoading: isLoadingRepositories,
    error: repositoriesError,
    refetch: refetchRepositories,
  } = useQuery(getRepositories, repositoriesQueryParams, {
    retry: customRetry,
  });

  const organizationsList = organizations?.organizations ?? [];
  const repositoriesList = repositoriesData?.repositories ?? [];
  const orgTotalPages = organizations?.totalPage ?? 1;
  const repoTotalPages = repositoriesData?.totalPage ?? 1;

  const handleOrgChange = (orgId: string | "all") => {
    setActiveOrgId(orgId);
    setRepoPage(1);
  };

  useEffect(() => {
    if (organizationsError && !isNotFoundError(organizationsError)) {
      toast.error("Error occurred while fetching organizations");
    }
  }, [organizationsError]);

  useEffect(() => {
    if (repositoriesError && !isNotFoundError(repositoriesError)) {
      toast.error("Error occurred while fetching repositories");
    }
  }, [repositoriesError]);

  useEffect(() => {
    if (organizationsVersion > 0) {
      refetchOrganizations();
    }
  }, [organizationsVersion, refetchOrganizations]);

  useEffect(() => {
    if (repositoriesVersion > 0) {
      refetchRepositories();
    }
  }, [repositoriesVersion, refetchRepositories]);

  return (
    <div className="min-h-[calc(100vh-4.5rem)] bg-background px-6 py-6">
      <div className="mx-auto flex h-full w-full max-w-6xl flex-col gap-6">
        <main className="grid flex-1 grid-cols-[260px_minmax(0,1fr)] gap-6 pt-2 min-h-0">
          <Card className="h-full overflow-hidden rounded shadow-none border border-border bg-card flex flex-col gap-0 py-0">
            <CardHeader className="flex items-center bg-transparent px-6 py-4 shrink-0 border-b border-border">
              <CardTitle className="text-sm font-semibold font-mono tracking-[-0.02em] text-foreground">
                Your organizations
              </CardTitle>
            </CardHeader>
            <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4">
              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => handleOrgChange("all")}
                  className={`flex w-full items-center justify-between rounded-none px-3 py-2.5 text-sm transition-all duration-200 active:scale-[0.99] active:translate-y-[0.5px] ${
                    activeOrgId === "all"
                      ? "bg-accent text-accent-foreground font-medium"
                      : "hover:bg-accent/80 hover:text-accent-foreground"
                  }`}
                >
                  <span>All organizations</span>
                </button>
                {isLoadingOrganizations ? (
                  Array.from({ length: 3 }).map((_, i) => (
                    <div
                      key={i}
                      className="flex w-full items-center rounded-none px-3 py-2.5"
                    >
                      <Skeleton className="h-4 w-24" />
                    </div>
                  ))
                ) : organizationsError &&
                  !isNotFoundError(organizationsError) ? (
                  <div className="flex flex-col items-center justify-center py-6 space-y-3">
                    <div className="text-xs text-muted-foreground">
                      Failed to load organizations
                    </div>
                  </div>
                ) : organizationsList.length === 0 &&
                  organizationsError &&
                  isNotFoundError(organizationsError) ? (
                  <div className="flex items-center justify-center py-6 text-xs text-muted-foreground">
                    No organizations found
                  </div>
                ) : (
                  organizationsList.map((org) => {
                    const isActive = activeOrgId === org.id;
                    return (
                      <div
                        key={org.id}
                        className={`group flex items-center justify-between rounded-none text-sm transition-all duration-200 ${
                          isActive
                            ? "bg-accent text-accent-foreground font-medium"
                            : "hover:bg-accent/80 hover:text-accent-foreground"
                        }`}
                      >
                        <button
                          type="button"
                          onClick={() => handleOrgChange(org.id)}
                          className="flex-1 text-left font-mono truncate px-3 py-2.5 active:opacity-70 transition-opacity"
                        >
                          {org.name}
                        </button>
                        <Link
                          href={`/organization/${org.id}`}
                          className={`mr-2 p-1.5 text-muted-foreground hover:text-foreground transition-all duration-200 rounded-md hover:bg-background/50 ${
                            isActive ? "opacity-100" : "opacity-0 group-hover:opacity-100 focus:opacity-100"
                          }`}
                          title="Organization Settings"
                        >
                          <Settings className="size-3.5 transition-transform duration-500 hover:rotate-45" />
                          <span className="sr-only">Settings</span>
                        </Link>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
            <Pagination
              currentPage={orgPage}
              totalPages={orgTotalPages}
              onPageChange={setOrgPage}
              disabled={isLoadingOrganizations}
              className="px-6 pb-4 shrink-0"
            />
          </Card>
          <Card className="h-full overflow-hidden rounded shadow-none border border-border bg-card flex flex-col gap-0 py-0">
            <CardHeader className="flex items-center justify-between bg-transparent px-6 py-4 shrink-0 border-b border-border">
              <div className="space-y-0.5">
                <CardTitle className="text-sm font-semibold font-mono tracking-[-0.02em] text-foreground">
                  Repositories
                </CardTitle>
                {activeOrgId !== "all" && (
                  <p className="text-xs text-muted-foreground">
                    Showing repositories in{" "}
                    {
                      organizationsList.find(
                          (organization) => organization.id === activeOrgId
                      )?.name
                    }
                  </p>
                )}
              </div>
              {isLoadingRepositories ? (
                <Skeleton className="h-4 w-12 bg-secondary/70" />
              ) : (
                <span className="text-xs text-muted-foreground font-mono tabular-nums">
                  {repositoriesList.length} repos
                </span>
              )}
            </CardHeader>
            <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4">
              <div className="space-y-2.5">
                {isLoadingRepositories ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <div
                      key={i}
                      className="flex items-center justify-between rounded-none border border-border/60 bg-card px-4 py-3"
                    >
                      <Skeleton className="h-5 w-32" />
                    </div>
                  ))
                ) : repositoriesList.length === 0 ? (
                  <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
                    No repositories found
                  </div>
                ) : (
                  repositoriesList.map((repo) => {
                    const visibility = reverseVisibilityMapper.get(
                      repo.visibility,
                    );
                    return (
                      <Link
                        href={`/repository/${repo.id}`}
                        key={repo.id}
                        className="hover:bg-accent/60 hover:border-primary/30 flex items-center justify-between rounded-none border border-border/65 bg-card px-4 py-3 text-sm transition-all duration-200 active:scale-[0.99] active:translate-y-[0.5px]"
                      >
                        <div className="flex items-center gap-2">
                          <span className="font-medium font-mono">{repo.name}</span>
                          {visibility && (
                            <span
                              className={`text-xs px-2 py-0.5 rounded-none font-mono uppercase tracking-wider transition-colors ${
                                visibility === "public"
                                  ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                                  : "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300"
                              }`}
                            >
                              {visibility}
                            </span>
                          )}
                        </div>
                      </Link>
                    );
                  })
                )}
              </div>
            </div>
            <Pagination
              currentPage={repoPage}
              totalPages={repoTotalPages}
              onPageChange={setRepoPage}
              disabled={isLoadingRepositories}
              className="px-6 pb-4 shrink-0"
            />
          </Card>
        </main>
      </div>

      <OrganizationDialogForm
        open={isCreateOrgDialogOpen}
        onOpenChange={setIsCreateOrgDialogOpen}
        onCancel={() => setIsCreateOrgDialogOpen(false)}
      />
    </div>
  );
}
