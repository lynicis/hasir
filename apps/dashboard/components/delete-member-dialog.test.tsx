import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DeleteMemberDialog } from "./delete-member-dialog";
import { type OrganizationMember } from "./member-item";

describe("DeleteMemberDialog", () => {
  const mockMember: OrganizationMember = {
    id: "1",
    email: "john.doe@example.com",
    name: "John Doe",
    permission: "author",
  };

  const mockOnOpenChange = vi.fn();
  const mockOnConfirm = vi.fn();

  it("renders dialog when open", () => {
    render(
      <DeleteMemberDialog
        open={true}
        member={mockMember}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", {
        name: /remove member from organization/i,
      })
    ).toBeInTheDocument();
    expect(screen.getByText(/john doe/i)).toBeInTheDocument();
    expect(screen.getByText(/john.doe@example.com/i)).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    render(
      <DeleteMemberDialog
        open={false}
        member={mockMember}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.queryByRole("heading", {
        name: /remove member from organization/i,
      })
    ).not.toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <DeleteMemberDialog
        open={true}
        member={mockMember}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const confirmButton = screen.getByRole("button", {
      name: /remove member/i,
    });
    await user.click(confirmButton);

    expect(mockOnConfirm).toHaveBeenCalledTimes(1);
  });

  it("calls onOpenChange when cancel button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <DeleteMemberDialog
        open={true}
        member={mockMember}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });

  it("handles null member gracefully", () => {
    render(
      <DeleteMemberDialog
        open={true}
        member={null}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", {
        name: /remove member from organization/i,
      })
    ).toBeInTheDocument();
  });
});
