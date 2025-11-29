import type { Mock } from "vitest";

import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { OrganizationSettingsForm } from "./organization-settings-form";

var toastSuccess: ReturnType<typeof vi.fn>;

var toastError: ReturnType<typeof vi.fn>;

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useParams: vi.fn(() => ({ id: "org-123" })),
  useRouter: vi.fn(() => ({
    push: mockPush,
  })),
}));

const mockUseQuery = vi.fn();
vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("sonner", () => {
  toastSuccess = vi.fn();
  toastError = vi.fn();
  return {
    toast: {
      success: toastSuccess,
      error: toastError,
    },
  };
});

const mockUpdateOrganization = vi.fn();
const mockDeleteOrganization = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

describe("OrganizationSettingsForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      updateOrganization: mockUpdateOrganization,
      deleteOrganization: mockDeleteOrganization,
    } as never);
    mockUpdateOrganization.mockReset();
    mockDeleteOrganization.mockReset();
    mockPush.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  it("shows loading state when fetching organization", () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    expect(screen.getByText("General Settings")).toBeInTheDocument();
    expect(
      screen.getByText("Update your organization information")
    ).toBeInTheDocument();

    const skeletons = document.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("shows not found message when organization is not found", () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    expect(screen.getByText("General Settings")).toBeInTheDocument();
    expect(screen.getByText("Organization not found.")).toBeInTheDocument();
  });

  it("renders form with organization data (public visibility)", async () => {
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Organization",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(
      /organization name/i
    ) as HTMLInputElement;
    expect(nameInput.value).toBe("Test Organization");

    const publicRadio = screen.getByRole("radio", { name: /public/i });
    expect(publicRadio).toBeChecked();
  });

  it("renders form with organization data (private visibility)", async () => {
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Private Org",
        visibility: Visibility.PRIVATE,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(
      /organization name/i
    ) as HTMLInputElement;
    expect(nameInput.value).toBe("Private Org");

    const privateRadio = screen.getByRole("radio", { name: /private/i });
    expect(privateRadio).toBeChecked();
  });

  it("allows editing organization name", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Original Name",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(/organization name/i);
    await user.clear(nameInput);
    await user.type(nameInput, "Updated Name");

    expect(nameInput).toHaveValue("Updated Name");
  });

  it("allows changing visibility from public to private", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(
        screen.getByRole("radio", { name: /public/i })
      ).toBeInTheDocument();
    });

    const privateRadio = screen.getByRole("radio", { name: /private/i });
    await user.click(privateRadio);

    expect(privateRadio).toBeChecked();
  });

  it("allows changing visibility from private to public", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PRIVATE,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(
        screen.getByRole("radio", { name: /private/i })
      ).toBeInTheDocument();
    });

    const publicRadio = screen.getByRole("radio", { name: /public/i });
    await user.click(publicRadio);

    expect(publicRadio).toBeChecked();
  });

  it("successfully updates organization", async () => {
    const user = userEvent.setup();
    mockUpdateOrganization.mockResolvedValue(undefined);

    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Original Name",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(/organization name/i);
    await user.clear(nameInput);
    await user.type(nameInput, "Updated Organization Name");

    const privateRadio = screen.getByRole("radio", { name: /private/i });
    await user.click(privateRadio);

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockUpdateOrganization).toHaveBeenCalledWith({
        id: "org-123",
        name: "Updated Organization Name",
        visibility: Visibility.PRIVATE,
      });
    });

    await waitFor(() => {
      expect(toastSuccess).toHaveBeenCalledWith(
        "Organization updated successfully."
      );
    });
  });

  it("shows validation error for empty organization name", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(/organization name/i);
    await user.clear(nameInput);

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText(/please enter an organization name/i)
      ).toBeInTheDocument();
    });

    expect(mockUpdateOrganization).not.toHaveBeenCalled();
  });

  it("shows validation error for organization name exceeding 100 characters", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(/organization name/i);
    await user.clear(nameInput);
    await user.type(nameInput, "a".repeat(101));

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText(/organization name must be at most 100 characters/i)
      ).toBeInTheDocument();
    });

    expect(mockUpdateOrganization).not.toHaveBeenCalled();
  });

  it("handles PermissionDenied error", async () => {
    const user = userEvent.setup();
    const error = new ConnectError("Permission denied", Code.PermissionDenied);
    mockUpdateOrganization.mockRejectedValue(error);

    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(
        "You don't have permission to update this organization."
      );
    });
  });

  it("handles NotFound error", async () => {
    const user = userEvent.setup();
    const error = new ConnectError("Not found", Code.NotFound);
    mockUpdateOrganization.mockRejectedValue(error);

    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith("Organization not found.");
    });
  });

  it("handles generic error", async () => {
    const user = userEvent.setup();
    const error = new Error("Network error");
    mockUpdateOrganization.mockRejectedValue(error);

    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const submitButton = screen.getByRole("button", { name: /save changes/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(
        "Failed to update organization. Please try again."
      );
    });
  });

  it("shows error toast when fetching organization fails", () => {
    const error = new Error("Failed to fetch");
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
    });

    render(<OrganizationSettingsForm />);

    expect(toastError).toHaveBeenCalledWith(
      "Error occurred while fetching organization"
    );
  });

  it("does not show error toast for NotFound error when fetching", () => {
    const error = new ConnectError("Not found", Code.NotFound);
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
    });

    render(<OrganizationSettingsForm />);

    expect(toastError).not.toHaveBeenCalled();
  });

  it("disables form fields while submitting", async () => {
    const user = userEvent.setup();
    let resolveUpdate: () => void;
    const updatePromise = new Promise<void>((resolve) => {
      resolveUpdate = resolve;
    });
    mockUpdateOrganization.mockReturnValue(updatePromise);

    mockUseQuery.mockReturnValue({
      data: {
        id: "org-123",
        name: "Test Org",
        visibility: Visibility.PUBLIC,
      },
      isLoading: false,
      error: null,
    });

    render(<OrganizationSettingsForm />);

    await waitFor(() => {
      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
    });

    const nameInput = screen.getByLabelText(/organization name/i);
    const submitButton = screen.getByRole("button", { name: /save changes/i });

    await user.click(submitButton);

    expect(submitButton).toBeDisabled();

    resolveUpdate!();
    await waitFor(() => {
      expect(mockUpdateOrganization).toHaveBeenCalled();
    });
  });

  describe("Delete Organization", () => {
    const renderFormWithOrganization = () => {
      mockUseQuery.mockReturnValue({
        data: {
          id: "org-123",
          name: "Test Organization",
          visibility: Visibility.PUBLIC,
        },
        isLoading: false,
        error: null,
      });
      return render(<OrganizationSettingsForm />);
    };

    it("renders Danger Zone section with delete button", async () => {
      renderFormWithOrganization();

      await waitFor(() => {
        expect(screen.getByText("Danger Zone")).toBeInTheDocument();
      });

      expect(
        screen.getByText("Irreversible and destructive actions")
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /delete organization/i })
      ).toBeInTheDocument();
    });

    it("opens delete dialog when delete button is clicked", async () => {
      const user = userEvent.setup();
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      expect(
        screen.getByText(/test organization/i, { exact: false })
      ).toBeInTheDocument();
    });

    it("successfully deletes organization", async () => {
      const user = userEvent.setup();
      mockDeleteOrganization.mockResolvedValue(undefined);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockDeleteOrganization).toHaveBeenCalledWith({
          id: "org-123",
        });
      });

      await waitFor(() => {
        expect(toastSuccess).toHaveBeenCalledWith(
          "Organization deleted successfully."
        );
      });

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith("/dashboard");
      });
    });

    it("handles PermissionDenied error when deleting", async () => {
      const user = userEvent.setup();
      const error = new ConnectError(
        "Permission denied",
        Code.PermissionDenied
      );
      mockDeleteOrganization.mockRejectedValue(error);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(toastError).toHaveBeenCalledWith(
          "You don't have permission to delete this organization."
        );
      });

      expect(mockPush).not.toHaveBeenCalled();
    });

    it("handles NotFound error when deleting", async () => {
      const user = userEvent.setup();
      const error = new ConnectError("Not found", Code.NotFound);
      mockDeleteOrganization.mockRejectedValue(error);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(toastError).toHaveBeenCalledWith("Organization not found.");
      });

      expect(mockPush).not.toHaveBeenCalled();
    });

    it("handles generic error when deleting", async () => {
      const user = userEvent.setup();
      const error = new Error("Network error");
      mockDeleteOrganization.mockRejectedValue(error);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(toastError).toHaveBeenCalledWith(
          "Failed to delete organization. Please try again."
        );
      });

      expect(mockPush).not.toHaveBeenCalled();
    });

    it("disables delete button while deleting", async () => {
      const user = userEvent.setup();
      let resolveDelete: () => void;
      const deletePromise = new Promise<void>((resolve) => {
        resolveDelete = resolve;
      });
      mockDeleteOrganization.mockReturnValue(deletePromise);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(confirmButton);

      await waitFor(() => {
        const buttons = screen.getAllByRole("button", {
          name: /delete organization/i,
        });

        expect(buttons[buttons.length - 1]).toBeDisabled();
      });

      resolveDelete!();
      await waitFor(() => {
        expect(mockDeleteOrganization).toHaveBeenCalled();
      });
    });

    it("prevents dialog from closing while deleting", async () => {
      const user = userEvent.setup();
      let resolveDelete: () => void;
      const deletePromise = new Promise<void>((resolve) => {
        resolveDelete = resolve;
      });
      mockDeleteOrganization.mockReturnValue(deletePromise);
      renderFormWithOrganization();

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const deleteButton = screen.getByRole("button", {
        name: /delete organization/i,
      });
      await user.click(deleteButton);

      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: /delete organization/i })
        ).toBeInTheDocument();
      });

      const allDeleteButtons = screen.getAllByRole("button", {
        name: /delete organization/i,
      });
      const dialogConfirmButton = allDeleteButtons[allDeleteButtons.length - 1];
      await user.click(dialogConfirmButton);

      await waitFor(() => {
        expect(mockDeleteOrganization).toHaveBeenCalled();
      });

      expect(
        screen.getByRole("heading", { name: /delete organization/i })
      ).toBeInTheDocument();

      resolveDelete!();
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith("/dashboard");
      });
    });

    it("does not delete when organization is not loaded", () => {
      mockUseQuery.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      });

      render(<OrganizationSettingsForm />);

      expect(
        screen.queryByRole("button", { name: /delete organization/i })
      ).not.toBeInTheDocument();

      expect(mockDeleteOrganization).not.toHaveBeenCalled();
    });
  });
});
