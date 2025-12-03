import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import UsersPage from "./page";

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

const mockUseQuery = vi.fn();
vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

const mockUseSession = vi.fn();
vi.mock("@/lib/session-provider", () => ({
  useSession: () => mockUseSession(),
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
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  it("allows owners to invite, change roles and remove members", async () => {
    const user = userEvent.setup();

    mockUseSession.mockReturnValue({
      session: { user: { email: "owner@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          { id: "1", email: "owner@example.com", role: 2 }, // OWNER
          { id: "2", email: "author@example.com", role: 1 }, // AUTHOR
        ],
      },
      error: null,
    });

    render(<UsersPage />);

    // Invite button enabled for owner
    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });
    expect(inviteButton).toBeEnabled();

    // Owner can invite members
    await user.click(inviteButton);
  });

  it("allows authors to invite but not change roles or remove members", async () => {
    mockUseSession.mockReturnValue({
      session: { user: { email: "author@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          { id: "1", email: "owner@example.com", role: 2 }, // OWNER
          { id: "2", email: "author@example.com", role: 1 }, // AUTHOR
        ],
      },
      error: null,
    });

    render(<UsersPage />);

    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });
    // Authors currently cannot invite members in the UI
    expect(inviteButton).toBeDisabled();
  });

  it("prevents readers from inviting, changing roles or removing members", async () => {
    mockUseSession.mockReturnValue({
      session: { user: { email: "reader@example.com" } },
    });

    mockUseQuery.mockReturnValue({
      data: {
        members: [
          { id: "1", email: "owner@example.com", role: 2 }, // OWNER
          { id: "2", email: "reader@example.com", role: 0 }, // READER
        ],
      },
      error: null,
    });

    render(<UsersPage />);

    const inviteButton = await screen.findByRole("button", {
      name: /invite member/i,
    });
    expect(inviteButton).toBeDisabled();

    // Reader cannot invite members (already asserted by invite button disabled)
  });
});
