import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { type OrganizationRepository } from "./repository-item";
import { RepositoriesList } from "./repositories-list";

describe("RepositoriesList", () => {
  const mockRepositories: OrganizationRepository[] = [
    {
      id: "1",
      name: "frontend-app",
      visibility: "public",
    },
    {
      id: "2",
      name: "backend-api",
      visibility: "private",
    },
  ];

  const mockOnDelete = vi.fn();
  const mockOnCreate = vi.fn();
  const mockOnPageChange = vi.fn();

  const defaultPaginationProps = {
    currentPage: 1,
    totalPages: 1,
    totalRepositories: 2,
    hasNextPage: false,
    hasPreviousPage: false,
    onPageChange: mockOnPageChange,
  };

  it("renders repositories list with correct information", () => {
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    expect(screen.getByText("Repositories")).toBeInTheDocument();
    expect(screen.getByText("frontend-app")).toBeInTheDocument();
    expect(screen.getByText("backend-api")).toBeInTheDocument();
  });

  it("renders create repository button", () => {
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    expect(
      screen.getByRole("button", { name: /create repository/i })
    ).toBeInTheDocument();
  });

  it("calls onCreate when create button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    const createButton = screen.getByRole("button", {
      name: /create repository/i,
    });
    await user.click(createButton);

    expect(mockOnCreate).toHaveBeenCalledTimes(1);
  });

  it("renders empty state when no repositories", () => {
    render(
      <RepositoriesList
        repositories={[]}
        {...defaultPaginationProps}
        totalRepositories={0}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    expect(screen.getByText("Repositories")).toBeInTheDocument();
    expect(screen.queryByText("frontend-app")).not.toBeInTheDocument();
  });

  it("renders all repositories in the list", () => {
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    expect(screen.getByText("frontend-app")).toBeInTheDocument();
    expect(screen.getByText("backend-api")).toBeInTheDocument();
  });
});
