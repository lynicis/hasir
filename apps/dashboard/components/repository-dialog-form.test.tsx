import type { Mock } from "vitest";

import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, expect, it, vi } from "vitest";
import { ConnectError, Code } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { RepositoryDialogForm } from "./repository-dialog-form";

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccess,
    error: toastError,
  },
}));

const mockUseQuery = vi.fn();

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

const mockUseRegistryStore = vi.fn();

vi.mock("@/stores/registry-store", () => ({
  useRegistryStore: (...args: unknown[]) => mockUseRegistryStore(...args),
}));

vi.mock(
  "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery",
  () => ({
    getOrganizations: { name: "getOrganizations" },
  })
);

const mockCreateRepository = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

const mockOrganizations = [
  { id: "org-1", name: "Acme Corp" },
  { id: "org-2", name: "Hasir Labs" },
  { id: "org-3", name: "Proto Systems" },
];

describe("RepositoryDialogForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      createRepository: mockCreateRepository,
    } as never);
    mockCreateRepository.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
    mockUseQuery.mockReset();

    mockUseRegistryStore.mockReset();

    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: { organizations: mockOrganizations },
          isLoading: false,
          error: null,
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    mockUseRegistryStore.mockImplementation(
      (selector: (state: unknown) => unknown) =>
        selector({
          organizationsVersion: 0,
          repositoriesVersion: 0,
          invalidateOrganizations: vi.fn(),
          invalidateRepositories: vi.fn(),
        })
    );
  });

  function setup(open = true) {
    const onOpenChange = vi.fn();
    const onCancel = vi.fn();

    render(
      <RepositoryDialogForm
        open={open}
        onOpenChange={onOpenChange}
        onCancel={onCancel}
      />
    );

    return { onOpenChange, onCancel };
  }

  it("renders dialog with organization, name and visibility fields when open", () => {
    setup(true);

    expect(
      screen.getByRole("heading", { name: /create repository/i })
    ).toBeInTheDocument();

    expect(screen.getByLabelText(/organization/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /public/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /private/i })).toBeInTheDocument();
  });

  it("creates a public repository when submitted with valid data", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup(true);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /create repository/i })
      ).toBeInTheDocument()
    );

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    await user.click(organizationSelect);

    await waitFor(() => {
      const options = screen.getAllByText("Acme Corp");
      return options.length > 0;
    });
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    await user.type(
      screen.getByLabelText(/name/i),
      "awesome-repository-example"
    );

    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(mockCreateRepository).toHaveBeenCalledWith({
        name: "awesome-repository-example",
        organizationId: "org-2",
        visibility: Visibility.PUBLIC,
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith(
        "Repository created successfully."
      )
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("creates a private repository when private visibility is selected", async () => {
    const user = userEvent.setup();

    setup(true);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /create repository/i })
      ).toBeInTheDocument()
    );

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    await user.click(organizationSelect);

    await waitFor(() => {
      const options = screen.getAllByText("Acme Corp");
      return options.length > 0;
    });
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    await user.type(screen.getByLabelText(/name/i), "secret-repository");
    await user.click(screen.getByRole("radio", { name: /private/i }));
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(mockCreateRepository).toHaveBeenCalledWith({
        name: "secret-repository",
        organizationId: "org-3",
        visibility: Visibility.PRIVATE,
      })
    );
  });

  it("shows an error toast when a ConnectError occurs", async () => {
    const user = userEvent.setup();

    mockCreateRepository.mockRejectedValueOnce(
      new ConnectError("invalid name", Code.InvalidArgument)
    );

    setup(true);

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    await user.click(organizationSelect);
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    await user.type(screen.getByLabelText(/name/i), "bad name");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Failed to create repository.")
    );
  });

  it("shows a generic error toast when an unexpected error occurs", async () => {
    const user = userEvent.setup();

    mockCreateRepository.mockRejectedValueOnce(new Error("network error"));

    setup(true);

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    await user.click(organizationSelect);
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    await user.type(screen.getByLabelText(/name/i), "any-name");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Failed to create repository.")
    );
  });

  it("calls onCancel and resets form when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const { onCancel } = setup(true);

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    await user.click(organizationSelect);
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement;
    await user.type(nameInput, "temp-name");

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
    expect(nameInput.value).toBe("");
  });

  it("shows validation error when organization is not selected", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.type(screen.getByLabelText(/name/i), "test-repository");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/please select an organization/i)
      ).toBeInTheDocument();
    });

    expect(mockCreateRepository).not.toHaveBeenCalled();
  });

  it("shows loading state when organizations are being fetched", () => {
    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: undefined,
          isLoading: true,
          error: null,
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    setup(true);

    const organizationSelect = screen.getByRole("combobox", {
      name: /organization/i,
    });
    expect(organizationSelect).toBeDisabled();
  });

  it("refetches organizations when organizationsVersion changes", async () => {
    const refetchOrganizations = vi.fn();

    mockUseQuery.mockImplementation((schema: { name: string }) => {
      if (schema.name === "getOrganizations") {
        return {
          data: { organizations: mockOrganizations },
          isLoading: false,
          error: null,
          refetch: refetchOrganizations,
        };
      }
      return { data: null, isLoading: false, error: null };
    });

    mockUseRegistryStore.mockImplementation(
      (selector: (state: unknown) => unknown) =>
        selector({
          organizationsVersion: 1,
          repositoriesVersion: 0,
          invalidateOrganizations: vi.fn(),
          invalidateRepositories: vi.fn(),
        })
    );

    setup(true);

    await waitFor(() => {
      expect(refetchOrganizations).toHaveBeenCalled();
    });
  });
});
