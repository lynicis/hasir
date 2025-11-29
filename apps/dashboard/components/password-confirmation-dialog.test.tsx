import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { PasswordConfirmationDialog } from "./password-confirmation-dialog";

describe("PasswordConfirmationDialog", () => {
  const mockOnConfirm = vi.fn();
  const mockOnCancel = vi.fn();
  const mockOnOpenChange = vi.fn();

  beforeEach(() => {
    mockOnConfirm.mockReset();
    mockOnCancel.mockReset();
    mockOnOpenChange.mockReset();
  });

  it("renders dialog when open is true", () => {
    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    expect(screen.getByText(/confirm your password/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/current password/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /confirm/i })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /cancel/i })).toBeInTheDocument();
  });

  it("does not render dialog when open is false", () => {
    render(
      <PasswordConfirmationDialog
        open={false}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    expect(
      screen.queryByText(/confirm your password/i)
    ).not.toBeInTheDocument();
  });

  it("calls onConfirm with password when form is submitted", async () => {
    const user = userEvent.setup();

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    await user.type(
      screen.getByLabelText(/current password/i),
      "currentpass123"
    );
    await user.click(screen.getByRole("button", { name: /confirm/i }));

    await waitFor(() =>
      expect(mockOnConfirm).toHaveBeenCalledWith({
        currentPassword: "currentpass123",
      })
    );
  });

  it("prevents submission when password is empty", async () => {
    const user = userEvent.setup();

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    const passwordInput = screen.getByLabelText(/current password/i);
    await user.click(screen.getByRole("button", { name: /confirm/i }));

    await waitFor(() => {
      expect(passwordInput).toHaveAttribute("aria-invalid", "true");
    });
    expect(mockOnConfirm).not.toHaveBeenCalled();
  });

  it("calls onCancel when cancel button is clicked", async () => {
    const user = userEvent.setup();

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(mockOnCancel).toHaveBeenCalled();
    expect(mockOnConfirm).not.toHaveBeenCalled();
  });

  it("resets form after successful submission", async () => {
    const user = userEvent.setup();
    mockOnConfirm.mockResolvedValue(undefined);

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    const passwordInput = screen.getByLabelText(
      /current password/i
    ) as HTMLInputElement;
    await user.type(passwordInput, "currentpass123");
    await user.click(screen.getByRole("button", { name: /confirm/i }));

    await waitFor(() => expect(mockOnConfirm).toHaveBeenCalled());
    expect(passwordInput.value).toBe("");
  });

  it("resets form when cancel is clicked", async () => {
    const user = userEvent.setup();

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    const passwordInput = screen.getByLabelText(
      /current password/i
    ) as HTMLInputElement;
    await user.type(passwordInput, "somepassword");
    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(passwordInput.value).toBe("");
  });

  it("handles async onConfirm callback", async () => {
    const user = userEvent.setup();
    let resolvePromise: () => void;
    const promise = new Promise<void>((resolve) => {
      resolvePromise = resolve;
    });
    mockOnConfirm.mockReturnValue(promise);

    render(
      <PasswordConfirmationDialog
        open={true}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    await user.type(
      screen.getByLabelText(/current password/i),
      "currentpass123"
    );
    await user.click(screen.getByRole("button", { name: /confirm/i }));

    await waitFor(() => expect(mockOnConfirm).toHaveBeenCalled());

    const confirmButton = screen.getByRole("button", { name: /confirm/i });
    expect(confirmButton).toBeDisabled();

    await act(async () => {
      resolvePromise!();
      await promise;
    });
  });
});
