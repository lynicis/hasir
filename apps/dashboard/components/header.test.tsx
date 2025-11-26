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
    const shortcutHint = screen.getByText("⌘K");

    expect(searchInput).toBeInTheDocument();
    expect(shortcutHint).toBeInTheDocument();
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

  it("renders profile button linking to the profile page", () => {
    render(<Header />);

    const profileLink = screen.getByRole("link", { name: /profile/i });

    expect(profileLink).toBeInTheDocument();
    expect(profileLink).toHaveAttribute("href", "/profile");
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
