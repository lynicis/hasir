"use client";

import { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { createContext } from "react";

export const RepositoryContext = createContext<{
  repository: Repository | undefined;
  isLoading: boolean;
  error: unknown;
} | null>(null);
