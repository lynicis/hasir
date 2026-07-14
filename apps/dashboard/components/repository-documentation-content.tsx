"use client";

import { getRecentCommit } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { useParams } from "next/navigation";
import { Book } from "lucide-react";
import { useContext } from "react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { RepositoryContext } from "@/lib/repository-context";
import { useDocumentation } from "@/lib/use-documentation";
import { Skeleton } from "@/components/ui/skeleton";
import { customRetry } from "@/lib/query-retry";

function DocumentationSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-6 w-40" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-11/12" />
      <Skeleton className="h-4 w-10/12" />
      <Skeleton className="h-64 w-full" />
    </div>
  );
}

function EmptyDocumentation() {
  return (
    <div className="text-muted-foreground space-y-2 text-sm">
      <p>No documentation is available for this repository yet.</p>
      <p>
        Once documentation is added to the registry, it will be displayed here.
      </p>
    </div>
  );
}

export default function RepositoryDocumentationContent() {
  const context = useContext(RepositoryContext);
  const params = useParams();
  const repositoryId = params.repositoryId as string;

  if (!context) {
    throw new Error(
      "RepositoryDocumentationContent must be used within RepositoryContext"
    );
  }

  const { repository, isLoading: isRepositoryLoading } = context;

  const { data: recentCommit, isLoading: isCommitLoading } = useQuery(
    getRecentCommit,
    { repositoryId },
    { retry: customRetry }
  );

  const {
    content,
    isLoading: isDocLoading,
    error,
  } = useDocumentation({
    organizationId: repository?.organizationId,
    repositoryId: repository?.id,
    commitHash: recentCommit?.id ?? "latest",
    enabled: !isRepositoryLoading && !isCommitLoading && !!repository,
  });

  const isLoading = isRepositoryLoading || isCommitLoading || isDocLoading;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Book className="size-5 text-muted-foreground" />
            <CardTitle>Documentation</CardTitle>
          </div>
          <CardDescription>
            View and manage repository documentation
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <DocumentationSkeleton />
          ) : error ? (
            <Alert variant="destructive">
              <AlertTitle>Error loading documentation</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : content ? (
            <MarkdownRenderer content={content} />
          ) : (
            <EmptyDocumentation />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
