"use client";

import type { Repository } from "@hasir/proto/gen/js/registry/v1/registry_pb"

import { createContext } from "react";

export const RepositoryContext = createContext<{
  repository: Repository | undefined;
  isLoading: boolean;
  error: unknown;
  refetch?: () => Promise<unknown>;
} | null>(null);
