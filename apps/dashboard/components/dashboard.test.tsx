import { render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useSession } from "@/lib/session-provider";

import { Dashboard } from "./dashboard";

const { toastError } = vi.hoisted(() => ({
  toastError: vi.fn(),
}));

vi.mock("@/lib/session-provider", () => ({
  useSession: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    error: toastError,
  },
}));

const mockUseQuery = vi.fn();

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

vi.mock(
  "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery",
  () => ({
    getOrganizations: { name: "getOrganizations" },
  })
);

vi.mock(
  "@buf/hasir_hasir.connectrpc_query-es/registry/v1/registry-RegistryService_connectquery",
  () => ({
    getRepositories: { name: "getRepositories" },
  })
);

const mockedUseSession = vi.mocked(useSession);

const mockOrganizations = [
  { id: "org-1", name: "Acme Corp" },
  { id: "org-2", name: "Hasir Labs" },
  { id: "org-3", name: "Proto Systems" },
];

const mockRepositories = [
  { id: "repo-1", name: "payment-protos", organizationId: "org-1" },
  { id: "repo-2", name: "analytics-protos", organizationId: "org-2" },
  { id: "repo-3", name: "core-registry", organizationId: "org-3" },
  { id: "repo-4", name: "internal-tools", organizationId: "org-2" },
];

const mockOrganizationsWithPagination = {
  organizations: mockOrganizations,
  totalPage: 1,
};

const mockOrganizationsMultiPage = {
  organizations: [
    { id: "org-1", name: "Acme Corp" },
    { id: "org-2", name: "Hasir Labs" },
    { id: "org-3", name: "Proto Systems" },
    { id: "org-4", name: "Test Corp" },
    { id: "org-5", name: "Demo Labs" },
  ],
  totalPage: 2,
};

const mockRepositoriesWithPagination = {
  repositories: mockRepositories,
  totalPage: 1,
};

