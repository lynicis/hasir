import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { RepositoryContext } from "@/lib/repository-context";

import RepositoryDocumentationContent from "./repository-documentation-content";

const mockFetch = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => ({ repositoryId: "repo-123" }),
}));

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: () => ({
    data: { id: "commit-abc123" },
    isLoading: false,
  }),
}));

function renderWithContext(
  ui: React.ReactElement,
  repository?: Partial<Repository>
) {
  const value = {
    repository: {
      id: "repo-123",
      name: "Test Repository",
      organizationId: "org-456",
      visibility: 0,
      sdkPreferences: [],
      ...repository,
    } as Repository,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  };

  return render(
    <RepositoryContext.Provider value={value}>{ui}</RepositoryContext.Provider>
  );
}

describe("RepositoryDocumentationContent", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => "# Test Docs",
    });

    (globalThis as unknown as { fetch: typeof fetch }).fetch =
      mockFetch as unknown as typeof fetch;
  });

  it("renders the Documentation card title", async () => {
    renderWithContext(<RepositoryDocumentationContent />);

    expect(await screen.findByText("Documentation")).toBeInTheDocument();
  });

  it("renders the card description", async () => {
    renderWithContext(<RepositoryDocumentationContent />);

    expect(
      await screen.findByText(/view and manage repository documentation/i)
    ).toBeInTheDocument();
  });

  it("loads and renders markdown content from the docs endpoint", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => "# Test Docs",
    });

    renderWithContext(<RepositoryDocumentationContent />);

    expect(await screen.findByText("Test Docs")).toBeInTheDocument();
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/docs/org-456/repo-123/commit-abc123",
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    );
  });

  it("handles authentication errors", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({ error: "Unauthenticated" }),
    });

    renderWithContext(<RepositoryDocumentationContent />);

    expect(await screen.findByText(/unauthenticated/i)).toBeInTheDocument();
  });

  it("shows a fallback message when no documentation is returned", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => "",
    });

    renderWithContext(<RepositoryDocumentationContent />);

    expect(
      await screen.findByText(
        /no documentation is available for this repository yet/i
      )
    ).toBeInTheDocument();
  });
});
