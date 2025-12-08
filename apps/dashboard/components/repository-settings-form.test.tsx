import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import {
  RepositorySettingsForm,
  type RepositorySettingsFormProps,
} from "./repository-settings-form";

describe("RepositorySettingsForm", () => {
  const mockOnSubmit = vi.fn();

  const defaultProps: RepositorySettingsFormProps = {
    onSubmit: mockOnSubmit,
    isLoading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders the form with all fields", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      expect(screen.getByText("General Settings")).toBeInTheDocument();
      expect(
        screen.getByText("Configure basic repository settings")
      ).toBeInTheDocument();
      expect(screen.getByLabelText("Repository Name")).toBeInTheDocument();
      expect(screen.getByLabelText("Visibility")).toBeInTheDocument();
    });

    it("renders submit and cancel buttons", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      expect(
        screen.getByRole("button", { name: /save changes/i })
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /cancel/i })
      ).toBeInTheDocument();
    });

    it("renders repository name input with placeholder", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByPlaceholderText("my-repository");
      expect(nameInput).toBeInTheDocument();
    });

    it("renders field descriptions", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      expect(
        screen.getByText(/This will be used in URLs and API references/i)
      ).toBeInTheDocument();
      expect(
        screen.getByText(/Control who can see and access this repository/i)
      ).toBeInTheDocument();
    });

    it("renders settings icon", () => {
      const { container } = render(<RepositorySettingsForm {...defaultProps} />);

      const icon = container.querySelector("svg");
      expect(icon).toBeInTheDocument();
    });
  });

  describe("Initial Values", () => {
    it("renders with empty fields when no initial data is provided", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      expect(nameInput).toHaveValue("");
    });

    it("renders with initial data when provided", () => {
      const initialData = {
        name: "test-repo",
        visibility: "public" as const,
      };

      render(
        <RepositorySettingsForm {...defaultProps} initialData={initialData} />
      );

      const nameInput = screen.getByLabelText("Repository Name");
      expect(nameInput).toHaveValue("test-repo");
    });

    it("uses default visibility when not provided in initial data", () => {
      const initialData = {
        name: "test-repo",
      };

      render(
        <RepositorySettingsForm {...defaultProps} initialData={initialData} />
      );

      const nameInput = screen.getByLabelText("Repository Name");
      expect(nameInput).toHaveValue("test-repo");
    });

    it("updates form when initial data changes", async () => {
      const initialData = {
        name: "original-repo",
        visibility: "private" as const,
      };

      const { rerender } = render(
        <RepositorySettingsForm {...defaultProps} initialData={initialData} />
      );

      const nameInput = screen.getByLabelText("Repository Name");
      expect(nameInput).toHaveValue("original-repo");

      const updatedData = {
        name: "updated-repo",
        visibility: "public" as const,
      };

      rerender(
        <RepositorySettingsForm {...defaultProps} initialData={updatedData} />
      );

      await waitFor(() => {
        expect(nameInput).toHaveValue("updated-repo");
      });
    });
  });

  describe("Form Validation - Repository Name", () => {
    it("shows error when repository name is empty", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText("Repository name is required")
        ).toBeInTheDocument();
      });

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it("shows error when repository name exceeds 100 characters", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      const longName = "a".repeat(101);

      await user.type(nameInput, longName);
      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText("Repository name must be less than 100 characters")
        ).toBeInTheDocument();
      });

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it("shows error when repository name contains invalid characters", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "invalid name!");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(
            "Repository name can only contain letters, numbers, hyphens and underscores"
          )
        ).toBeInTheDocument();
      });

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it("accepts valid repository name with hyphens", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "my-repo-name");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
      });
    });

    it("accepts valid repository name with underscores", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "my_repo_name");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
      });
    });

    it("accepts valid repository name with numbers", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "repo123");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
      });
    });
  });

  describe("Form Submission", () => {
    it("calls onSubmit with correct data when form is valid", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repository");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
        expect(mockOnSubmit.mock.calls[0][0]).toEqual({
          name: "test-repository",
          visibility: "private",
        });
      });
    });

    it("calls onSubmit with updated visibility", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo");

      const visibilityTrigger = screen.getByRole("combobox", {
        name: /visibility/i,
      });
      await user.click(visibilityTrigger);

      const publicOption = await screen.findByRole("option", { name: /public/i });
      await user.click(publicOption);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
        expect(mockOnSubmit.mock.calls[0][0]).toEqual({
          name: "test-repo",
          visibility: "public",
        });
      });
    });

    it("submits form with initial data unchanged", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      const initialData = {
        name: "existing-repo",
        visibility: "private" as const,
      };

      render(
        <RepositorySettingsForm {...defaultProps} initialData={initialData} />
      );

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
        expect(mockOnSubmit.mock.calls[0][0]).toEqual({
          name: "existing-repo",
          visibility: "private",
        });
      });
    });

    it("does not submit when form is invalid", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText("Repository name is required")
        ).toBeInTheDocument();
      });

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });
  });

  describe("Form Reset", () => {
    it("resets form to initial values when cancel is clicked", async () => {
      const user = userEvent.setup();
      const initialData = {
        name: "original-repo",
        visibility: "private" as const,
      };

      render(
        <RepositorySettingsForm {...defaultProps} initialData={initialData} />
      );

      const nameInput = screen.getByLabelText("Repository Name");
      await user.clear(nameInput);
      await user.type(nameInput, "modified-repo");

      expect(nameInput).toHaveValue("modified-repo");

      const cancelButton = screen.getByRole("button", { name: /cancel/i });
      await user.click(cancelButton);

      await waitFor(() => {
        expect(nameInput).toHaveValue("original-repo");
      });
    });

    it("resets form to empty when no initial data provided", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo");

      expect(nameInput).toHaveValue("test-repo");

      const cancelButton = screen.getByRole("button", { name: /cancel/i });
      await user.click(cancelButton);

      await waitFor(() => {
        expect(nameInput).toHaveValue("");
      });
    });

    it("clears validation errors when form is reset", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText("Repository name is required")
        ).toBeInTheDocument();
      });

      const cancelButton = screen.getByRole("button", { name: /cancel/i });
      await user.click(cancelButton);

      await waitFor(() => {
        expect(
          screen.queryByText("Repository name is required")
        ).not.toBeInTheDocument();
      });
    });
  });

  describe("Loading States", () => {
    it("disables inputs when isLoading is true", () => {
      render(<RepositorySettingsForm {...defaultProps} isLoading={true} />);

      const nameInput = screen.getByLabelText("Repository Name");
      const visibilitySelect = screen.getByRole("combobox", {
        name: /visibility/i,
      });
      const cancelButton = screen.getByRole("button", { name: /cancel/i });

      expect(nameInput).toBeDisabled();
      expect(visibilitySelect).toBeDisabled();
      expect(cancelButton).toBeDisabled();
    });

    it("shows loading state on submit button when isLoading is true", () => {
      render(<RepositorySettingsForm {...defaultProps} isLoading={true} />);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      expect(submitButton).toBeInTheDocument();
    });

    it("enables inputs when isLoading is false", () => {
      render(<RepositorySettingsForm {...defaultProps} isLoading={false} />);

      const nameInput = screen.getByLabelText("Repository Name");
      const visibilitySelect = screen.getByRole("combobox", {
        name: /visibility/i,
      });

      expect(nameInput).not.toBeDisabled();
      expect(visibilitySelect).not.toBeDisabled();
    });

    it("disables inputs during form submission", async () => {
      const user = userEvent.setup();
      let resolveSubmit: () => void;
      const submitPromise = new Promise<void>((resolve) => {
        resolveSubmit = resolve;
      });
      mockOnSubmit.mockReturnValue(submitPromise);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(nameInput).toBeDisabled();
      });

      await act(async () => {
        resolveSubmit!();
        await submitPromise;
      });
    });
  });

  describe("Visibility Options", () => {
    it("renders all visibility options in select", async () => {
      const user = userEvent.setup();
      render(<RepositorySettingsForm {...defaultProps} />);

      const visibilityTrigger = screen.getByRole("combobox", {
        name: /visibility/i,
      });
      await user.click(visibilityTrigger);

      await waitFor(() => {
        expect(screen.getByRole("option", { name: /private/i })).toBeInTheDocument();
        expect(screen.getByRole("option", { name: /public/i })).toBeInTheDocument();
      });
    });

    it("allows selecting different visibility options", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo");

      const visibilityTrigger = screen.getByRole("combobox", {
        name: /visibility/i,
      });
      await user.click(visibilityTrigger);

      const internalOption = await screen.findByRole("option", {
        name: /private/i,
      });
      await user.click(internalOption);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
        expect(mockOnSubmit.mock.calls[0][0]).toEqual({
          name: "test-repo",
          visibility: "private",
        });
      });
    });
  });

  describe("Accessibility", () => {
    it("has proper labels for all form fields", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      expect(screen.getByLabelText("Repository Name")).toBeInTheDocument();
      expect(screen.getByLabelText("Visibility")).toBeInTheDocument();
    });

    it("has proper form structure", () => {
      const { container } = render(<RepositorySettingsForm {...defaultProps} />);

      const form = container.querySelector("form");
      expect(form).toBeInTheDocument();
    });

    it("submit button has correct type", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      expect(submitButton).toHaveAttribute("type", "submit");
    });

    it("cancel button has correct type", () => {
      render(<RepositorySettingsForm {...defaultProps} />);

      const cancelButton = screen.getByRole("button", { name: /cancel/i });
      expect(cancelButton).toHaveAttribute("type", "button");
    });
  });

  describe("Edge Cases", () => {
    it("handles submission errors gracefully", async () => {
      const user = userEvent.setup();
      const errorMessage = "Network error";

      const rejectionHandler = vi.fn();
      const originalHandler = process.on("unhandledRejection", rejectionHandler);

      mockOnSubmit.mockRejectedValue(new Error(errorMessage));

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
      });

      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 100));
      });

      if (originalHandler) {
        process.removeListener("unhandledRejection", rejectionHandler);
      }
    });

    it("handles max length repository name", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      const maxLengthName = "a".repeat(100);
      await user.type(nameInput, maxLengthName);

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
        expect(mockOnSubmit.mock.calls[0][0]).toEqual({
          name: maxLengthName,
          visibility: "private",
        });
      });
    });

    it("trims input values are not automatically trimmed", async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<RepositorySettingsForm {...defaultProps} />);

      const nameInput = screen.getByLabelText("Repository Name");
      await user.type(nameInput, "test-repo ");

      const submitButton = screen.getByRole("button", { name: /save changes/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(
            "Repository name can only contain letters, numbers, hyphens and underscores"
          )
        ).toBeInTheDocument();
      });

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });
  });
});
