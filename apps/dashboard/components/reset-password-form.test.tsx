import type { Mock } from "vitest";

import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { ResetPasswordForm } from "./reset-password-form";

const { mockRouterPush, toastSuccess, toastError } = vi.hoisted(() => ({
  mockRouterPush: vi.fn(),
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockRouterPush,
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccess,
    error: toastError,
  },
}));

const mockResetPassword = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

describe("ResetPasswordForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      resetPassword: mockResetPassword,
    } as never);
    mockResetPassword.mockReset();
    mockRouterPush.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  it("submits valid password and navigates to login", async () => {
    const user = userEvent.setup();
    const testToken = "test-reset-token-123";

    mockResetPassword.mockResolvedValueOnce(undefined);

    render(<ResetPasswordForm token={testToken} />);

    await user.type(screen.getByLabelText(/new password/i), "newpassword123");
    await user.type(
      screen.getByLabelText(/confirm password/i),
      "newpassword123"
    );

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(mockResetPassword).toHaveBeenCalledWith({
        token: testToken,
        newPassword: "newpassword123",
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith(
        "Password reset successful! You can now log in."
      )
    );

    expect(toastError).not.toHaveBeenCalled();

    await waitFor(() => expect(mockRouterPush).toHaveBeenCalledWith("/login"), {
      timeout: 2000,
    });
  });

  it("passes the token prop to the API call", async () => {
    const user = userEvent.setup();
    const customToken = "custom-token-456";

    mockResetPassword.mockResolvedValueOnce(undefined);

    render(<ResetPasswordForm token={customToken} />);

    await user.type(screen.getByLabelText(/new password/i), "validpass123");
    await user.type(screen.getByLabelText(/confirm password/i), "validpass123");

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(mockResetPassword).toHaveBeenCalledWith({
        token: customToken,
        newPassword: "validpass123",
      })
    );
  });

  it("prevents submission when passwords do not match", async () => {
    const user = userEvent.setup();

    render(<ResetPasswordForm token="test-token" />);

    await user.type(screen.getByLabelText(/new password/i), "password123");
    const confirmInput = screen.getByLabelText(/confirm password/i);
    await user.type(confirmInput, "different123");

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(confirmInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockResetPassword).not.toHaveBeenCalled();
  });

  it("prevents submission when password is too short", async () => {
    const user = userEvent.setup();

    render(<ResetPasswordForm token="test-token" />);

    const passwordInput = screen.getByLabelText(
      /new password/i
    ) as HTMLInputElement;
    await user.type(passwordInput, "short");
    await user.type(screen.getByLabelText(/confirm password/i), "short");

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(passwordInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockResetPassword).not.toHaveBeenCalled();
  });

  it("prevents submission when password exceeds max length", async () => {
    const user = userEvent.setup();

    render(<ResetPasswordForm token="test-token" />);

    const longPassword = "a".repeat(21);
    const passwordInput = screen.getByLabelText(
      /new password/i
    ) as HTMLInputElement;
    await user.type(passwordInput, longPassword);
    await user.type(screen.getByLabelText(/confirm password/i), longPassword);

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(passwordInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockResetPassword).not.toHaveBeenCalled();
  });

  it("prevents submission when password is missing", async () => {
    const user = userEvent.setup();

    render(<ResetPasswordForm token="test-token" />);

    const passwordInput = screen.getByLabelText(
      /new password/i
    ) as HTMLInputElement;
    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() => expect(mockResetPassword).not.toHaveBeenCalled());
    expect(passwordInput.validity.valueMissing).toBe(true);
  });

  it("prevents submission when confirm password is missing", async () => {
    const user = userEvent.setup();

    render(<ResetPasswordForm token="test-token" />);

    const confirmInput = screen.getByLabelText(
      /confirm password/i
    ) as HTMLInputElement;
    await user.type(screen.getByLabelText(/new password/i), "password123");
    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() => expect(mockResetPassword).not.toHaveBeenCalled());
    expect(confirmInput.validity.valueMissing).toBe(true);
  });

  it("displays error toast when API call fails", async () => {
    const user = userEvent.setup();

    mockResetPassword.mockRejectedValueOnce(
      new Error("Invalid or expired token")
    );

    render(<ResetPasswordForm token="invalid-token" />);

    await user.type(screen.getByLabelText(/new password/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Invalid or expired token")
    );
    expect(toastSuccess).not.toHaveBeenCalled();
    expect(mockRouterPush).not.toHaveBeenCalled();
  });

  it("displays generic error when unknown error occurs", async () => {
    const user = userEvent.setup();

    mockResetPassword.mockRejectedValueOnce("Unknown error");

    render(<ResetPasswordForm token="test-token" />);

    await user.type(screen.getByLabelText(/new password/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(
      screen.getByRole("button", { name: /reset password/i })
    );

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while resetting password."
      )
    );
    expect(toastSuccess).not.toHaveBeenCalled();

    await waitFor(
      () => {
        expect(mockRouterPush).not.toHaveBeenCalled();
      },
      { timeout: 800 }
    );
  });

  it("shows link to login page", () => {
    render(<ResetPasswordForm token="test-token" />);

    const loginLink = screen.getByRole("link", { name: /login/i });
    expect(loginLink).toBeInTheDocument();
    expect(loginLink).toHaveAttribute("href", "/login");
  });

  it("shows password requirements hint", () => {
    render(<ResetPasswordForm token="test-token" />);

    expect(
      screen.getByText(/must be at least 8 characters long/i)
    ).toBeInTheDocument();
  });
});
