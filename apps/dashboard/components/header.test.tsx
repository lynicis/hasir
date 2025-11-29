import type { ComponentProps, ReactNode } from "react";

import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Header } from "./header";

vi.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    ...props
  }: ComponentProps<"a"> & { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}));

vi.mock("@/lib/session-provider", () => ({
  useSession: () => ({
    session: {
      user: { id: "user-123", email: "test@example.com" },
    },
    loading: false,
    refreshSession: vi.fn(),
  }),
  SessionProvider: ({ children }: { children: ReactNode }) => children,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    refresh: vi.fn(),
  }),
  usePathname: () => "/",
}));

describe("Header", () => {
  it("renders brand link with correct label and href", () => {
    render(<Header />);

    const brandLink = screen.getByRole("link", { name: "Hasir" });

    expect(brandLink).toBeInTheDocument();
    expect(brandLink).toHaveAttribute("href", "/dashboard");
  });

  it("renders search input with shortcut hint", () => {
    render(<Header />);

    const searchInput = screen.getByRole("searchbox", { name: /search/i });
    const modifierKey = screen.getByText(/⌘|Ctrl/);
    const keyHint = screen.getByText("K");

    expect(searchInput).toBeInTheDocument();
    expect(modifierKey).toBeInTheDocument();
    expect(keyHint).toBeInTheDocument();
  });

  it("opens the create popover when the create button is clicked", async () => {
    const user = userEvent.setup();

    render(<Header />);

    const createButton = screen.getByRole("button", { name: /create/i });
    await user.click(createButton);

    expect(
      screen.getByRole("button", { name: /create organization/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /create repository/i })
    ).toBeInTheDocument();
  });

  it("opens the create repository dialog with name and visibility fields", async () => {
    const user = userEvent.setup();

    render(<Header />);

    const createButton = screen.getByRole("button", { name: /create/i });
    await user.click(createButton);

    const createRepoItem = screen.getByRole("button", {
      name: /create repository/i,
    });
    await user.click(createRepoItem);

    expect(
      screen.getByRole("heading", { name: /create repository/i })
    ).toBeInTheDocument();

    expect(screen.getByRole("textbox", { name: /name/i })).toBeInTheDocument();

    expect(screen.getByRole("radio", { name: /public/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /private/i })).toBeInTheDocument();
  });

  it("renders profile button with user menu", async () => {
    const user = userEvent.setup();
    render(<Header />);

    // Find the avatar button that opens the user menu
    const avatarButton = screen.getByRole("button", {
      name: /open user menu/i,
    });
    expect(avatarButton).toBeInTheDocument();

    // Click to open the popover
    await user.click(avatarButton);

    // Check that the profile button is in the popover
    expect(
      screen.getByRole("button", { name: /profile/i })
    ).toBeInTheDocument();
  });

  it("focuses the search input when ⌘K is pressed", () => {
    render(<Header />);

    const searchInput = screen.getByRole("searchbox", { name: /search/i });
    expect(searchInput).not.toHaveFocus();

    fireEvent.keyDown(window, { key: "k", metaKey: true });

    expect(searchInput).toHaveFocus();
  });

  it("focuses the search input when Ctrl+K is pressed", () => {
    render(<Header />);

    const searchInput = screen.getByRole("searchbox", { name: /search/i });
    expect(searchInput).not.toHaveFocus();

    fireEvent.keyDown(window, { key: "k", ctrlKey: true });

    expect(searchInput).toHaveFocus();
  });
});
