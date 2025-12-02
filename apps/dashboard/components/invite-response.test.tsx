import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import * as useClientModule from "@/lib/use-client";

import { InviteResponse } from "./invite-response";

const mockPush = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
}));

const mockIsInvitationValid = vi.fn();
const mockRespondToInvitation = vi.fn();

vi.spyOn(useClientModule, "useClient").mockReturnValue({
  isInvitationValid: mockIsInvitationValid,
  respondToInvitation: mockRespondToInvitation,
} as never);

describe("InviteResponse", () => {
  beforeEach(() => {
    mockIsInvitationValid.mockReset();
    mockRespondToInvitation.mockReset();
    mockPush.mockReset();
  });

  describe("token validation", () => {
    it("shows loading state initially", () => {
      mockIsInvitationValid.mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<InviteResponse invitationToken="test-token" />);

      expect(screen.queryByText(/you're invited/i)).not.toBeInTheDocument();
      expect(
        screen.queryByText(/invalid invitation/i)
      ).not.toBeInTheDocument();
    });

    it("shows invite form when token is valid", async () => {
      mockIsInvitationValid.mockResolvedValue({});

      render(<InviteResponse invitationToken="valid-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );
      expect(
        screen.getByRole("button", { name: /accept invitation/i })
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /decline/i })
      ).toBeInTheDocument();
    });

    it("shows invalid message when token is not found", async () => {
      mockIsInvitationValid.mockRejectedValue(
        new ConnectError("Not found", Code.NotFound)
      );

      render(<InviteResponse invitationToken="invalid-token" />);

      await waitFor(() =>
        expect(screen.getByText(/invalid invitation/i)).toBeInTheDocument()
      );
      expect(
        screen.getByText(/invalid, has expired, or has already been used/i)
      ).toBeInTheDocument();
    });

    it("shows invalid message when permission denied", async () => {
      mockIsInvitationValid.mockRejectedValue(
        new ConnectError("Permission denied", Code.PermissionDenied)
      );

      render(<InviteResponse invitationToken="forbidden-token" />);

      await waitFor(() =>
        expect(screen.getByText(/invalid invitation/i)).toBeInTheDocument()
      );
    });

    it("shows error state on unexpected error", async () => {
      mockIsInvitationValid.mockRejectedValue(
        new ConnectError("Internal error", Code.Internal)
      );

      render(<InviteResponse invitationToken="error-token" />);

      await waitFor(() =>
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      );
    });

    it("shows invalid message when token is empty", async () => {
      render(<InviteResponse invitationToken="" />);

      await waitFor(() =>
        expect(screen.getByText(/invalid invitation/i)).toBeInTheDocument()
      );
      expect(mockIsInvitationValid).not.toHaveBeenCalled();
    });

    it("calls isInvitationValid with correct token", async () => {
      mockIsInvitationValid.mockResolvedValue({});

      render(<InviteResponse invitationToken="my-test-token" />);

      await waitFor(() =>
        expect(mockIsInvitationValid).toHaveBeenCalledWith({
          token: "my-test-token",
        })
      );
    });
  });

  describe("accepting invitation", () => {
    beforeEach(() => {
      mockIsInvitationValid.mockResolvedValue({});
    });

    it("calls respondToInvitation with status=true", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockResolvedValue({});

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(
        screen.getByRole("button", { name: /accept invitation/i })
      );

      await waitFor(() =>
        expect(mockRespondToInvitation).toHaveBeenCalledWith({
          invitationId: "test-token",
          status: true,
        })
      );
    });

    it("shows success message after accepting", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockResolvedValue({});

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(
        screen.getByRole("button", { name: /accept invitation/i })
      );

      await waitFor(() =>
        expect(screen.getByText(/welcome to the team/i)).toBeInTheDocument()
      );
    });

    it("disables decline button while accepting", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 1000))
      );

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(
        screen.getByRole("button", { name: /accept invitation/i })
      );

      await waitFor(() =>
        expect(screen.getByRole("button", { name: /decline/i })).toBeDisabled()
      );
    });
  });

  describe("declining invitation", () => {
    beforeEach(() => {
      mockIsInvitationValid.mockResolvedValue({});
    });

    it("calls respondToInvitation with status=false", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockResolvedValue({});

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /decline/i }));

      await waitFor(() =>
        expect(mockRespondToInvitation).toHaveBeenCalledWith({
          invitationId: "test-token",
          status: false,
        })
      );
    });

    it("shows declined message after declining", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockResolvedValue({});

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /decline/i }));

      await waitFor(() =>
        expect(screen.getByText(/invitation declined/i)).toBeInTheDocument()
      );
    });

    it("disables accept button while declining", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 1000))
      );

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /decline/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /accept invitation/i })
        ).toBeDisabled()
      );
    });
  });

  describe("error handling", () => {
    beforeEach(() => {
      mockIsInvitationValid.mockResolvedValue({});
    });

    it("shows error when invitation not found during response", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockRejectedValue(
        new ConnectError("Not found", Code.NotFound)
      );

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(
        screen.getByRole("button", { name: /accept invitation/i })
      );

      await waitFor(() =>
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      );
    });

    it("shows error when already a member", async () => {
      const user = userEvent.setup();
      mockRespondToInvitation.mockRejectedValue(
        new ConnectError("Already exists", Code.AlreadyExists)
      );

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(
        screen.getByRole("button", { name: /accept invitation/i })
      );

      await waitFor(() =>
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      );
    });
  });

  describe("navigation", () => {
    it("navigates to login from invalid invitation page", async () => {
      const user = userEvent.setup();
      mockIsInvitationValid.mockRejectedValue(
        new ConnectError("Not found", Code.NotFound)
      );

      render(<InviteResponse invitationToken="invalid-token" />);

      await waitFor(() =>
        expect(screen.getByText(/invalid invitation/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /go to login/i }));

      expect(mockPush).toHaveBeenCalledWith("/login");
    });

    it("navigates to login from declined page", async () => {
      const user = userEvent.setup();
      mockIsInvitationValid.mockResolvedValue({});
      mockRespondToInvitation.mockResolvedValue({});

      render(<InviteResponse invitationToken="test-token" />);

      await waitFor(() =>
        expect(screen.getByText(/you're invited/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /decline/i }));

      await waitFor(() =>
        expect(screen.getByText(/invitation declined/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /go to login/i }));

      expect(mockPush).toHaveBeenCalledWith("/login");
    });

    it("navigates to login from error page", async () => {
      const user = userEvent.setup();
      mockIsInvitationValid.mockRejectedValue(
        new ConnectError("Internal error", Code.Internal)
      );

      render(<InviteResponse invitationToken="error-token" />);

      await waitFor(() =>
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      );

      await user.click(screen.getByRole("button", { name: /go to login/i }));

      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });
});
