import { render, screen } from "@testing-library/react";

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

describe("ProfilePageContent", () => {
  it("renders the profile tabs", () => {
    render(<ProfilePageContent />);

    expect(screen.getByRole("tab", { name: /profile/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /ssh.*api key/i })).toBeInTheDocument();
  });

  it("renders the profile form by default", () => {
    render(<ProfilePageContent />);

    expect(screen.getByText("Profile Settings")).toBeInTheDocument();
  });

  it("renders the danger zone", () => {
    render(<ProfilePageContent />);

    expect(screen.getByText("Danger Zone")).toBeInTheDocument();
  });
});
