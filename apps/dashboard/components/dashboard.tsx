"use client";

import { getOrganizations } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { getRepositories } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { customRetry } from "@/lib/query-retry";
import { isNotFoundError } from "@/lib/utils";

const DEFAULT_PAGINATION = {
  page: 1,
  pageLimit: 5,
};

export function Dashboard() {
  const [activeOrgId, setActiveOrgId] = useState<string | "all">("all");

  const {
    data: organizations,
    isLoading: isLoadingOrganizations,
    error: organizationsError,
  } = useQuery(
    getOrganizations,
    {
      pagination: DEFAULT_PAGINATION,
    },
    { retry: customRetry }
  );

  const {
    data: repositoriesData,
    isLoading: isLoadingRepositories,
    error: repositoriesError,
  } = useQuery(
    getRepositories,
    activeOrgId === "all"
      ? {
          pagination: DEFAULT_PAGINATION,
        }
      : {
          pagination: DEFAULT_PAGINATION,
          organizationId: activeOrgId,
        },
    { retry: customRetry }
  );

  const organizationsList = organizations?.organizations ?? [];
  const repositoriesList = repositoriesData?.repositories ?? [];

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

  return (
    <div className="h-[calc(100vh-4.5rem)] bg-background px-6 py-6 overflow-hidden">
      <div className="mx-auto flex h-full w-full max-w-6xl flex-col gap-6">
        <main className="grid flex-1 grid-cols-[260px_minmax(0,1fr)] gap-6 pt-2">
          <Card className="h-full gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
            <CardHeader className="flex items-center bg-primary px-6 py-4">
              <CardTitle className="text-sm font-medium text-secondary">
                Your organizations
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 py-4">
              <button
                type="button"
                onClick={() => setActiveOrgId("all")}
                className={`flex w-full items-center justify-between rounded-md px-3 py-2.5 text-sm transition-colors ${
                  activeOrgId === "all"
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                }`}
              >
                <span>All organizations</span>
              </button>
              {isLoadingOrganizations ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <div
                    key={i}
                    className="flex w-full items-center rounded-md px-3 py-2.5"
                  >
                    <Skeleton className="h-4 w-24" />
                  </div>
                ))
              ) : organizationsError && !isNotFoundError(organizationsError) ? (
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
                    <button
                      key={org.id}
                      type="button"
                      onClick={() => setActiveOrgId(org.id)}
                      className={`flex w-full items-center justify-between rounded-md px-3 py-2.5 text-sm transition-colors ${
                        isActive
                          ? "bg-accent text-accent-foreground"
                          : "hover:bg-accent hover:text-accent-foreground"
                      }`}
                    >
                      <span>{org.name}</span>
                    </button>
                  );
                })
              )}
            </CardContent>
          </Card>
          <Card className="h-full gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
            <CardHeader className="flex items-center justify-between bg-primary px-6 py-4">
              <div className="space-y-0.5">
                <CardTitle className="text-sm font-medium text-secondary">
                  Repositories
                </CardTitle>
                {activeOrgId !== "all" && (
                  <p className="text-xs text-secondary/70">
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
                <span className="text-xs text-secondary/70">
                  {repositoriesList.length} repos
                </span>
              )}
            </CardHeader>
            <CardContent className="space-y-2.5 py-4">
              {isLoadingRepositories ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between rounded-xl border border-border/60 bg-card px-4 py-3"
                  >
                    <Skeleton className="h-5 w-32" />
                  </div>
                ))
              ) : repositoriesList.length === 0 ? (
                <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
                  No repositories found
                </div>
              ) : (
                repositoriesList.map((repo) => (
                  <div
                    key={repo.id}
                    className="hover:bg-accent/60 flex items-center justify-between rounded-xl border border-border/60 bg-card px-4 py-3 text-sm transition-colors"
                  >
                    <div className="flex flex-col">
                      <span className="font-medium">{repo.name}</span>
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  );
}
