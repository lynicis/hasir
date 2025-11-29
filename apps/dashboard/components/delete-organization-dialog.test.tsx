import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DeleteOrganizationDialog } from "./delete-organization-dialog";

describe("DeleteOrganizationDialog", () => {
  const mockOnOpenChange = vi.fn();
  const mockOnConfirm = vi.fn();

  it("renders dialog when open", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", { name: /delete organization/i })
    ).toBeInTheDocument();
    expect(screen.getByText(/test organization/i)).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    render(
      <DeleteOrganizationDialog
        open={false}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.queryByRole("heading", { name: /delete organization/i })
    ).not.toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const confirmButton = screen.getByRole("button", {
      name: /delete organization/i,
    });
    await user.click(confirmButton);

    expect(mockOnConfirm).toHaveBeenCalledTimes(1);
  });

  it("calls onOpenChange when cancel button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });

  it("handles null organization name gracefully", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName={null}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", { name: /delete organization/i })
    ).toBeInTheDocument();
  });

  it("displays warning message about permanent deletion", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByText(
        /cannot be undone and all data will be permanently deleted/i
      )
    ).toBeInTheDocument();
  });

  it("disables buttons when isDeleting is true", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        isDeleting={true}
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    const confirmButton = screen.getByRole("button", {
      name: /delete organization/i,
    });

    expect(cancelButton).toBeDisabled();
    expect(confirmButton).toBeDisabled();
  });

  it("shows spinner when isDeleting is true", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        isDeleting={true}
      />
    );

    const spinner = screen.getByRole("status", { name: /loading/i });
    expect(spinner).toBeInTheDocument();
  });

  it("does not disable buttons when isDeleting is false", () => {
    render(
      <DeleteOrganizationDialog
        open={true}
        organizationName="Test Organization"
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        isDeleting={false}
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    const confirmButton = screen.getByRole("button", {
      name: /delete organization/i,
    });

    expect(cancelButton).not.toBeDisabled();
    expect(confirmButton).not.toBeDisabled();
  });
});
