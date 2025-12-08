import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import * as useClientModule from "@/lib/use-client";

import { InviteUserDialog } from "./invite-user-dialog";

const mockInviteMember = vi.fn().mockResolvedValue({});

vi.spyOn(useClientModule, "useClient").mockReturnValue({
  inviteMember: mockInviteMember,
} as never);

describe("InviteUserDialog", () => {
  beforeEach(() => {
    mockInviteMember.mockClear();
    mockInviteMember.mockResolvedValue({});
  });

  it("renders dialog when open", () => {
    render(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    expect(screen.getByText("Invite User")).toBeInTheDocument();
    expect(
      screen.getByText("Send an invitation to join this organization via email")
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Email Address")).toBeInTheDocument();
    expect(screen.getByLabelText("Role")).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    render(
      <InviteUserDialog
        open={false}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    expect(screen.queryByText("Invite User")).not.toBeInTheDocument();
  });

  it("validates email input", async () => {
    const user = userEvent.setup();
    render(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Please enter a valid email address.")
      ).toBeInTheDocument();
    });
  });

  it("prevents submission with invalid email format", async () => {
    const user = userEvent.setup();

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "invalid-email");

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => expect(mockInviteMember).not.toHaveBeenCalled());
  });

  it("calls onOpenChange when cancel is clicked", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={onOpenChange}
        organizationId="org-123"
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("successfully sends invitation with valid email and default role", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={onOpenChange}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "test@example.com");

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockInviteMember).toHaveBeenCalledWith({
        id: "org-123",
        email: "test@example.com",
        role: Role.READER,
      });
    });

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("successfully sends invitation with selected role", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={onOpenChange}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "test@example.com");

    const roleSelect = screen.getByLabelText("Role");
    await user.click(roleSelect);

    const authorOption = await screen.findByRole("option", { name: /author/i });
    await user.click(authorOption);

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockInviteMember).toHaveBeenCalledWith({
        id: "org-123",
        email: "test@example.com",
        role: Role.AUTHOR,
      });
    });

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("resets form after successful submission", async () => {
    const user = userEvent.setup();

    const { rerender } = render(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "test@example.com");

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => expect(mockInviteMember).toHaveBeenCalled());

    rerender(
      <InviteUserDialog
        open={false}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    rerender(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    expect(screen.getByLabelText("Email Address")).toHaveValue("");
  });

  it("disables buttons while submitting", async () => {
    const user = userEvent.setup();
    mockInviteMember.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 1000))
    );

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={vi.fn()}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "test@example.com");

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /cancel/i })).toBeDisabled();
    });
  });

  it("prevents closing dialog while submitting", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    mockInviteMember.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 1000))
    );

    render(
      <InviteUserDialog
        open={true}
        onOpenChange={onOpenChange}
        organizationId="org-123"
      />
    );

    const emailInput = screen.getByLabelText("Email Address");
    await user.type(emailInput, "test@example.com");

    const submitButton = screen.getByRole("button", {
      name: /send invitation/i,
    });
    await user.click(submitButton);

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    expect(onOpenChange).not.toHaveBeenCalled();
  });
});
