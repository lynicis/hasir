"use client";

import type { Timestamp } from "@bufbuild/protobuf/wkt";

import { getCommits } from "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery";
import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { GitCommit, User } from "lucide-react";
import { DateTime } from "luxon";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { customRetry } from "@/lib/query-retry";

interface RepositoryCommitsContentProps {
  repositoryId: string;
}

function timestampToDate(timestamp: Timestamp): Date {
  const milliseconds =
    Number(timestamp.seconds) * 1000 + Math.floor(timestamp.nanos / 1000000);
  return DateTime.fromMillis(milliseconds).toJSDate();
}

function getInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

function formatRelativeTime(timestamp: Date): string {
  const date = DateTime.fromJSDate(timestamp);
  const now = DateTime.now();
  const diff = now.diff(date, ["days", "hours", "minutes"]);

  if (diff.days >= 1) {
    return `${Math.floor(diff.days)} day${
      Math.floor(diff.days) > 1 ? "s" : ""
    } ago`;
  } else if (diff.hours >= 1) {
    return `${Math.floor(diff.hours)} hour${
      Math.floor(diff.hours) > 1 ? "s" : ""
    } ago`;
  } else {
    return `${Math.floor(diff.minutes)} minute${
      Math.floor(diff.minutes) > 1 ? "s" : ""
    } ago`;
  }
}

function formatAbsoluteTime(timestamp: Date): string {
  const date = DateTime.fromJSDate(timestamp);
  return date.toFormat("MMM dd, yyyy 'at' HH:mm");
}

function CommitSkeleton() {
  return (
    <div className="flex items-start gap-4">
      <Skeleton className="mt-1 size-10 rounded-full" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-5 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
      </div>
      <Skeleton className="h-6 w-20" />
    </div>
  );
}

export default function RepositoryCommitsContent({
  repositoryId,
}: RepositoryCommitsContentProps) {
  const { data, isLoading, error } = useQuery(
    getCommits,
    { id: repositoryId },
    { retry: customRetry }
  );

  const commits = data?.commits ?? [];

  const isNotFound =
    error instanceof ConnectError && error.code === Code.NotFound;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <GitCommit className="size-5 text-muted-foreground" />
            <CardTitle>Commit History</CardTitle>
          </div>
          <CardDescription className="mt-1.5">
            View the complete commit history for this repository
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {Array.from({ length: 3 }).map((_, index) => (
                <div key={index}>
                  <CommitSkeleton />
                  {index < 2 && <Separator className="mt-4" />}
                </div>
              ))}
            </div>
          ) : error && !isNotFound ? (
            <Alert variant="destructive">
              <AlertTitle>Error loading commits</AlertTitle>
              <AlertDescription>
                {error.message ||
                  "Failed to load commit history. Please try again later."}
              </AlertDescription>
            </Alert>
          ) : commits.length === 0 || isNotFound ? (
            <div className="text-muted-foreground py-8 text-center text-sm">
              No commits found for this repository.
            </div>
          ) : (
            <div className="space-y-4">
              {commits.map((commit, index) => (
                <div key={commit.id}>
                  <div className="flex items-start gap-4">
                    <Avatar className="mt-1 size-10">
                      <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                        {commit.user ? getInitials(commit.user.username) : "??"}
                      </AvatarFallback>
                    </Avatar>

                    <div className="flex-1 space-y-1">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <p className="font-medium leading-tight">
                            {commit.message}
                          </p>
                          <div className="text-muted-foreground mt-1 flex items-center gap-2 text-sm">
                            <User className="size-3" />
                            <span>
                              {commit.user?.username || "Unknown user"}
                            </span>
                            {commit.commitedAt && (
                              <>
                                <span>&middot;</span>
                                <span
                                  title={formatAbsoluteTime(
                                    timestampToDate(commit.commitedAt)
                                  )}
                                  className="cursor-help"
                                >
                                  {formatRelativeTime(
                                    timestampToDate(commit.commitedAt)
                                  )}
                                </span>
                              </>
                            )}
                          </div>
                        </div>
                        <div className="text-muted-foreground flex items-center gap-2 text-xs">
                          <code className="rounded bg-muted px-2 py-1 font-mono">
                            {commit.id.slice(0, 7)}
                          </code>
                        </div>
                      </div>
                    </div>
                  </div>

                  {index < commits.length - 1 && <Separator className="mt-4" />}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
