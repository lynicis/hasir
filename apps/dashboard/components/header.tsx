"use client";

import type { Organization } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { search } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { useRef, useEffect, useState, useSyncExternalStore, useMemo } from "react";
import { useQuery } from "@connectrpc/connect-query";
import { Box, Plus, Search } from "lucide-react";
import { useRouter } from "next/navigation";
import Link from "next/link";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useDebounce } from "@/lib/use-debounce";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Kbd } from "@/components/ui/kbd";

import { OrganizationDialogForm } from "./organization-dialog-form";
import { RepositoryDialogForm } from "./repository-dialog-form";
import { InputGroupAddon } from "./ui/input-group";
import { SearchDropdown } from "./search-dropdown";
import { ModeToggle } from "./theme-toggle";

function useIsMac() {
  return useSyncExternalStore(
    () => () => {},
    () => /mac|iphone|ipad|ipod/i.test(navigator.userAgent),
    () => true
  );
}

const DEFAULT_PAGE_LIMIT = 10;
const SEARCH_DEBOUNCE_MS = 300;

export function Header() {
  const router = useRouter();
  const searchRef = useRef<HTMLInputElement | null>(null);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const [isCreatePopoverOpen, setIsCreatePopoverOpen] = useState(false);
  const [isCreateRepoDialogOpen, setIsCreateRepoDialogOpen] = useState(false);
  const [isCreateOrgDialogOpen, setIsCreateOrgDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [lastSearchResults, setLastSearchResults] = useState<{
    organizations: Organization[];
    repositories: Repository[];
    query: string;
  } | null>(null);
  const isMac = useIsMac();

  const debouncedQuery = useDebounce(searchQuery, SEARCH_DEBOUNCE_MS);

  const searchQueryParams = useMemo(
    () => ({
      query: debouncedQuery,
      pagination: {
        page: 1,
        pageLimit: DEFAULT_PAGE_LIMIT,
      },
    }),
    [debouncedQuery]
  );

  const {
    data: searchResults,
    isLoading,
    error,
  } = useQuery(search, searchQueryParams, {
    enabled: debouncedQuery.length > 0,
  });

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      const isModifierPressed = event.metaKey || event.ctrlKey;
      if (!isModifierPressed || event.key.toLowerCase() !== "k") return;

      const target = event.target as HTMLElement | null;
      const isEditable =
        target &&
        (target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.isContentEditable);
      if (isEditable) return;

      event.preventDefault();
      searchRef.current?.focus();
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const currentResults = useMemo(() => {
    if (searchResults && debouncedQuery.length > 0) {
      return {
        organizations: searchResults.organizations ?? [],
        repositories: searchResults.repositories ?? [],
        query: debouncedQuery,
      };
    }
    return null;
  }, [searchResults, debouncedQuery]);

  const lastResultsToShow = currentResults || lastSearchResults;

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      const isOutsideSearch = searchRef.current && !searchRef.current.contains(target);
      const isOutsideDropdown = dropdownRef.current && !dropdownRef.current.contains(target);

      if (isOutsideSearch && isOutsideDropdown) {
        if (currentResults) {
          setLastSearchResults(currentResults);
        }
        setIsSearchOpen(false);
        setSearchQuery("");
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [currentResults]);

  return (
    <>
      <header className="mx-auto mt-4 flex w-full max-w-6xl items-center gap-4 rounded-full border bg-card/80 px-4 py-2 shadow-sm">
        <Link
          href="/dashboard"
          className="flex items-center gap-2 text-base font-medium text-foreground"
        >
          <Box className="size-5 text-primary" aria-hidden="true" />
          <span>Hasir</span>
        </Link>
        <div className="flex flex-1 justify-center">
          <div className="relative w-full max-w-xl">
            <Search
              className="text-muted-foreground pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2"
              aria-hidden="true"
            />
            <Input
              ref={searchRef}
              type="search"
              placeholder="Search..."
              className="h-9 rounded-full border-0 bg-muted/60 pl-9 pr-16 text-sm shadow-none focus-visible:ring-1 [&::-webkit-search-cancel-button]:hidden"
              aria-label="Search"
              value={searchQuery}
              onChange={(e) => {
                const newValue = e.target.value;
                setSearchQuery(newValue);
                if (newValue.length === 0 && currentResults) {
                  setLastSearchResults(currentResults);
                }

                if (newValue.length > 0) {
                  setIsSearchOpen(true);
                }
              }}
              onFocus={() => {
                if (searchQuery.length === 0 && lastSearchResults) {
                  setIsSearchOpen(true);
                }
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter" && searchQuery.trim()) {
                  if (currentResults) {
                    setLastSearchResults(currentResults);
                  }
                  router.push(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
                  setIsSearchOpen(false);
                  setSearchQuery("");
                }

                if (e.key === "Escape") {
                  if (currentResults) {
                    setLastSearchResults(currentResults);
                  }
                  setIsSearchOpen(false);
                  setSearchQuery("");
                }
              }}
            />
            <InputGroupAddon
              align="inline-end"
              className="absolute inset-y-0 right-3"
            >
              <Kbd>{isMac ? "⌘" : "Ctrl"}</Kbd>
              <Kbd className="pointer-events-none select-none text-[10px]">
                K
              </Kbd>
            </InputGroupAddon>
            {isSearchOpen && (
              <SearchDropdown
                ref={dropdownRef}
                query={lastResultsToShow?.query || ""}
                organizations={lastResultsToShow?.organizations ?? []}
                repositories={lastResultsToShow?.repositories ?? []}
                isLoading={debouncedQuery.length > 0 && isLoading}
                error={debouncedQuery.length > 0 ? error : null}
                onResultClick={() => {
                  if (currentResults) {
                    setLastSearchResults(currentResults);
                  }
                  setIsSearchOpen(false);
                  setSearchQuery("");
                }}
              />
            )}
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Popover
            open={isCreatePopoverOpen}
            onOpenChange={setIsCreatePopoverOpen}
          >
            <PopoverTrigger asChild>
              <Button variant="default" size="default" className="gap-1.5">
                <Plus className="size-4" aria-hidden="true" />
                <span>Create</span>
              </Button>
            </PopoverTrigger>
            <PopoverContent align="end" className="w-52">
              <div className="flex flex-col gap-1">
                <button
                  type="button"
                  className="hover:bg-accent hover:text-accent-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-left"
                  onClick={() => {
                    setIsCreatePopoverOpen(false);
                    setIsCreateOrgDialogOpen(true);
                  }}
                >
                  <span>Create organization</span>
                </button>
                <button
                  type="button"
                  className="hover:bg-accent hover:text-accent-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-left"
                  onClick={() => {
                    setIsCreatePopoverOpen(false);
                    setIsCreateRepoDialogOpen(true);
                  }}
                >
                  <span>Create repository</span>
                </button>
              </div>
            </PopoverContent>
          </Popover>
          <ModeToggle />
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="rounded-full p-0"
                aria-label="Open user menu"
              >
                <Avatar className="h-7 w-7">
                  <AvatarImage src="" alt="User profile" />
                  <AvatarFallback>HS</AvatarFallback>
                </Avatar>
              </Button>
            </PopoverTrigger>
            <PopoverContent align="end" className="w-40">
              <div className="flex flex-col gap-1">
                <button
                  type="button"
                  className="hover:bg-accent hover:text-accent-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-left"
                  onClick={() => router.push("/profile")}
                >
                  <span>Profile</span>
                </button>
                <button
                  type="button"
                  className="hover:bg-accent hover:text-accent-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-left text-destructive"
                  onClick={async () => {
                    await fetch("/api/auth/logout", { method: "POST" });
                    router.push("/login");
                    router.refresh();
                  }}
                >
                  <span>Log out</span>
                </button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </header>

      <RepositoryDialogForm
        open={isCreateRepoDialogOpen}
        onOpenChange={setIsCreateRepoDialogOpen}
        onCancel={() => setIsCreateRepoDialogOpen(false)}
      />

      <OrganizationDialogForm
        open={isCreateOrgDialogOpen}
        onOpenChange={setIsCreateOrgDialogOpen}
        onCancel={() => setIsCreateOrgDialogOpen(false)}
      />
    </>
  );
}
