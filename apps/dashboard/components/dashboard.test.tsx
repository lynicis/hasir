import { render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useUserStore } from "@/stores/user-store-provider";

import { Dashboard } from "./dashboard";

// eslint-disable-next-line no-var
var toastError: ReturnType<typeof vi.fn>;

vi.mock("@/stores/user-store-provider", () => ({
  useUserStore: vi.fn(),
}));

vi.mock("sonner", () => {
  toastError = vi.fn();
  return {
    toast: {
      error: toastError,
    },
  };
});

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

const mockedUseUserStore = vi.mocked(useUserStore);

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

describe("Dashboard", () => {
  beforeEach(() => {
    mockedUseUserStore.mockReturnValue({ id: "user-123" } as never);
    mockUseQuery.mockReset();
    toastError.mockReset();

    // Default mock implementation: return loaded data
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
          data: { repositories: mockRepositories },
          isLoading: false,
          error: null,
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

    // Should show "All organizations" button even while loading
    expect(
      screen.getByRole("button", { name: /all organizations/i })
    ).toBeInTheDocument();

    // Should show skeletons (using data-slot attribute)
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

    // Find the call for getRepositories
    const repositoriesCall = mockUseQuery.mock.calls.find(
      (call) => call[0]?.name === "getRepositories"
    );

    expect(repositoriesCall).toBeDefined();
    expect(repositoriesCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
      filter: {
        case: "byUserId",
        value: "user-123",
      },
    });
  });

  it("passes correct parameters to useQuery for organizations", () => {
    render(<Dashboard />);

    // Find the call for getOrganizations
    const organizationsCall = mockUseQuery.mock.calls.find(
      (call) => call[0]?.name === "getOrganizations"
    );

    expect(organizationsCall).toBeDefined();
    expect(organizationsCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
      userId: "user-123",
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

    // Find the most recent call for getRepositories
    const repositoriesCalls = mockUseQuery.mock.calls.filter(
      (call) => call[0]?.name === "getRepositories"
    );
    const lastCall = repositoriesCalls[repositoriesCalls.length - 1];

    expect(lastCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
      filter: {
        case: "organizationId",
        value: {
          userId: "user-123",
          organizationId: "org-2",
        },
      },
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

    // Find the most recent call for getRepositories
    const repositoriesCalls = mockUseQuery.mock.calls.filter(
      (call) => call[0]?.name === "getRepositories"
    );
    const lastCall = repositoriesCalls[repositoriesCalls.length - 1];

    expect(lastCall?.[1]).toEqual({
      pagination: {
        page: 1,
        pageLimit: 5,
      },
      filter: {
        case: "byUserId",
        value: "user-123",
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
});
