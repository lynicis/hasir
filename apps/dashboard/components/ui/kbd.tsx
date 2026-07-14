"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

function Kbd({
  className,
  ...props
}: React.ComponentProps<"kbd">) {
  return (
    <kbd
      data-slot="kbd"
      className={cn(
        "inline-flex items-center justify-center gap-1 rounded border bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground",
        className
      )}
      {...props}
    />
  );
}

export { Kbd };
