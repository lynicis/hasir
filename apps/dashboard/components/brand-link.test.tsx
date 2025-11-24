import type { ComponentProps, ReactNode } from "react";

import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { BrandLink } from "./brand-link";

vi.mock("next/link", () => ({
  __esModule: true,
  default: ({ children, ...props }: ComponentProps<"a"> & { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}));

describe("BrandLink", () => {
  it("renders the default label and href", () => {
    render(<BrandLink />);

    const link = screen.getByRole("link", {
      name: "Hasir Proto Schema Registry",
    });

    expect(link).toHaveAttribute("href", "/");
  });

  it("renders with a custom label and href", () => {
    render(<BrandLink label="Docs" href="/docs" />);

    const link = screen.getByRole("link", { name: "Docs" });

    expect(link).toHaveAttribute("href", "/docs");
  });

  it("merges custom class names with defaults", () => {
    render(<BrandLink className="text-muted" />);

    const link = screen.getByRole("link", {
      name: "Hasir Proto Schema Registry",
    });

    expect(link.className).toContain("flex");
    expect(link.className).toContain("text-muted");
  });
});


