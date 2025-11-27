"use client";

import { useRef, useEffect, useState, useSyncExternalStore } from "react";
import { Box, Plus, Search } from "lucide-react";
import Link from "next/link";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Kbd } from "@/components/ui/kbd";

import { OrganizationDialogForm } from "./organization-dialog-form";
import { RepositoryDialogForm } from "./repository-dialog-form";
import { InputGroupAddon } from "./ui/input-group";
import { ModeToggle } from "./theme-toggle";

function useIsMac() {
  return useSyncExternalStore(
    () => () => {},
    () => /mac|iphone|ipad|ipod/i.test(navigator.userAgent),
    () => true
  );
}

export function Header() {
  const searchRef = useRef<HTMLInputElement | null>(null);
  const [isCreatePopoverOpen, setIsCreatePopoverOpen] = useState(false);
  const [isCreateRepoDialogOpen, setIsCreateRepoDialogOpen] = useState(false);
  const [isCreateOrgDialogOpen, setIsCreateOrgDialogOpen] = useState(false);
  const isMac = useIsMac();

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
          <Button
            variant="outline"
            size="icon"
            asChild
            className="rounded-full p-0"
          >
            <Link href="/profile" aria-label="Profile">
              <Avatar className="h-7 w-7">
                <AvatarImage src="" alt="User profile" />
                <AvatarFallback>HS</AvatarFallback>
              </Avatar>
            </Link>
          </Button>
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
