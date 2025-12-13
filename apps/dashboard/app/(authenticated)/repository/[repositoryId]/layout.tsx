"use client";

import { getRepository } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import {
  ArrowLeft,
  Book,
  Files,
  GitCommit,
  Settings,
  Wrench,
} from "lucide-react";
import { useParams, usePathname, useRouter } from "next/navigation";
import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo } from "react";
import { toast } from "sonner";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { type OrganizationRepository } from "@/components/repository-item";
import { RepositoryContext } from "@/components/repository-context";
import { reverseVisibilityMapper } from "@/lib/visibility-mapper";
import { CloneUrls } from "@/components/clone-urls";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { customRetry } from "@/lib/query-retry";

type TabType =
  | "documentation"
  | "files"
  | "commits"
  | "sdk-preferences"
  | "settings";

export default function RepositoryLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams();
  const router = useRouter();
  const pathname = usePathname();
  const repositoryId = params.repositoryId as string;

  const {
    data: repositoryData,
    isLoading,
    error,
  } = useQuery(getRepository, { id: repositoryId }, { retry: customRetry });

  const repository = useMemo((): OrganizationRepository | undefined => {
    if (!repositoryData) return undefined;

    return {
      id: repositoryData.id,
      name: repositoryData.name,
      visibility:
        reverseVisibilityMapper.get(repositoryData.visibility) || "private",
    };
  }, [repositoryData]);

  const activeTab = useMemo((): TabType => {
    if (pathname.includes("/documentation")) return "documentation";
    if (pathname.includes("/files")) return "files";
    if (pathname.includes("/commits")) return "commits";
    if (pathname.includes("/sdk-preferences")) return "sdk-preferences";
    if (pathname.includes("/settings")) return "settings";
    return "documentation";
  }, [pathname]);

  useEffect(() => {
    if (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          toast.error("Repository not found");
          router.push("/dashboard");
          return;
        }
        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to view this repository");
          router.push("/dashboard");
          return;
        }
      }
      toast.error("Failed to load repository");
    }
  }, [error, router]);

  useEffect(() => {
    if (!isLoading && !repository && !pathname.endsWith(repositoryId)) {
      toast.error("Repository not found");
      router.push("/dashboard");
    }
  }, [isLoading, repository, router, pathname, repositoryId]);

  const tabs = [
    {
      id: "documentation" as const,
      label: "Documentation",
      icon: Book,
      href: `/repository/${repositoryId}/documentation`,
    },
    {
      id: "files" as const,
      label: "Files",
      icon: Files,
      href: `/repository/${repositoryId}/files`,
    },
    {
      id: "commits" as const,
      label: "Commits",
      icon: GitCommit,
      href: `/repository/${repositoryId}/commits`,
    },
    {
      id: "sdk-preferences" as const,
      label: "SDK Generation",
      icon: Wrench,
      href: `/repository/${repositoryId}/sdk-preferences`,
    },
    {
      id: "settings" as const,
      label: "Settings",
      icon: Settings,
      href: `/repository/${repositoryId}/settings`,
    },
  ];

  const handleBack = () => {
    router.push("/dashboard");
  };

  const handleNavigate = (href: string) => {
    router.push(href);
  };

  if (isLoading) {
    return (
      <div className="h-[calc(100vh-4.5rem)] overflow-hidden bg-background px-6 py-6">
        <div className="mx-auto flex h-full w-full max-w-6xl gap-6">
          <Card className="h-full w-64 gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
            <CardHeader className="flex items-center bg-primary px-6 py-4">
              <Skeleton className="h-5 w-20" />
            </CardHeader>
            <CardContent className="space-y-2 py-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </CardContent>
          </Card>
          <div className="flex-1 space-y-6 overflow-y-auto">
            <div className="flex items-center gap-4">
              <Skeleton className="size-10 shrink-0" />
              <div>
                <Skeleton className="h-8 w-48" />
                <Skeleton className="mt-1 h-4 w-64" />
              </div>
            </div>
            <Skeleton className="h-64 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (!repository && !pathname.endsWith(repositoryId)) {
    return null;
  }

  return (
    <div className="h-[calc(100vh-4.5rem)] overflow-hidden bg-background px-6 py-6">
      <div className="mx-auto flex h-full w-full max-w-6xl gap-6">
        <Card className="h-full w-64 gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
          <CardHeader className="flex items-center bg-primary px-6 py-4">
            <CardTitle className="text-sm font-medium text-secondary">
              {repository?.name || "Loading..."}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 py-4">
            <nav className="space-y-1">
              {tabs.map((tab) => {
                const Icon = tab.icon;
                const isActive = activeTab === tab.id;

                return (
                  <button
                    key={tab.id}
                    onClick={() => handleNavigate(tab.href)}
                    className={`flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors ${
                      isActive
                        ? "bg-accent text-accent-foreground"
                        : "hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    <Icon className="size-4" />
                    <span>{tab.label}</span>
                  </button>
                );
              })}
            </nav>
          </CardContent>
        </Card>

        <main className="flex-1 space-y-6 overflow-y-auto">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={handleBack}
              className="shrink-0"
            >
              <ArrowLeft className="size-4" />
            </Button>
            <div>
              <h1 className="text-2xl font-semibold">
                {repository?.name || "Loading..."}
              </h1>
              <p className="text-muted-foreground mt-1 text-sm">
                Repository details and configuration
              </p>
            </div>
            <div className="ml-auto">
              {repository && <CloneUrls repositoryId={repository.id} />}
            </div>
          </div>

          <RepositoryContext.Provider
            value={{ repository: repositoryData, isLoading, error }}
          >
            <div>{children}</div>
          </RepositoryContext.Provider>
        </main>
      </div>
    </div>
  );
}
