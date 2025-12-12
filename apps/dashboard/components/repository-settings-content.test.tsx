import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import RepositorySettingsContent from "./repository-settings-content";

const mockRefetch = vi.fn();
const mockPush = vi.fn();
const mockInvalidateRepositories = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => ({ repositoryId: "repo-123" }),
  useRouter: () => ({
    push: mockPush,
  }),
}));

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: () => ({
    data: {
      id: "repo-123",
      name: "test-repo",
      visibility: 1,
    },
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateRepository: vi.fn(),
    deleteRepository: vi.fn(),
  }),
}));

vi.mock("@/stores/registry-store", () => ({
  useRegistryStore: () => mockInvalidateRepositories,
}));

vi.mock("@/components/repository-settings-form", () => ({
  RepositorySettingsForm: ({
    initialData,
    isLoading,
  }: {
    initialData: { name: string; visibility: string };
    isLoading: boolean;
  }) => (
    <div data-testid="repository-settings-form" data-loading={isLoading}>
      Repository: {initialData.name} - {initialData.visibility}
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

describe("RepositorySettingsContent", () => {
  it("renders the repository settings form", () => {
    render(<RepositorySettingsContent />);

    expect(screen.getByTestId("repository-settings-form")).toBeInTheDocument();
  });

  it("renders the Danger Zone card", () => {
    render(<RepositorySettingsContent />);

    expect(screen.getByText("Danger Zone")).toBeInTheDocument();
    expect(
      screen.getByText(/irreversible and destructive actions/i)
    ).toBeInTheDocument();
  });

  it("renders the delete repository section", () => {
    render(<RepositorySettingsContent />);

    expect(
      screen.getByText(/delete this repository/i)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/once deleted, this repository and all its data cannot be recovered/i)
    ).toBeInTheDocument();
  });

  it("renders the delete repository button", () => {
    render(<RepositorySettingsContent />);

    expect(
      screen.getByRole("button", { name: /delete repository/i })
    ).toBeInTheDocument();
  });

  it("renders the delete repository dialog", () => {
    render(<RepositorySettingsContent />);

    expect(screen.getByTestId("delete-repository-dialog")).toBeInTheDocument();
  });
});
