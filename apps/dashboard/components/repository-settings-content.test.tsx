import { render, screen } from "@testing-library/react";

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

describe("RepositorySettingsContent", () => {
  it("renders the repository settings form", () => {
    render(<RepositorySettingsContent />);

    expect(screen.getByText("General Settings")).toBeInTheDocument();
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


});
