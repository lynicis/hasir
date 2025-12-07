"use client";

import { getRepository } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { RegistryService } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { Code, ConnectError } from "@connectrpc/connect";
import { useParams, useRouter } from "next/navigation";
import { AlertTriangle, Trash2 } from "lucide-react";
import { useQuery } from "@connectrpc/connect-query";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DeleteRepositoryDialog } from "@/components/delete-repository-dialog";
import { RepositorySettingsForm } from "@/components/repository-settings-form";
import { type OrganizationRepository } from "@/components/repository-item";
import { reverseVisibilityMapper } from "@/lib/visibility-mapper";
import { useRegistryStore } from "@/stores/registry-store";
import { customRetry } from "@/lib/query-retry";
import { Button } from "@/components/ui/button";
import { useClient } from "@/lib/use-client";

export default function SettingsPage() {
  const params = useParams();
  const router = useRouter();
  const repositoryId = params.repositoryId as string;
  const registryApiClient = useClient(RegistryService);
  const invalidateRepositories = useRegistryStore(
    (state) => state.invalidateRepositories
  );

  const [deleteDialog, setDeleteDialog] = useState(false);

  const {
    data: repositoryData,
    isLoading,
    error,
    refetch,
  } = useQuery(
    getRepository,
    { id: repositoryId },
    { retry: customRetry }
  );

  const repository = useMemo((): OrganizationRepository | undefined => {
    if (!repositoryData) return undefined;

    return {
      id: repositoryData.id,
      name: repositoryData.name,
      visibility: reverseVisibilityMapper.get(repositoryData.visibility) || "private",
    };
  }, [repositoryData]);

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

  const handleFormSubmit = async (data: {
    name: string;
    visibility: "private" | "public" | "internal";
    description?: string;
  }) => {
    try {
      // TODO: Implement update API call when available
      console.log("Update repository:", data);
      await new Promise((resolve) => setTimeout(resolve, 1000));
      await refetch();
      invalidateRepositories();
      toast.success("Repository settings updated successfully");
    } catch (error) {
      toast.error("Failed to update repository settings");
      throw error;
    }
  };

  const handleDeleteConfirm = async () => {
    try {
      await registryApiClient.deleteRepository({
        repositoryId,
      });

      toast.success("Repository deleted successfully");
      invalidateRepositories();
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ConnectError) {
        if (err.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete this repository");
        } else if (err.code === Code.NotFound) {
          toast.error("Repository not found");
        } else {
          toast.error("Failed to delete repository");
        }
      } else {
        toast.error("Failed to delete repository");
      }
    } finally {
      setDeleteDialog(false);
    }
  };

  if (!repository && !isLoading) {
    return null;
  }

  return (
    <div className="space-y-6">
      <RepositorySettingsForm
        initialData={{
          name: repository?.name || "",
          visibility: repository?.visibility || "private",
        }}
        onSubmit={handleFormSubmit}
        isLoading={isLoading}
      />

      <Card>
        <CardHeader>
          <CardTitle>Access Control</CardTitle>
          <CardDescription>
            Manage who can access this repository
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-muted-foreground max-w-2xl space-y-4">
            <p className="text-sm">
              Access control settings will be configured here, including:
            </p>
            <ul className="ml-6 list-disc space-y-1 text-sm">
              <li>Team access permissions</li>
              <li>Individual user permissions</li>
              <li>API token access</li>
              <li>Webhook configurations</li>
            </ul>
          </div>
        </CardContent>
      </Card>

      <Card className="border-destructive">
        <CardHeader>
          <div className="flex items-center gap-2">
            <AlertTriangle className="size-5 text-destructive" />
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
          </div>
          <CardDescription>
            Irreversible and destructive actions
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex max-w-2xl items-center justify-between rounded-lg border border-destructive/50 bg-destructive/5 p-4">
            <div className="space-y-1">
              <h4 className="text-sm font-medium">Delete this repository</h4>
              <p className="text-muted-foreground text-xs">
                Once deleted, this repository and all its data cannot be
                recovered.
              </p>
            </div>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setDeleteDialog(true)}
            >
              <Trash2 className="mr-2 size-4" />
              Delete Repository
            </Button>
          </div>
        </CardContent>
      </Card>

      <DeleteRepositoryDialog
        open={deleteDialog}
        repository={repository || null}
        onOpenChange={setDeleteDialog}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
