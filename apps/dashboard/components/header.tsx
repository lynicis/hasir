"use client";

import { Box, Plus, Search } from "lucide-react";
import * as React from "react";
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

export function Header() {
  const searchRef = React.useRef<HTMLInputElement | null>(null);

  React.useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const isMetaK = event.metaKey && event.key.toLowerCase() === "k";
      const isCtrlK = event.ctrlKey && event.key.toLowerCase() === "k";

      if (isMetaK || isCtrlK) {
        event.preventDefault();
        searchRef.current?.focus();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  return (
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
            className="h-9 rounded-full border-0 bg-muted/60 pl-9 pr-16 text-sm shadow-none focus-visible:ring-1"
            aria-label="Search"
          />
          <div className="pointer-events-none absolute inset-y-0 right-3 flex items-center">
            <Kbd className="pointer-events-none select-none text-[10px]">
              ⌘K
            </Kbd>
          </div>
        </div>
      </div>

      {/* Right: icon + create + profile (right-to-left visual order inside) */}
      <div className="flex items-center gap-3">
        <Popover>
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
              >
                <span>Create organization</span>
              </button>
              <button
                type="button"
                className="hover:bg-accent hover:text-accent-foreground flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-left"
              >
                <span>Create repository</span>
              </button>
            </div>
          </PopoverContent>
        </Popover>
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
  );
}
