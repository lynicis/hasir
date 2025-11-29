import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { LoginForm } from "./login-form";

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

const mockPush = vi.fn();
const mockFetch = vi.fn();

global.fetch = mockFetch;

describe("LoginForm", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockPush.mockReset();
  });

  it("submits credentials when the form is valid", async () => {
    const user = userEvent.setup();
    mockFetch.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve({
          user: { id: "123", email: "hello@example.com" },
          accessToken: "token",
        }),
    });

    render(<LoginForm />);

    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() =>
      expect(mockFetch).toHaveBeenCalledWith("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: "hello@example.com",
          password: "password123",
        }),
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

    await waitFor(() => expect(mockFetch).not.toHaveBeenCalled());
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
    expect(mockFetch).not.toHaveBeenCalled();
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
    expect(mockFetch).not.toHaveBeenCalled();
  });

  it("prevents submission when the password is missing", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(
      /password/i
    ) as HTMLInputElement;
    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() => expect(mockFetch).not.toHaveBeenCalled());
    expect(passwordInput.validity.valueMissing).toBe(true);
  });

  it("prevents submission when the email is missing", async () => {
    const user = userEvent.setup();

    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /login/i }));

    await waitFor(() => expect(mockFetch).not.toHaveBeenCalled());
    expect(emailInput.validity.valueMissing).toBe(true);
  });
});
