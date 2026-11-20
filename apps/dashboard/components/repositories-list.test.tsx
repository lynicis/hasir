import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

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

  it("does not render pagination for a single page", () => {
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    expect(
      screen.queryByRole("navigation", { name: "pagination" })
    ).not.toBeInTheDocument();
  });

  it("changes pages through the shared pagination", async () => {
    const user = userEvent.setup();
    render(
      <RepositoriesList
        repositories={mockRepositories}
        {...defaultPaginationProps}
        currentPage={1}
        totalPages={3}
        onDelete={mockOnDelete}
        onCreate={mockOnCreate}
      />
    );

    const previous = screen.getByRole("button", { name: "Go to previous page" });
    const next = screen.getByRole("button", { name: "Go to next page" });

    expect(previous).toBeDisabled();
    expect(next).toBeEnabled();
    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute(
      "aria-current",
      "page"
    );

    await user.click(screen.getByRole("button", { name: "3" }));
    expect(mockOnPageChange).toHaveBeenCalledWith(3);

    await user.click(next);
    expect(mockOnPageChange).toHaveBeenCalledWith(2);
  });
});
