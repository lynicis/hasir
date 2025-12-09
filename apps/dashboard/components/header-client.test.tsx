import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { HeaderClient } from "./header-client";

vi.mock("@/components/header", () => ({
  Header: () => <div data-testid="header">Header Component</div>,
}));

vi.mock("next/dynamic", () => ({
  __esModule: true,
  default: () => {
    const MockedHeader = () => <div data-testid="header">Header Component</div>;
    return MockedHeader;
  },
}));

describe("HeaderClient", () => {
  it("renders the Header component", () => {
    render(<HeaderClient />);

    expect(screen.getByTestId("header")).toBeInTheDocument();
    expect(screen.getByText("Header Component")).toBeInTheDocument();
  });
});
