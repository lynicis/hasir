import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import OrganizationRepositoriesContent from "./organization-repositories-content";

const mockRefetch = vi.fn();
const mockPush = vi.fn();
const mockInvalidateRepositories = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "org-123" }),
  useRouter: () => ({
    push: mockPush,
  }),
  useSearchParams: () => ({
    get: () => null,
    toString: () => "",
  }),
}));

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: () => ({
    data: {
      repositories: [],
      totalPage: 1,
      nextPage: 1,
    },
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    deleteRepository: vi.fn(),
  }),
}));

vi.mock("@/stores/registry-store", () => ({
  useRegistryStore: () => mockInvalidateRepositories,
}));

vi.mock("@/components/repositories-list", () => ({
  RepositoriesList: ({
    repositories,
    isLoading,
  }: {
    repositories: unknown[];
    isLoading: boolean;
  }) => (
    <div data-testid="repositories-list" data-loading={isLoading}>
      {repositories.length} repositories
    </div>
  ),
}));

vi.mock("@/components/delete-repository-dialog", () => ({
  DeleteRepositoryDialog: ({ open }: { open: boolean }) => (
    <div data-testid="delete-repository-dialog" data-open={open}>
      Delete Dialog
    </div>
  ),
}));

vi.mock("@/components/repository-dialog-form", () => ({
  RepositoryDialogForm: ({ open }: { open: boolean }) => (
    <div data-testid="repository-dialog-form" data-open={open}>
      Repository Form
    </div>
  ),
}));

describe("OrganizationRepositoriesContent", () => {
  it("renders the page heading", () => {
    render(<OrganizationRepositoriesContent />);

    expect(
      screen.getByRole("heading", { name: /repository settings/i })
    ).toBeInTheDocument();
  });

  it("renders the page description", () => {
    render(<OrganizationRepositoriesContent />);

    expect(
      screen.getByText(/manage organization repositories/i)
    ).toBeInTheDocument();
  });

  it("renders the repositories list", () => {
    render(<OrganizationRepositoriesContent />);

    expect(screen.getByTestId("repositories-list")).toBeInTheDocument();
  });

  it("renders the delete repository dialog", () => {
    render(<OrganizationRepositoriesContent />);

    expect(screen.getByTestId("delete-repository-dialog")).toBeInTheDocument();
  });

  it("renders the repository creation form dialog", () => {
    render(<OrganizationRepositoriesContent />);

    expect(screen.getByTestId("repository-dialog-form")).toBeInTheDocument();
  });
});
