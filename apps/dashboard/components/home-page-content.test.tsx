import type { ComponentProps, ReactNode } from "react";

import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import HomePageContent from "./home-page-content";

vi.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    ...props
  }: ComponentProps<"a"> & { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}));

vi.mock("framer-motion", () => ({
  motion: {
    div: ({ children, ...props }: ComponentProps<"div"> & { children?: ReactNode }) => (
      <div {...props}>{children}</div>
    ),
    h1: ({ children, ...props }: ComponentProps<"h1"> & { children?: ReactNode }) => (
      <h1 {...props}>{children}</h1>
    ),
    p: ({ children, ...props }: ComponentProps<"p"> & { children?: ReactNode }) => (
      <p {...props}>{children}</p>
    ),
  },
  Variants: {},
}));

const mockUseSession = vi.fn();

vi.mock("@/lib/session-provider", () => ({
  useSession: () => mockUseSession(),
}));

describe("HomePageContent", () => {
  it("renders the main heading and description", () => {
    mockUseSession.mockReturnValue({
      session: null,
    });

    render(<HomePageContent />);

    expect(screen.getByText("Hasir")).toBeInTheDocument();
    expect(
      screen.getByText(/Your modern Proto Schema Registry dashboard/i)
    ).toBeInTheDocument();
  });

  it("renders login and register buttons when user is not logged in", () => {
    mockUseSession.mockReturnValue({
      session: null,
    });

    render(<HomePageContent />);

    expect(screen.getByRole("link", { name: /login/i })).toHaveAttribute(
      "href",
      "/login"
    );
    expect(screen.getByRole("link", { name: /register/i })).toHaveAttribute(
      "href",
      "/register"
    );
  });

  it("renders dashboard and profile buttons when user is logged in", () => {
    mockUseSession.mockReturnValue({
      session: {
        user: {
          id: "123",
          email: "test@example.com",
        },
      },
    });

    render(<HomePageContent />);

    expect(screen.getByRole("link", { name: /dashboard/i })).toHaveAttribute(
      "href",
      "/dashboard"
    );
    expect(screen.getByRole("link", { name: /profile/i })).toHaveAttribute(
      "href",
      "/profile"
    );
  });

  it("displays welcome message with user email when logged in", () => {
    mockUseSession.mockReturnValue({
      session: {
        user: {
          id: "123",
          email: "test@example.com",
        },
      },
    });

    render(<HomePageContent />);

    expect(screen.getByText(/Welcome back/i)).toBeInTheDocument();
    expect(screen.getByText("test@example.com")).toBeInTheDocument();
  });

  it("does not display welcome message when not logged in", () => {
    mockUseSession.mockReturnValue({
      session: null,
    });

    render(<HomePageContent />);

    expect(screen.queryByText(/Welcome back/i)).not.toBeInTheDocument();
  });
});
