import type { Mock } from "vitest";

import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { OrganizationDialogForm } from "./organization-dialog-form";

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

const mockCreateOrganization = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

describe("OrganizationDialogForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      createOrganization: mockCreateOrganization,
    } as never);
    mockCreateOrganization.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  function setup(open = true) {
    const onOpenChange = vi.fn();
    const onCancel = vi.fn();

    render(
      <OrganizationDialogForm
        open={open}
        onOpenChange={onOpenChange}
        onCancel={onCancel}
      />
    );

    return { onOpenChange, onCancel };
  }

  it("renders step 1 with name and visibility fields when open", () => {
    setup(true);

    expect(
      screen.getByRole("heading", { name: /create organization/i })
    ).toBeInTheDocument();

    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /public/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /private/i })).toBeInTheDocument();
  });

  it("shows validation error when trying to proceed without name", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByText(/please enter an organization name/i)
      ).toBeInTheDocument()
    );
  });

  it("navigates to step 2 when Next is clicked with valid data", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.type(screen.getByLabelText(/name/i), "My Organization");
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /invite your friends/i })
      ).toBeInTheDocument()
    );
  });

  it("can add and remove email invites on step 2", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.type(screen.getByLabelText(/name/i), "My Organization");
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(screen.getByText(/no invites yet/i)).toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /add email/i }));

    await waitFor(() =>
      expect(
        screen.getByPlaceholderText(/friend@example.com/i)
      ).toBeInTheDocument()
    );

    await user.type(
      screen.getByPlaceholderText(/friend@example.com/i),
      "test@example.com"
    );

    const removeButtons = screen.getAllByRole("button");
    const removeButton = removeButtons.find((btn) =>
      btn.querySelector("svg.lucide-x")
    );

    if (removeButton) {
      await user.click(removeButton);
    }

    await waitFor(() =>
      expect(screen.getByText(/no invites yet/i)).toBeInTheDocument()
    );
  });

  it("can navigate back from step 2 to step 1", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.type(screen.getByLabelText(/name/i), "My Organization");
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /invite your friends/i })
      ).toBeInTheDocument()
    );

    const backButton = screen.getByRole("button", { name: /go back/i });
    await user.click(backButton);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /create organization/i })
      ).toBeInTheDocument()
    );
  });

  it("creates organization with invites when submitted", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup(true);

    await user.type(screen.getByLabelText(/name/i), "Test Org");
    await user.click(screen.getByRole("radio", { name: /private/i }));
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /invite your friends/i })
      ).toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /add email/i }));
    await user.type(
      screen.getByPlaceholderText(/friend@example.com/i),
      "friend@test.com"
    );

    await user.click(screen.getByRole("button", { name: /create & invite/i }));

    await waitFor(() =>
      expect(mockCreateOrganization).toHaveBeenCalledWith({
        name: "Test Org",
        visibility: Visibility.PRIVATE,
        inviteEmails: ["friend@test.com"],
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith(
        "Organization created successfully."
      )
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("creates organization without invites when no emails added", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup(true);

    await user.type(screen.getByLabelText(/name/i), "Solo Org");
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /invite your friends/i })
      ).toBeInTheDocument()
    );

    expect(
      screen.getByRole("button", { name: /^create$/i })
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^create$/i }));

    await waitFor(() =>
      expect(mockCreateOrganization).toHaveBeenCalledWith({
        name: "Solo Org",
        visibility: Visibility.PUBLIC,
        inviteEmails: [],
      })
    );

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("calls onCancel and resets form when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const { onCancel } = setup(true);

    const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement;
    await user.type(nameInput, "temp-name");

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
    expect(nameInput.value).toBe("");
  });

  it("prevents submission with empty email field", async () => {
    const user = userEvent.setup();
    setup(true);

    await user.type(screen.getByLabelText(/name/i), "My Organization");
    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: /invite your friends/i })
      ).toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /add email/i }));

    await user.click(screen.getByRole("button", { name: /create & invite/i }));

    await waitFor(() => {
      expect(mockCreateOrganization).not.toHaveBeenCalled();
    });
  });
});
