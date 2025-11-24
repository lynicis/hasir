import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useUserStore } from "@/stores/user-store-provider";
import { useClient } from "@/lib/use-client";

import { LoginForm } from "./login-form";

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock("@/stores/user-store-provider", () => ({
  useUserStore: vi.fn(),
}));

const mockLogin = vi.fn();
const mockPush = vi.fn();
const mockedUseClient = vi.mocked(useClient);
const mockedUseUserStore = vi.mocked(useUserStore);

describe("LoginForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      login: mockLogin,
    } as never);
    mockedUseUserStore.mockReturnValue({
      setTokens: vi.fn(),
    } as never);
    mockLogin.mockReset();
    mockPush.mockReset();
  });

  it("submits credentials when the form is valid", async () => {
    const user = userEvent.setup();
    const mockSetTokens = vi.fn();
    mockedUseUserStore.mockReturnValue({
      setTokens: mockSetTokens,
    } as never);
    mockLogin.mockResolvedValue({
      accessToken: "token",
      refreshToken: "refresh",
    });

    render(<LoginForm />);

    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() =>
      expect(mockLogin).toHaveBeenCalledWith({
        email: "hello@example.com",
        password: "password123",
      })
    );
  });

  it("prevents submission when the email is invalid", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    await user.type(emailInput, "not-an-email");
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() => expect(mockLogin).not.toHaveBeenCalled());
    expect(emailInput.checkValidity()).toBe(false);
  });

  it("prevents submission when the password is too short", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(
      /password/i
    ) as HTMLInputElement;
    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(passwordInput, "short");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() =>
      expect(passwordInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it("prevents submission when the password exceeds the max length", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(
      /password/i
    ) as HTMLInputElement;
    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(passwordInput, "a".repeat(21));
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() =>
      expect(passwordInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it("prevents submission when the password is missing", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(
      /password/i
    ) as HTMLInputElement;
    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() => expect(mockLogin).not.toHaveBeenCalled());
    expect(passwordInput.validity.valueMissing).toBe(true);
  });

  it("prevents submission when the email is missing", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() => expect(mockLogin).not.toHaveBeenCalled());
    expect(emailInput.validity.valueMissing).toBe(true);
  });
});
