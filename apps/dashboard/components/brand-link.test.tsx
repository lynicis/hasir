import type { ComponentProps, ReactNode } from "react";

import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { BrandLink } from "./brand-link";

vi.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    ...props
  }: ComponentProps<"a"> & { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}));

vi.mock("next/image", () => ({
  __esModule: true,
  default: ({ src, alt, ...props }: ComponentProps<"img">) => (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={src} alt={alt} {...props} />
  ),
}));

describe("BrandLink", () => {
  it("renders the default label and href", () => {
    render(<BrandLink />);

    const link = screen.getByRole("link", {
      name: "Hasir LogoHasir Proto Schema Registry",
    });

    expect(link).toHaveAttribute("href", "/");
  });

  it("renders with a custom label and href", () => {
    render(<BrandLink label="Docs" href="/docs" />);

    const link = screen.getByRole("link", { name: "Hasir LogoDocs" });

    expect(link).toHaveAttribute("href", "/docs");
  });

  it("merges custom class names with defaults", () => {
    render(<BrandLink className="text-muted" />);

    const link = screen.getByRole("link", {
      name: "Hasir LogoHasir Proto Schema Registry",
    });

    expect(link.className).toContain("flex");
    expect(link.className).toContain("text-muted");
  });

  it("renders the logo image with correct attributes", () => {
    render(<BrandLink />);

    const image = screen.getByAltText("Hasir Logo");

    expect(image).toHaveAttribute("src", "/logo.webp");
    expect(image).toHaveAttribute("width", "24");
    expect(image).toHaveAttribute("height", "24");
    expect(image).toHaveClass("size-6");
  });
});
