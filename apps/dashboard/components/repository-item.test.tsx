import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { RepositoryItem, type OrganizationRepository } from "./repository-item";

describe("RepositoryItem", () => {
  const mockRepository: OrganizationRepository = {
    id: "1",
    name: "frontend-app",
    visibility: "public",
  };

  const mockOnDelete = vi.fn();

  it("renders repository information correctly", () => {
    render(
      <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
    );

    expect(screen.getByText("frontend-app")).toBeInTheDocument();
    expect(screen.getByText("public")).toBeInTheDocument();
  });

  it("calls onDelete when delete button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
    );

    const deleteButton = screen.getByRole("button");
    await user.click(deleteButton);

    expect(mockOnDelete).toHaveBeenCalledWith(mockRepository);
  });

  it("applies correct styling for public visibility", () => {
    render(
      <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
    );

    const visibilityBadge = screen.getByText("public");
    expect(visibilityBadge).toHaveClass("bg-green-100");
  });

  it("applies correct styling for private visibility", () => {
    const privateRepo: OrganizationRepository = {
      id: "2",
      name: "backend-api",
      visibility: "private",
    };

    render(<RepositoryItem repository={privateRepo} onDelete={mockOnDelete} />);

    const visibilityBadge = screen.getByText("private");
    expect(visibilityBadge).toHaveClass("bg-gray-100");
  });
});
