/* eslint-disable */
import { render, screen } from "@testing-library/react";

const MockHeader = () => {
  const React = require("react");
  return React.createElement("div", { "data-testid": "header" }, "Header");
};

// Override next/dynamic to bypass async loading entirely.
// The preload's mock.module factory captures __nextDynamic by reference,
// so mutating .default on the existing object is visible to the factory.
// This override returns a wrapper that renders the mock component directly.
(globalThis as any).__nextDynamic.default =
  () =>
  (props: Record<string, unknown>) => {
    const React = require("react");
    return React.createElement(MockHeader, props);
  };

// Use require() to avoid import hoisting — the override above must run first.
// This loads header-client.tsx which calls dynamic() at module scope.
const { HeaderClient } = require("./header-client");

describe("HeaderClient", () => {
  it("renders the Header component", () => {
    render(<HeaderClient />);
    expect(screen.getByTestId("header")).toBeInTheDocument();
  });
});
