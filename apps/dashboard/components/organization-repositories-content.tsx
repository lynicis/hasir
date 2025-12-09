"use client";

import { getRepositories } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { RegistryService } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { DeleteRepositoryDialog } from "@/components/delete-repository-dialog";
import { type OrganizationRepository } from "@/components/repository-item";
import { RepositoryDialogForm } from "@/components/repository-dialog-form";
import { RepositoriesList } from "@/components/repositories-list";
import { reverseVisibilityMapper } from "@/lib/visibility-mapper";
import { useRegistryStore } from "@/stores/registry-store";
import { customRetry } from "@/lib/query-retry";
import { isNotFoundError } from "@/lib/utils";
import { useClient } from "@/lib/use-client";

const PAGE_SIZE = 10;

export default function OrganizationRepositoriesContent() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const organizationId = params.id as string;
  const registryApiClient = useClient(RegistryService);
  const invalidateRepositories = useRegistryStore(
    (state) => state.invalidateRepositories
  );

  const currentPage = useMemo(() => {
    const page = searchParams.get("page");
    return page ? Math.max(1, parseInt(page, 10)) : 1;
  }, [searchParams]);

  const {
    data: repositoriesData,
    isLoading,
    error,
    refetch,
  } = useQuery(
    getRepositories,
    {
      pagination: {
        page: currentPage,
        pageLimit: PAGE_SIZE,
      },
      organizationId,
    },
    { retry: customRetry }
  );

  const repositories = useMemo((): OrganizationRepository[] => {
    return (
      repositoriesData?.repositories?.map(({ id, name, visibility }) => ({
        id,
        name,
        visibility: reverseVisibilityMapper.get(visibility) || "private",
      })) ?? []
    );
  }, [repositoriesData]);

  const totalRepositories = repositories.length;
  const totalPages = repositoriesData?.totalPage ?? 1;
  const nextPage = repositoriesData?.nextPage ?? 1;
  const hasNextPage = currentPage > nextPage;
  const hasPreviousPage = currentPage < nextPage;

  const [deleteRepoDialog, setDeleteRepoDialog] = useState<{
    open: boolean;
    repo: OrganizationRepository | null;
  }>({ open: false, repo: null });

  const [createRepoDialog, setCreateRepoDialog] = useState(false);

  useEffect(() => {
    if (error && !isNotFoundError(error)) {
      toast.error("Error occurred while fetching repositories");
    }
  }, [error]);

  function handleDeleteRepository(repo: OrganizationRepository) {
    setDeleteRepoDialog({ open: true, repo });
  }

  async function confirmDeleteRepository() {
    if (!deleteRepoDialog.repo) return;

    try {
      await registryApiClient.deleteRepository({
        repositoryId: deleteRepoDialog.repo.id,
      });

      toast.success(
        `Repository ${deleteRepoDialog.repo.name} has been deleted.`
      );

      await refetch();
      invalidateRepositories();

      setDeleteRepoDialog({ open: false, repo: null });
    } catch (err) {
      if (err instanceof ConnectError) {
        if (err.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete this repository.");
          setDeleteRepoDialog({ open: false, repo: null });
          return;
        }

        if (err.code === Code.NotFound) {
          toast.error("Repository not found.");
          setDeleteRepoDialog({ open: false, repo: null });
          return;
        }
      }

      toast.error("Failed to delete repository. Please try again.");
      setDeleteRepoDialog({ open: false, repo: null });
    }
  }

  function handleCreateRepository() {
    setCreateRepoDialog(true);
  }

  function handlePageChange(newPage: number) {
    const params = new URLSearchParams(searchParams.toString());
    if (newPage === 1) {
      params.delete("page");
    } else {
      params.set("page", newPage.toString());
    }
    router.push(`?${params.toString()}`);
  }

  return (
    <div className="h-full space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Repository Settings</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage organization repositories
        </p>
      </div>
      <RepositoriesList
        repositories={repositories}
        isLoading={isLoading}
        currentPage={currentPage}
        totalPages={totalPages}
        totalRepositories={totalRepositories}
        hasNextPage={hasNextPage}
        hasPreviousPage={hasPreviousPage}
        onPageChange={handlePageChange}
        onDelete={handleDeleteRepository}
        onCreate={handleCreateRepository}
      />
      <DeleteRepositoryDialog
        open={deleteRepoDialog.open}
        repository={deleteRepoDialog.repo}
        onOpenChange={(open) =>
          setDeleteRepoDialog({ open, repo: deleteRepoDialog.repo })
        }
        onConfirm={confirmDeleteRepository}
      />
      <RepositoryDialogForm
        open={createRepoDialog}
        onOpenChange={setCreateRepoDialog}
        onCancel={() => setCreateRepoDialog(false)}
      />
    </div>
  );
}
