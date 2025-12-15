import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";

import { useDocumentation } from "@/lib/use-documentation";

const mockFetch = vi.fn();

describe("useDocumentation", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    (globalThis as unknown as { fetch: typeof fetch }).fetch =
      mockFetch as unknown as typeof fetch;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const defaultOptions = {
    organizationId: "org-123",
    repositoryId: "repo-456",
    commitHash: "abc123",
  };

  describe("successful fetching", () => {
    it("fetches documentation on mount", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => "# Documentation",
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.content).toBe("# Documentation");
      expect(result.current.error).toBeNull();
    });

    it("constructs correct URL with encoded parameters", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => "content",
      });

      renderHook(() =>
        useDocumentation({
          organizationId: "org/with/slashes",
          repositoryId: "repo name",
          commitHash: "hash#special",
        })
      );

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/docs/org%2Fwith%2Fslashes/repo%20name/hash%23special",
        expect.any(Object)
      );
    });

    it("passes abort signal to fetch", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => "content",
      });

      renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          signal: expect.any(AbortSignal),
        })
      );
    });
  });

  describe("loading state", () => {
    it("sets isLoading to true while fetching", async () => {
      let resolvePromise: (value: unknown) => void;
      const fetchPromise = new Promise((resolve) => {
        resolvePromise = resolve;
      });

      mockFetch.mockReturnValueOnce(fetchPromise);

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      expect(result.current.isLoading).toBe(true);

      await act(async () => {
        resolvePromise!({
          ok: true,
          text: async () => "content",
        });
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe("error handling", () => {
    it("handles JSON error responses", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.error).toBe("Unauthorized");
      expect(result.current.content).toBeNull();
    });

    it("handles text error responses when JSON fails", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error("Not JSON");
        },
        text: async () => "Internal Server Error",
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.error).toBe("Internal Server Error");
    });

    it("handles network errors", async () => {
      mockFetch.mockRejectedValueOnce(new Error("Network error"));

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.error).toBe("Network error");
    });
  });

  describe("enabled option", () => {
    it("does not fetch when enabled is false", async () => {
      renderHook(() =>
        useDocumentation({
          ...defaultOptions,
          enabled: false,
        })
      );

      await new Promise((r) => setTimeout(r, 50));

      expect(mockFetch).not.toHaveBeenCalled();
    });

    it("does not fetch when organizationId is undefined", async () => {
      renderHook(() =>
        useDocumentation({
          ...defaultOptions,
          organizationId: undefined,
        })
      );

      await new Promise((r) => setTimeout(r, 50));

      expect(mockFetch).not.toHaveBeenCalled();
    });

    it("does not fetch when repositoryId is undefined", async () => {
      renderHook(() =>
        useDocumentation({
          ...defaultOptions,
          repositoryId: undefined,
        })
      );

      await new Promise((r) => setTimeout(r, 50));

      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe("refetch", () => {
    it("provides a refetch function", async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        text: async () => "content",
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(typeof result.current.refetch).toBe("function");
    });

    it("refetches when refetch is called", async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        text: async () => "content",
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(mockFetch).toHaveBeenCalledTimes(1);

      await act(async () => {
        result.current.refetch();
      });

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe("empty content", () => {
    it("sets content to null when response is empty", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => "",
      });

      const { result } = renderHook(() => useDocumentation(defaultOptions));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.content).toBeNull();
    });
  });
});
