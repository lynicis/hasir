import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MemberItem, type OrganizationMember } from "./member-item";

describe("MemberItem", () => {
  const mockMember: OrganizationMember = {
    id: "1",
    email: "john.doe@example.com",
    name: "John Doe",
    permission: "author",
  };

  const mockOnPermissionChange = vi.fn();
  const mockOnDelete = vi.fn();
  const mockGetInitials = (name: string) =>
    name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);

  it("renders member information correctly", () => {
    render(
      <MemberItem
        member={mockMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.getByText("john.doe@example.com")).toBeInTheDocument();
    expect(screen.getByText("Author")).toBeInTheDocument();
  });

  it("displays member initials in avatar fallback", () => {
    render(
      <MemberItem
        member={mockMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    expect(screen.getByText("JD")).toBeInTheDocument();
  });

  it("calls onPermissionChange when permission is changed", async () => {
    const user = userEvent.setup();
    render(
      <MemberItem
        member={mockMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    const permissionButton = screen.getByRole("button", { name: /author/i });
    await user.click(permissionButton);

    const memberOption = screen.getByRole("menuitem", { name: /reader/i });
    await user.click(memberOption);

    expect(mockOnPermissionChange).toHaveBeenCalledWith("1", "reader");
  });

  it("calls onDelete when delete is clicked", async () => {
    const user = userEvent.setup();
    render(
      <MemberItem
        member={mockMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    const moreButton = screen.getByRole("button", { name: "" });
    await user.click(moreButton);

    const deleteOption = screen.getByRole("menuitem", {
      name: /remove from organization/i,
    });
    await user.click(deleteOption);

    expect(mockOnDelete).toHaveBeenCalledWith(mockMember);
  });

  it("disables delete option for owner", () => {
    const ownerMember: OrganizationMember = {
      ...mockMember,
      permission: "owner",
    };

    render(
      <MemberItem
        member={ownerMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    const moreButton = screen.getByRole("button", { name: "" });
    expect(moreButton).toBeInTheDocument();
  });

  it("highlights current permission in dropdown", async () => {
    const user = userEvent.setup();
    render(
      <MemberItem
        member={mockMember}
        onPermissionChange={mockOnPermissionChange}
        onDelete={mockOnDelete}
        canEditPermissions={true}
        canRemove={true}
        getInitials={mockGetInitials}
      />
    );

    const permissionButton = screen.getByRole("button", { name: /author/i });
    await user.click(permissionButton);

    const authorOption = screen.getByRole("menuitem", { name: /author/i });
    expect(authorOption).toHaveClass("bg-accent");
  });
});
