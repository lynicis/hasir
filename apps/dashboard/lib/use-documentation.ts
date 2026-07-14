import { useQuery } from "@tanstack/react-query";

interface UseDocumentationOptions {
  organizationId: string | undefined;
  repositoryId: string | undefined;
  commitHash: string;
  enabled?: boolean;
}

interface UseDocumentationResult {
  content: string | null;
  isLoading: boolean;
  error: string | null;
  refetch: () => void;
}

function buildDocsUrl(
  organizationId: string,
  repositoryId: string,
  commitHash: string
): string {
  return `/api/docs/${encodeURIComponent(organizationId)}/${encodeURIComponent(repositoryId)}/${encodeURIComponent(commitHash)}`;
}

async function parseErrorResponse(
  response: Response
): Promise<string> {
  try {
    const data = await response.json();
    return data.error ?? data.details ?? `Request failed (${response.status})`;
  } catch {
    try {
      return await response.text();
    } catch {
      return `Request failed (${response.status})`;
    }
  }
}

export function useDocumentation({
  organizationId,
  repositoryId,
  commitHash,
  enabled = true,
}: UseDocumentationOptions): UseDocumentationResult {
  const canFetch = enabled && !!organizationId && !!repositoryId && !!commitHash;

  const queryKey = ["docs", organizationId, repositoryId, commitHash] as const;

  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: async ({ signal }) => {
      const url = buildDocsUrl(organizationId!, repositoryId!, commitHash);
      const response = await fetch(url, { signal });

      if (!response.ok) {
        throw new Error(await parseErrorResponse(response));
      }

      return (await response.text()) || null;
    },
    enabled: canFetch,
    retry: false,
  });

  return {
    content: data ?? null,
    isLoading,
    error: error ? (error instanceof Error ? error.message : String(error)) : null,
    refetch: () => { void refetch(); },
  };
}


