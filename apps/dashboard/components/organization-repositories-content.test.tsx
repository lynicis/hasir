import { render, screen } from "@testing-library/react";

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
      screen.getAllByText(/manage organization repositories/i)[0]
    ).toBeInTheDocument();
  });

  it("renders the repositories list", () => {
    render(<OrganizationRepositoriesContent />);

    expect(screen.getByText("Repositories")).toBeInTheDocument();
  });
});
