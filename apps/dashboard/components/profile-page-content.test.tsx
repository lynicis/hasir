import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import ProfilePageContent from "./profile-page-content";

const mockRefreshSession = vi.fn();
const mockPush = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

vi.mock("@/lib/session-provider", () => ({
  useSession: () => ({
    session: {
      user: {
        id: "123",
        email: "test@example.com",
      },
      accessToken: "mock-token",
    },
    refreshSession: mockRefreshSession,
  }),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateUser: vi.fn(),
    deleteAccount: vi.fn(),
  }),
}));

vi.mock("@/components/profile-form", () => ({
  ProfileForm: ({ onSubmit }: { onSubmit: () => void }) => (
    <div data-testid="profile-form">
      <button onClick={onSubmit}>Submit Profile</button>
    </div>
  ),
}));

vi.mock("@/components/ssh-api-key-panel", () => ({
  SshApiKeyPanel: () => <div data-testid="ssh-api-key-panel">SSH/API Key Panel</div>,
}));

vi.mock("@/components/password-confirmation-dialog", () => ({
  PasswordConfirmationDialog: ({
    open,
    onConfirm,
  }: {
    open: boolean;
    onConfirm: () => void;
  }) => (
    <div data-testid="password-dialog" data-open={open}>
      <button onClick={onConfirm}>Confirm Password</button>
    </div>
  ),
}));

vi.mock("@/components/danger-zone", () => ({
  DangerZone: ({
    onDelete,
    isDeleting,
  }: {
    onDelete: () => void;
    isDeleting: boolean;
  }) => (
    <div data-testid="danger-zone" data-deleting={isDeleting}>
      <button onClick={onDelete}>Delete Account</button>
    </div>
  ),
}));

describe("ProfilePageContent", () => {
  it("renders the profile tabs", () => {
    render(<ProfilePageContent />);

    expect(screen.getByRole("tab", { name: /profile/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /ssh.*api key/i })).toBeInTheDocument();
  });

  it("renders the profile form by default", () => {
    render(<ProfilePageContent />);

    expect(screen.getByTestId("profile-form")).toBeInTheDocument();
  });

  it("renders the danger zone", () => {
    render(<ProfilePageContent />);

    expect(screen.getByTestId("danger-zone")).toBeInTheDocument();
  });

  it("renders the password confirmation dialog", () => {
    render(<ProfilePageContent />);

    expect(screen.getByTestId("password-dialog")).toBeInTheDocument();
  });
});
