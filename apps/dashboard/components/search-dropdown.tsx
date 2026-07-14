"use client";

import type { Organization } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { Building2, Package, Search as SearchIcon } from "lucide-react";
import { forwardRef } from "react";
import Link from "next/link";

import { Skeleton } from "@/components/ui/skeleton";

interface SearchDropdownProps {
  query: string;
  organizations: Organization[];
  repositories: Repository[];
  isLoading: boolean;
  error: Error | null;
  onResultClick: () => void;
}

export const SearchDropdown = forwardRef<HTMLDivElement, SearchDropdownProps>(
  function SearchDropdown(
    { query, organizations, repositories, isLoading, error, onResultClick },
    ref
  ) {
    const hasResults = organizations.length > 0 || repositories.length > 0;
    if (!query && !hasResults) {
      return null;
    }

    return (
      <div
        ref={ref}
        className="absolute left-0 right-0 top-full z-50 mt-2 max-h-[80vh] overflow-y-auto rounded-2xl border bg-card shadow-lg"
      >
      <div className="p-4">
        {isLoading && (
          <div className="space-y-3">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        )}

        {error && (
          <div className="text-destructive py-4 text-center text-sm">
            Error loading search results: {error.message}
          </div>
        )}

        {!isLoading && !error && !hasResults && query && (
          <div className="py-8 text-center">
            <SearchIcon className="text-muted-foreground mx-auto mb-3 h-8 w-8" />
            <p className="text-muted-foreground text-sm">
              No results found for &quot;{query}&quot;
            </p>
          </div>
        )}

        {!isLoading && !error && hasResults && (
          <div className="space-y-4">
            {organizations.length > 0 && (
              <div>
                <h3 className="text-muted-foreground mb-2 text-xs font-medium uppercase">
                  Organizations ({organizations.length})
                </h3>
                <div className="space-y-1">
                  {organizations.map((org) => (
                    <Link
                      key={org.id}
                      href={`/organization/${org.id}/users`}
                      onClick={onResultClick}
                      className="hover:bg-accent block rounded-lg p-3 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <div className="bg-primary/10 text-primary flex h-9 w-9 items-center justify-center rounded-lg">
                          <Building2 className="h-4 w-4" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="font-medium text-sm">{org.name}</div>
                          <div className="text-muted-foreground text-xs">
                            {org.visibility === 0 ? "Public" : "Private"}{" "}
                            organization
                          </div>
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {repositories.length > 0 && (
              <div>
                <h3 className="text-muted-foreground mb-2 text-xs font-medium uppercase">
                  Repositories ({repositories.length})
                </h3>
                <div className="space-y-1">
                  {repositories.map((repo) => (
                    <Link
                      key={repo.id}
                      href={`/repository/${repo.id}`}
                      onClick={onResultClick}
                      className="hover:bg-accent block rounded-lg p-3 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <div className="bg-accent text-accent-foreground flex h-9 w-9 items-center justify-center rounded-lg">
                          <Package className="h-4 w-4" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="font-medium text-sm">{repo.name}</div>
                          <div className="text-muted-foreground text-xs">
                            {repo.visibility === 0 ? "Public" : "Private"}{" "}
                            repository
                          </div>
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {(organizations.length > 0 || repositories.length > 0) && (
              <div className="border-t pt-3">
                <Link
                  href={`/search?q=${encodeURIComponent(query)}`}
                  onClick={onResultClick}
                  className="text-primary hover:underline block text-center text-sm"
                >
                  View all results
                </Link>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
});
