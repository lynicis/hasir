import { useCallback, useEffect, useRef, useState } from "react";

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
  const [content, setContent] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const canFetch = enabled && !!organizationId && !!repositoryId && !!commitHash;

  const fetchDocumentation = useCallback(async () => {
    if (!canFetch || !organizationId || !repositoryId) return;

    abortControllerRef.current?.abort();
    const controller = new AbortController();
    abortControllerRef.current = controller;

    setIsLoading(true);
    setError(null);

    try {
      const url = buildDocsUrl(organizationId, repositoryId, commitHash);
      const response = await fetch(url, { signal: controller.signal });

      if (!response.ok) {
        throw new Error(await parseErrorResponse(response));
      }

      const text = await response.text();
      
      if (!controller.signal.aborted) {
        setContent(text || null);
      }
    } catch (err) {
      if (err instanceof DOMException && err.name === "AbortError") {
        return;
      }
      if (!controller.signal.aborted) {
        setError(err instanceof Error ? err.message : "Failed to load documentation");
      }
    } finally {
      if (!controller.signal.aborted) {
        setIsLoading(false);
      }
    }
  }, [canFetch, organizationId, repositoryId, commitHash]);

  useEffect(() => {
    void fetchDocumentation();

    return () => {
      abortControllerRef.current?.abort();
    };
  }, [fetchDocumentation]);

  return {
    content,
    isLoading,
    error,
    refetch: fetchDocumentation,
  };
}


