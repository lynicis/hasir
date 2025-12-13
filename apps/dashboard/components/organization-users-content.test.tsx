import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import UsersPage from "./organization-users-content";

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

const mockUseQuery = vi.fn();
const mockRefetch = vi.fn();
vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

const mockUseSession = vi.fn();
vi.mock("@/lib/session-provider", () => ({
  useSession: () => mockUseSession(),
}));

const mockUpdateMemberRole = vi.fn();
const mockDeleteMember = vi.fn();
const mockUseClient = vi.fn();
vi.mock("@/lib/use-client", () => ({
  useClient: () => mockUseClient(),
}));

vi.mock("next/navigation", () => ({
  useParams: vi.fn(() => ({ id: "org-123" })),
}));

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccess,
    error: toastError,
  },
}));

describe("UsersPage", () => {
  beforeEach(() => {
    mockUseQuery.mockReset();
    mockUseSession.mockReset();
    mockUseClient.mockReset();
    mockUpdateMemberRole.mockReset();
    mockDeleteMember.mockReset();
    mockRefetch.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();

    mockUseClient.mockReturnValue({
      updateMemberRole: mockUpdateMemberRole,
      deleteMember: mockDeleteMember,
    });

    mockRefetch.mockResolvedValue(undefined);
  });

  it("allows owners to invite, change roles and remove members", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "author",
            email: "author@example.com",
            role: 2,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersPage />);

    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });
    expect(inviteButton).toBeEnabled();

    await user.click(inviteButton);
  });

  it("allows authors to invite but not change roles or remove members", async () => {
    mockUseSession.mockReturnValue({
      session: { user: { email: "author@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "author",
            email: "author@example.com",
            role: 2,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersPage />);

    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });

    expect(inviteButton).toBeEnabled();
  });

  it("prevents readers from inviting, changing roles or removing members", async () => {
    mockUseSession.mockReturnValue({
      session: { user: { email: "reader@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "reader",
            email: "reader@example.com",
            role: 1,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersPage />);

    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });

    expect(inviteButton).toBeDisabled();
  });

  it("successfully updates member role when owner changes permission", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "member",
            email: "member@example.com",
            role: 1,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    mockUpdateMemberRole.mockResolvedValue({});

    render(<UsersPage />);

    const readerButtons = await screen.findAllByRole("button", {
      name: /reader/i,
    });
    expect(readerButtons.length).toBeGreaterThan(0);

    await user.click(readerButtons[0]);

    const authorOption = await screen.findByRole("menuitem", {
      name: /author/i,
    });
    await user.click(authorOption);

    expect(mockUpdateMemberRole).toHaveBeenCalledWith({
      organizationId: "org-123",
      memberId: "2",
      role: 2,
    });

    expect(toastSuccess).toHaveBeenCalledWith(
      "Permission updated successfully"
    );
    expect(mockRefetch).toHaveBeenCalled();
  });

  it("shows error when updating member role fails", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "member",
            email: "member@example.com",
            role: 1,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    const error = new Error("Network error");
    mockUpdateMemberRole.mockRejectedValue(error);

    render(<UsersPage />);

    // Find and click the "Reader" button to open the dropdown
    const readerButtons = await screen.findAllByRole("button", {
      name: /reader/i,
    });
    await user.click(readerButtons[0]);

    // Find and click the "Author" option in the dropdown menu
    const authorOption = await screen.findByRole("menuitem", {
      name: /author/i,
    });
    await user.click(authorOption);

    expect(toastError).toHaveBeenCalledWith(
      "Failed to update permission. Please try again."
    );
    expect(mockRefetch).not.toHaveBeenCalled();
  });

  it("successfully deletes a member when owner confirms deletion", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "member",
            email: "member@example.com",
            role: 1,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    mockDeleteMember.mockResolvedValue({});

    render(<UsersPage />);

    const menuButtons = screen.getAllByRole("button", { name: "" });
    const memberMenuButton = menuButtons[menuButtons.length - 1];
    await user.click(memberMenuButton);

    const removeMenuItem = await screen.findByRole("menuitem", {
      name: /remove from organization/i,
    });
    await user.click(removeMenuItem);

    const confirmButton = await screen.findByRole("button", {
      name: /remove member/i,
    });
    await user.click(confirmButton);

    expect(mockDeleteMember).toHaveBeenCalledWith({
      organizationId: "org-123",
      memberId: "2",
    });

    expect(toastSuccess).toHaveBeenCalledWith(
      "member has been removed from the organization"
    );
    expect(mockRefetch).toHaveBeenCalled();
  });

  it("shows error when deleting member fails", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          {
            id: "1",
            username: "owner",
            email: "owner@example.com",
            role: 3,
          },
          {
            id: "2",
            username: "member",
            email: "member@example.com",
            role: 1,
          },
        ],
      },
      error: null,
      refetch: mockRefetch,
    });

    const error = new Error("Network error");
    mockDeleteMember.mockRejectedValue(error);

    render(<UsersPage />);

    const menuButtons = screen.getAllByRole("button", { name: "" });
    const memberMenuButton = menuButtons[menuButtons.length - 1];
    await user.click(memberMenuButton);

    const removeMenuItem = await screen.findByRole("menuitem", {
      name: /remove from organization/i,
    });
    await user.click(removeMenuItem);

    const confirmButton = await screen.findByRole("button", {
      name: /remove member/i,
    });
    await user.click(confirmButton);

    expect(toastError).toHaveBeenCalledWith(
      "Failed to remove member. Please try again."
    );
    expect(mockRefetch).not.toHaveBeenCalled();
  });
});
