"use client";

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
  const [isCreatePopoverOpen, setIsCreatePopoverOpen] = useState(false);
  const [isCreateRepoDialogOpen, setIsCreateRepoDialogOpen] = useState(false);
  const [isCreateOrgDialogOpen, setIsCreateOrgDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [isSearchOpen, setIsSearchOpen] = useState(false);
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

  useEffect(() => {
    setIsSearchOpen(debouncedQuery.length > 0);
  }, [debouncedQuery]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        searchRef.current &&
        !searchRef.current.contains(event.target as Node)
      ) {
        setIsSearchOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

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
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && searchQuery.trim()) {
                  router.push(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
                  setIsSearchOpen(false);
                }
                if (e.key === "Escape") {
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
                query={debouncedQuery}
                organizations={searchResults?.organizations ?? []}
                repositories={searchResults?.repositories ?? []}
                isLoading={isLoading}
                error={error}
                onResultClick={() => {
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
