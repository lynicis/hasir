import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { type OrganizationMember } from "./member-item";
import { MembersList } from "./members-list";

describe("MembersList", () => {
  const mockMembers: OrganizationMember[] = [
    {
      id: "1",
      email: "john.doe@example.com",
      name: "John Doe",
      permission: "owner",
    },
    {
      id: "2",
      email: "jane.smith@example.com",
      name: "Jane Smith",
      permission: "author",
    },
  ];

  const mockOnPermissionChange = vi.fn();
  const mockOnDelete = vi.fn();
  const mockOnInvite = vi.fn();
  const mockGetInitials = (name: string) =>
    name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);

  it("renders members list with correct information", () => {
    render(
      <MembersList
        members={mockMembers}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        onInvite={mockOnInvite}
        canInvite={true}
        canEditPermissions={true}
        canRemoveMembers={true}
        getInitials={mockGetInitials}
      />
    );

    expect(screen.getByText("Members")).toBeInTheDocument();
    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.getByText("Jane Smith")).toBeInTheDocument();
    expect(screen.getByText("john.doe@example.com")).toBeInTheDocument();
    expect(screen.getByText("jane.smith@example.com")).toBeInTheDocument();
  });

  it("renders invite member button", () => {
    render(
      <MembersList
        members={mockMembers}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        onInvite={mockOnInvite}
        canInvite={true}
        canEditPermissions={true}
        canRemoveMembers={true}
        getInitials={mockGetInitials}
      />
    );

    expect(
      screen.getByRole("button", { name: /invite member/i })
    ).toBeInTheDocument();
  });

  it("calls onInvite when invite button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <MembersList
        members={mockMembers}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        onInvite={mockOnInvite}
        canInvite={true}
        canEditPermissions={true}
        canRemoveMembers={true}
        getInitials={mockGetInitials}
      />
    );

    const inviteButton = screen.getByRole("button", { name: /invite member/i });
    await user.click(inviteButton);

    expect(mockOnInvite).toHaveBeenCalledTimes(1);
  });

  it("renders empty state when no members", () => {
    render(
      <MembersList
        members={[]}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        onInvite={mockOnInvite}
        canInvite={true}
        canEditPermissions={true}
        canRemoveMembers={true}
        getInitials={mockGetInitials}
      />
    );

    expect(screen.getByText("Members")).toBeInTheDocument();
    expect(screen.queryByText("John Doe")).not.toBeInTheDocument();
  });

  it("renders all members in the list", () => {
    render(
      <MembersList
        members={mockMembers}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        onInvite={mockOnInvite}
        canInvite={true}
        canEditPermissions={true}
        canRemoveMembers={true}
        getInitials={mockGetInitials}
      />
    );

    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.getByText("Jane Smith")).toBeInTheDocument();
  });
});
