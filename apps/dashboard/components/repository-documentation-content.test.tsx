import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import RepositoryDocumentationContent from "./repository-documentation-content";

describe("RepositoryDocumentationContent", () => {
  it("renders the Documentation card title", () => {
    render(<RepositoryDocumentationContent />);

    expect(screen.getByText("Documentation")).toBeInTheDocument();
  });

  it("renders the card description", () => {
    render(<RepositoryDocumentationContent />);

    expect(
      screen.getByText(/view and manage repository documentation/i)
    ).toBeInTheDocument();
  });

  it("displays the placeholder message", () => {
    render(<RepositoryDocumentationContent />);

    expect(
      screen.getByText(/repository documentation will be displayed here/i)
    ).toBeInTheDocument();
  });

  it("lists the planned documentation features", () => {
    render(<RepositoryDocumentationContent />);

    expect(screen.getByText(/this section will include:/i)).toBeInTheDocument();
    expect(
      screen.getByText(/readme and markdown documentation/i)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/api reference documentation/i)
    ).toBeInTheDocument();
    expect(screen.getByText(/usage examples and guides/i)).toBeInTheDocument();
    expect(screen.getByText(/schema documentation/i)).toBeInTheDocument();
  });
});
