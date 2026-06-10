import type { ComponentProps, ReactNode } from "react";

import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

import { SshConfigurationContent } from "./ssh-configuration-content";

vi.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    ...props
  }: ComponentProps<"a"> & { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}));

beforeEach(() => {
  vi.stubGlobal(
    "IntersectionObserver",
    class {
      observe = vi.fn();
      unobserve = vi.fn();
      disconnect = vi.fn();
    },
  );
});

describe("SshConfigurationContent", () => {
  it("renders the page title", () => {
    render(<SshConfigurationContent />);

    expect(
      screen.getByRole("heading", { name: "SSH Configuration Guide" }),
    ).toBeInTheDocument();
  });

  it("renders all section headings", () => {
    render(<SshConfigurationContent />);

    const sectionHeadings = [
      "Prerequisites",
      "Generate SSH Key Pair",
      "Add SSH Key to ssh-agent",
      "Add Public Key to Hasir",
      "Clone a Repository",
      "Push Changes",
      "Troubleshooting",
    ];

    for (const heading of sectionHeadings) {
      expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
    }
  });

  it("renders the TOC sidebar with navigation items", () => {
    render(<SshConfigurationContent />);

    const tocLabels = [
      "Prerequisites",
      "Generate SSH Key Pair",
      "Add Key to ssh-agent",
      "Add Public Key to Hasir",
      "Clone a Repository",
      "Push Changes",
      "Troubleshooting",
    ];

    for (const label of tocLabels) {
      expect(
        screen.getByRole("button", { name: label }),
      ).toBeInTheDocument();
    }
  });

  it("renders a link to the profile page", () => {
    render(<SshConfigurationContent />);

    const profileLink = screen.getByRole("link", {
      name: /Go to Profile Settings/,
    });

    expect(profileLink).toBeInTheDocument();
    expect(profileLink).toHaveAttribute("href", "/profile");
  });

  it("renders the Hasir brand link", () => {
    render(<SshConfigurationContent />);

    const brandLink = screen.getByRole("link", { name: /Hasir/ });

    expect(brandLink).toBeInTheDocument();
    expect(brandLink).toHaveAttribute("href", "/");
  });

  it("renders the back to dashboard link", () => {
    render(<SshConfigurationContent />);

    const dashboardLink = screen.getByRole("link", {
      name: /Back to Dashboard/,
    });

    expect(dashboardLink).toBeInTheDocument();
    expect(dashboardLink).toHaveAttribute("href", "/dashboard");
  });
});
