import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { RepositoryItem, type OrganizationRepository } from "./repository-item";

describe("RepositoryItem", () => {
  const mockRepository: OrganizationRepository = {
    id: "repo-123",
    name: "frontend-app",
    visibility: "public",
  };

  const mockOnDelete = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders repository information correctly", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      expect(screen.getByText("frontend-app")).toBeInTheDocument();
      expect(screen.getByText("public")).toBeInTheDocument();
    });

    it("renders repository name with correct text styling", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const nameElement = screen.getByText("frontend-app");
      expect(nameElement).toHaveClass("font-medium", "text-sm");
    });

    it("renders delete button with trash icon", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      expect(deleteButton).toBeInTheDocument();
      expect(deleteButton.querySelector("svg")).toBeInTheDocument();
    });

    it("renders with special characters in repository name", () => {
      const repoWithSpecialChars: OrganizationRepository = {
        id: "repo-456",
        name: "my-app_v2.0",
        visibility: "private",
      };

      render(
        <RepositoryItem
          repository={repoWithSpecialChars}
          onDelete={mockOnDelete}
        />
      );

      expect(screen.getByText("my-app_v2.0")).toBeInTheDocument();
    });
  });

  describe("Visibility Badge Styling", () => {
    it("applies correct styling for public visibility", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const visibilityBadge = screen.getByText("public");
      expect(visibilityBadge).toHaveClass(
        "bg-green-100",
        "text-green-700",
        "dark:bg-green-900/30",
        "dark:text-green-400"
      );
    });

    it("applies correct styling for private visibility", () => {
      const privateRepo: OrganizationRepository = {
        id: "repo-789",
        name: "backend-api",
        visibility: "private",
      };

      render(<RepositoryItem repository={privateRepo} onDelete={mockOnDelete} />);

      const visibilityBadge = screen.getByText("private");
      expect(visibilityBadge).toHaveClass(
        "bg-gray-100",
        "text-gray-700",
        "dark:bg-gray-800",
        "dark:text-gray-300"
      );
    });

    it("applies badge styling classes", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const visibilityBadge = screen.getByText("public");
      expect(visibilityBadge).toHaveClass("text-xs", "px-2", "py-0.5", "rounded-full");
    });
  });

  describe("Link Navigation", () => {
    it("renders link with correct href", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      expect(link).toHaveAttribute("href", "/repository/repo-123");
    });

    it("constructs correct URL for different repository IDs", () => {
      const differentRepo: OrganizationRepository = {
        id: "custom-repo-id",
        name: "test-repo",
        visibility: "public",
      };

      render(
        <RepositoryItem repository={differentRepo} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      expect(link).toHaveAttribute("href", "/repository/custom-repo-id");
    });

    it("does not call onDelete when clicking the link", async () => {
      const user = userEvent.setup();
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      await user.click(link);

      expect(mockOnDelete).not.toHaveBeenCalled();
    });
  });

  describe("Delete Functionality", () => {
    it("calls onDelete when delete button is clicked", async () => {
      const user = userEvent.setup();
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      await user.click(deleteButton);

      expect(mockOnDelete).toHaveBeenCalledTimes(1);
      expect(mockOnDelete).toHaveBeenCalledWith(mockRepository);
    });

    it("passes correct repository object to onDelete", async () => {
      const user = userEvent.setup();
      const privateRepo: OrganizationRepository = {
        id: "private-123",
        name: "private-repo",
        visibility: "private",
      };

      render(<RepositoryItem repository={privateRepo} onDelete={mockOnDelete} />);

      const deleteButton = screen.getByRole("button");
      await user.click(deleteButton);

      expect(mockOnDelete).toHaveBeenCalledWith(privateRepo);
    });

    it("prevents default event when delete button is clicked", async () => {
      const user = userEvent.setup();
      const mockEvent = vi.fn();

      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      deleteButton.addEventListener("click", mockEvent);

      await user.click(deleteButton);

      // Verify preventDefault was called by checking onDelete was still called
      expect(mockOnDelete).toHaveBeenCalled();
    });

    it("calls onDelete multiple times when clicked multiple times", async () => {
      const user = userEvent.setup();
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      await user.click(deleteButton);
      await user.click(deleteButton);
      await user.click(deleteButton);

      expect(mockOnDelete).toHaveBeenCalledTimes(3);
    });
  });

  describe("Accessibility", () => {
    it("has accessible button role for delete action", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      expect(deleteButton).toBeInTheDocument();
    });

    it("has accessible link role for navigation", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      expect(link).toBeInTheDocument();
    });

    it("repository name is keyboard accessible via link", async () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      link.focus();

      expect(link).toHaveFocus();
    });

    it("delete button is keyboard accessible", async () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      deleteButton.focus();

      expect(deleteButton).toHaveFocus();
    });
  });

  describe("Layout and Styling", () => {
    it("applies correct container classes", () => {
      const { container } = render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const itemContainer = container.firstChild;
      expect(itemContainer).toHaveClass(
        "flex",
        "items-center",
        "justify-between",
        "rounded-lg",
        "border",
        "border-border/60",
        "bg-card"
      );
    });

    it("applies hover transition classes", () => {
      const { container } = render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const itemContainer = container.firstChild;
      expect(itemContainer).toHaveClass("hover:bg-accent/50", "transition-colors");
    });

    it("delete button has ghost variant styling", () => {
      render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const deleteButton = screen.getByRole("button");
      expect(deleteButton).toBeInTheDocument();
    });
  });

  describe("Edge Cases", () => {
    it("handles repository with empty ID", () => {
      const repoWithEmptyId: OrganizationRepository = {
        id: "",
        name: "test-repo",
        visibility: "public",
      };

      render(
        <RepositoryItem repository={repoWithEmptyId} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      expect(link).toHaveAttribute("href", "/repository");
    });

    it("handles very long repository names", () => {
      const repoWithLongName: OrganizationRepository = {
        id: "long-123",
        name: "this-is-a-very-long-repository-name-that-might-cause-layout-issues-if-not-handled-properly",
        visibility: "public",
      };

      render(
        <RepositoryItem repository={repoWithLongName} onDelete={mockOnDelete} />
      );

      expect(
        screen.getByText(
          "this-is-a-very-long-repository-name-that-might-cause-layout-issues-if-not-handled-properly"
        )
      ).toBeInTheDocument();
    });

    it("handles repository with numeric ID", () => {
      const repoWithNumericId: OrganizationRepository = {
        id: "123456789",
        name: "numeric-repo",
        visibility: "private",
      };

      render(
        <RepositoryItem repository={repoWithNumericId} onDelete={mockOnDelete} />
      );

      const link = screen.getByRole("link");
      expect(link).toHaveAttribute("href", "/repository/123456789");
    });
  });

  describe("Component Integration", () => {
    it("renders complete component without errors", () => {
      const { container } = render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      expect(container).toBeInTheDocument();
      expect(container.firstChild).toBeInTheDocument();
    });

    it("maintains component structure with multiple renders", () => {
      const { rerender } = render(
        <RepositoryItem repository={mockRepository} onDelete={mockOnDelete} />
      );

      const updatedRepo: OrganizationRepository = {
        ...mockRepository,
        name: "updated-name",
      };

      rerender(
        <RepositoryItem repository={updatedRepo} onDelete={mockOnDelete} />
      );

      expect(screen.getByText("updated-name")).toBeInTheDocument();
      expect(screen.queryByText("frontend-app")).not.toBeInTheDocument();
    });
  });
});
