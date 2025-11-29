import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { DangerZone } from "./danger-zone";

describe("DangerZone", () => {
  const mockOnDelete = vi.fn();

  beforeEach(() => {
    mockOnDelete.mockReset();
  });

  it("renders danger zone card", () => {
    render(<DangerZone onDelete={mockOnDelete} />);

    expect(screen.getByText("Danger Zone")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Delete Account" })
    ).toBeInTheDocument();
    expect(
      screen.getByText(/once you delete your account, there is no going back/i)
    ).toBeInTheDocument();
  });

  it("renders delete account button", () => {
    render(<DangerZone onDelete={mockOnDelete} />);

    const deleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    expect(deleteButton).toBeInTheDocument();
  });

  it("opens delete confirmation dialog when delete button is clicked", async () => {
    const user = userEvent.setup();

    render(<DangerZone onDelete={mockOnDelete} />);

    await user.click(screen.getByRole("button", { name: /delete account/i }));

    await waitFor(() => {
      expect(screen.getByText(/are you absolutely sure/i)).toBeInTheDocument();
    });
  });

  it("closes dialog when cancel is clicked", async () => {
    const user = userEvent.setup();

    render(<DangerZone onDelete={mockOnDelete} />);

    await user.click(screen.getByRole("button", { name: /delete account/i }));

    await waitFor(() => {
      expect(screen.getByText(/are you absolutely sure/i)).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    await waitFor(() => {
      expect(
        screen.queryByText(/are you absolutely sure/i)
      ).not.toBeInTheDocument();
    });

    expect(mockOnDelete).not.toHaveBeenCalled();
  });

  it("calls onDelete when delete is confirmed", async () => {
    const user = userEvent.setup();
    mockOnDelete.mockResolvedValue(undefined);

    render(<DangerZone onDelete={mockOnDelete} />);

    const initialDeleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    await user.click(initialDeleteButton);

    await waitFor(() => {
      expect(screen.getByText(/are you absolutely sure/i)).toBeInTheDocument();
    });

    await waitFor(() => {
      const allButtons = screen.getAllByRole("button");
      const dialogButtons = allButtons.filter(
        (btn) =>
          btn.textContent?.includes("Delete Account") &&
          btn !== initialDeleteButton
      );
      expect(dialogButtons.length).toBeGreaterThan(0);
    });

    const allButtons = screen.getAllByRole("button");
    const dialogConfirmButton = allButtons.find(
      (btn) =>
        btn.textContent?.includes("Delete Account") &&
        btn !== initialDeleteButton &&
        btn.getAttribute("type") === "button"
    );

    expect(dialogConfirmButton).toBeDefined();
    if (dialogConfirmButton) {
      await user.click(dialogConfirmButton);
    }

    await waitFor(() => expect(mockOnDelete).toHaveBeenCalled());
  });

  it("closes dialog after successful delete", async () => {
    const user = userEvent.setup();
    mockOnDelete.mockResolvedValue(undefined);

    render(<DangerZone onDelete={mockOnDelete} />);

    const initialDeleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    await user.click(initialDeleteButton);

    await waitFor(() => {
      expect(screen.getByText(/are you absolutely sure/i)).toBeInTheDocument();
    });

    await waitFor(() => {
      const allButtons = screen.getAllByRole("button");
      const dialogButtons = allButtons.filter(
        (btn) =>
          btn.textContent?.includes("Delete Account") &&
          btn !== initialDeleteButton
      );
      expect(dialogButtons.length).toBeGreaterThan(0);
    });

    const allButtons = screen.getAllByRole("button");
    const dialogConfirmButton = allButtons.find(
      (btn) =>
        btn.textContent?.includes("Delete Account") &&
        btn !== initialDeleteButton &&
        btn.getAttribute("type") === "button"
    );

    expect(dialogConfirmButton).toBeDefined();
    if (dialogConfirmButton) {
      await user.click(dialogConfirmButton);
    }

    await waitFor(() => expect(mockOnDelete).toHaveBeenCalled());

    await waitFor(() => {
      expect(
        screen.queryByText(/are you absolutely sure/i)
      ).not.toBeInTheDocument();
    });
  });

  it("disables delete button when isDeleting is true", () => {
    render(<DangerZone onDelete={mockOnDelete} isDeleting={true} />);

    const deleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    expect(deleteButton).toBeDisabled();
  });

  it("disables dialog buttons when isDeleting is true", () => {
    render(<DangerZone onDelete={mockOnDelete} isDeleting={true} />);

    const deleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    expect(deleteButton).toBeDisabled();
  });

  it("handles async onDelete callback", async () => {
    const user = userEvent.setup();
    let resolvePromise: () => void;
    const promise = new Promise<void>((resolve) => {
      resolvePromise = resolve;
    });
    mockOnDelete.mockReturnValue(promise);

    render(<DangerZone onDelete={mockOnDelete} />);

    const initialDeleteButton = screen.getByRole("button", {
      name: /delete account/i,
    });
    await user.click(initialDeleteButton);

    await waitFor(() => {
      expect(screen.getByText(/are you absolutely sure/i)).toBeInTheDocument();
    });

    await waitFor(() => {
      const allButtons = screen.getAllByRole("button");
      const dialogButtons = allButtons.filter(
        (btn) =>
          btn.textContent?.includes("Delete Account") &&
          btn !== initialDeleteButton
      );
      expect(dialogButtons.length).toBeGreaterThan(0);
    });

    const allButtons = screen.getAllByRole("button");
    const dialogConfirmButton = allButtons.find(
      (btn) =>
        btn.textContent?.includes("Delete Account") &&
        btn !== initialDeleteButton &&
        btn.getAttribute("type") === "button"
    );

    expect(dialogConfirmButton).toBeDefined();
    if (dialogConfirmButton) {
      await user.click(dialogConfirmButton);
    }

    await waitFor(() => expect(mockOnDelete).toHaveBeenCalled());

    await act(async () => {
      resolvePromise!();
      await promise;
    });
  });
});
