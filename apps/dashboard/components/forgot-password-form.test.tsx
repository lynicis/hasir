import type { Mock } from "vitest";

import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { ForgotPasswordForm } from "./forgot-password-form";

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccess,
    error: toastError,
  },
}));

const mockForgotPassword = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

describe("ForgotPasswordForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      forgotPassword: mockForgotPassword,
    } as never);
    mockForgotPassword.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  it("submits email and shows success message", async () => {
    const user = userEvent.setup();

    mockForgotPassword.mockResolvedValueOnce(undefined);

    render(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() =>
      expect(mockForgotPassword).toHaveBeenCalledWith({
        email: "hello@example.com",
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith("Password reset email sent!")
    );

    expect(toastError).not.toHaveBeenCalled();

    await waitFor(() =>
      expect(screen.getByText(/check your email/i)).toBeInTheDocument()
    );
  });

  it("displays confirmation screen after successful submission", async () => {
    const user = userEvent.setup();

    mockForgotPassword.mockResolvedValueOnce(undefined);

    render(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() =>
      expect(
        screen.getByText(/we've sent you an email with instructions/i)
      ).toBeInTheDocument()
    );
  });

  it("allows user to try again from confirmation screen", async () => {
    const user = userEvent.setup();

    mockForgotPassword.mockResolvedValueOnce(undefined);

    render(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() =>
      expect(screen.getByText(/check your email/i)).toBeInTheDocument()
    );

    const tryAgainButton = screen.getByRole("button", { name: /try again/i });
    await user.click(tryAgainButton);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /send reset link/i })
    ).toBeInTheDocument();
  });

  it("prevents submission when email is invalid", async () => {
    const user = userEvent.setup();

    render(<ForgotPasswordForm />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    await user.type(emailInput, "not-an-email");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() => expect(mockForgotPassword).not.toHaveBeenCalled());
    expect(emailInput.checkValidity()).toBe(false);
  });

  it("prevents submission when email is missing", async () => {
    const user = userEvent.setup();

    render(<ForgotPasswordForm />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() => expect(mockForgotPassword).not.toHaveBeenCalled());
    expect(emailInput.validity.valueMissing).toBe(true);
  });

  it("displays error toast when API call fails", async () => {
    const user = userEvent.setup();

    mockForgotPassword.mockRejectedValueOnce(
      new Error("Failed to send reset email")
    );

    render(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email/i), "error@example.com");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith(
        "Failed to send reset email"
      )
    );
    expect(toastSuccess).not.toHaveBeenCalled();
  });

  it("displays generic error when unknown error occurs", async () => {
    const user = userEvent.setup();

    mockForgotPassword.mockRejectedValueOnce("Unknown error");

    render(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email/i), "error@example.com");
    await user.click(
      screen.getByRole("button", { name: /send reset link/i })
    );

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while sending reset email."
      )
    );
    expect(toastSuccess).not.toHaveBeenCalled();
  });

  it("shows link to login page", () => {
    render(<ForgotPasswordForm />);

    const loginLink = screen.getByRole("link", { name: /login/i });
    expect(loginLink).toBeInTheDocument();
    expect(loginLink).toHaveAttribute("href", "/login");
  });
});