describe("Dashboard", () => {
  beforeEach(() => {
    mockedUseSession.mockReturnValue({
      session: { user: { id: "user-123" } },
      loading: false,
      refreshSession: vi.fn(),
    } as never);
    mockUseQuery.mockReset();
    toastError.mockReset();

    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: mockOrganizationsWithPagination,
          isLoading: false,
          error: null,
          refetch: vi.fn(),
        };
      }
      if (schema.name === "getRepositories") {
        return {
          data: mockRepositoriesWithPagination,
          isLoading: false,
          error: null,
          refetch: vi.fn(),
        };
      }
      return { data: null, isLoading: false, error: null };
    });
  });

  it("shows loading skeletons initially", () => {
    mockUseQuery.mockImplementation(() => ({
      data: undefined,
      isLoading: true,
      error: null,
    }));

    render(<Dashboard />);

    expect(
      screen.getByRole("button", { name: /all organizations/i })
    ).toBeInTheDocument();

    const skeletons = document.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders organizations list after loading", async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Acme Corp" })
      ).toBeInTheDocument();
    });

    expect(
      screen.getByRole("button", { name: "Hasir Labs" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Proto Systems" })
    ).toBeInTheDocument();
  });

  it("renders repositories list after loading", async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText("payment-protos")).toBeInTheDocument();
    });

    expect(screen.getByText("analytics-protos")).toBeInTheDocument();
    expect(screen.getByText("core-registry")).toBeInTheDocument();
    expect(screen.getByText("internal-tools")).toBeInTheDocument();
  });

  it("shows repository count after loading", async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText("4 repos")).toBeInTheDocument();
    });
  });

  it("passes correct parameters to useQuery for repositories", () => {
    render(<Dashboard />);

    const repositoriesCall = mockUseQuery.mock.calls.find(
      (call) => call[0]?.name === "getRepositories"
    );

    expect(repositoriesCall).toBeDefined();
    expect(repositoriesCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
    });
  });

  it("passes correct parameters to useQuery for organizations", () => {
    render(<Dashboard />);

    const organizationsCall = mockUseQuery.mock.calls.find(
      (call) => call[0]?.name === "getOrganizations"
    );

    expect(organizationsCall).toBeDefined();
    expect(organizationsCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
    });
  });

  it("updates query parameters when an organization is selected", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Hasir Labs" })
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Hasir Labs" }));

    const repositoriesCalls = mockUseQuery.mock.calls.filter(
      (call) => call[0]?.name === "getRepositories"
    );
    const lastCall = repositoriesCalls[repositoriesCalls.length - 1];

    expect(lastCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
      organizationId: "org-2",
    });
  });

  it("shows organization name when filtering by organization", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Acme Corp" })
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Acme Corp" }));

    await waitFor(() => {
      expect(
        screen.getByText(/showing repositories in acme corp/i)
      ).toBeInTheDocument();
    });
  });

  it("shows 'No repositories found' when there are no repositories", async () => {
    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: { organizations: mockOrganizations },
          isLoading: false,
          error: null,
        };
      }
      if (schema.name === "getRepositories") {
        return {
          data: { repositories: [] },
          isLoading: false,
          error: null,
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(/no repositories found/i)).toBeInTheDocument();
    });
  });

  it("shows error toast when fetching repositories fails", async () => {
    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: { organizations: mockOrganizations },
          isLoading: false,
          error: null,
        };
      }
      if (schema.name === "getRepositories") {
        return {
          data: undefined,
          isLoading: false,
          error: new Error("Network error"),
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    render(<Dashboard />);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while fetching repositories"
      );
    });
  });

  it("shows error toast when fetching organizations fails", async () => {
    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: undefined,
          isLoading: false,
          error: new Error("Network error"),
        };
      }
      if (schema.name === "getRepositories") {
        return {
          data: { repositories: mockRepositories },
          isLoading: false,
          error: null,
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    render(<Dashboard />);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while fetching organizations"
      );
    });
  });

  it("resets to all repositories when 'All organizations' is selected", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Acme Corp" })
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Acme Corp" }));

    await waitFor(() => {
      expect(
        screen.getByText(/showing repositories in acme corp/i)
      ).toBeInTheDocument();
    });

    await user.click(
      screen.getByRole("button", { name: /all organizations/i })
    );

    const repositoriesCalls = mockUseQuery.mock.calls.filter(
      (call) => call[0]?.name === "getRepositories"
    );
    const lastCall = repositoriesCalls[repositoriesCalls.length - 1];

    expect(lastCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
    });

    expect(
      screen.queryByText(/showing repositories in/i)
    ).not.toBeInTheDocument();
  });

  it("highlights the selected organization", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Hasir Labs" })
      ).toBeInTheDocument();
    });

    const hasirLabsButton = screen.getByRole("button", { name: "Hasir Labs" });

    await user.click(hasirLabsButton);

    expect(hasirLabsButton.className).toContain("bg-accent");
  });

  describe("Pagination", () => {
    it("does not show pagination when there is only one page", async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Acme Corp" })
        ).toBeInTheDocument();
      });

      expect(
        screen.queryByRole("button", { name: "1" })
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "2" })
      ).not.toBeInTheDocument();
    });

    it("shows pagination when there are multiple pages", async () => {
      mockUseQuery.mockImplementation((schema: { name: string }) => {
        if (schema.name === "getOrganizations") {
          return {
            data: mockOrganizationsMultiPage,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        if (schema.name === "getRepositories") {
          return {
            data: mockRepositoriesWithPagination,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        return { data: null, isLoading: false, error: null };
      });

      render(<Dashboard />);

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Acme Corp" })
        ).toBeInTheDocument();
      });

      expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
    });

    it("handles organization page changes correctly", async () => {
      const user = userEvent.setup();

      mockUseQuery.mockImplementation((schema: { name: string }) => {
        if (schema.name === "getOrganizations") {
          return {
            data: mockOrganizationsMultiPage,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        if (schema.name === "getRepositories") {
          return {
            data: mockRepositoriesWithPagination,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        return { data: null, isLoading: false, error: null };
      });

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: "2" }));

      const orgCalls = mockUseQuery.mock.calls.filter(
        (call) => call[0]?.name === "getOrganizations"
      );
      const lastOrgCall = orgCalls[orgCalls.length - 1];

      expect(lastOrgCall?.[1]).toEqual({
        pagination: {
          page: 2,
          pageLimit: 5,
        },
      });
    });

    it("resets repository page to 1 when organization changes", async () => {
      const user = userEvent.setup();

      mockUseQuery.mockImplementation((schema: { name: string }) => {
        if (schema.name === "getOrganizations") {
          return {
            data: mockOrganizationsMultiPage,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        if (schema.name === "getRepositories") {
          return {
            data: { repositories: mockRepositories, totalPage: 2 },
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        return { data: null, isLoading: false, error: null };
      });

      render(<Dashboard />);

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Acme Corp" })
        ).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: "Acme Corp" }));

      const repoCalls = mockUseQuery.mock.calls.filter(
        (call) => call[0]?.name === "getRepositories"
      );
      const lastRepoCall = repoCalls[repoCalls.length - 1];

      expect(lastRepoCall?.[1]).toEqual({
        pagination: {
          page: 1,
          pageLimit: 5,
        },
        organizationId: "org-1",
      });
    });

    it("disables pagination controls during loading", async () => {
      mockUseQuery.mockImplementation((schema: { name: string }) => {
        if (schema.name === "getOrganizations") {
          return {
            data: mockOrganizationsMultiPage,
            isLoading: true, // Set loading to true
            error: null,
            refetch: vi.fn(),
          };
        }
        if (schema.name === "getRepositories") {
          return {
            data: mockRepositoriesWithPagination,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
          };
        }
        return { data: null, isLoading: false, error: null };
      });

      render(<Dashboard />);

      await waitFor(() => {
        const pageButton = screen.queryByRole("button", { name: "1" });
        if (pageButton) {
          expect(pageButton).toBeDisabled();
        }
      });
    });
  });
});
