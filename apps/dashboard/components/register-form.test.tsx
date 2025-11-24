import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ConnectError, Code } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { RegisterForm } from "./register-form";

const { mockRouterPush } = vi.hoisted(() => ({
  mockRouterPush: vi.fn(),
}));

const { toastSuccess, toastError } = vi.hoisted(() => ({
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

const mockRegister = vi.fn();
const mockedUseClient = vi.mocked(useClient);

describe("RegisterForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      register: mockRegister,
    } as never);
    mockRegister.mockReset();
    mockRouterPush.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  it("submits valid data, notifies success, and navigates to login", async () => {
    const user = userEvent.setup();

    mockRegister.mockResolvedValueOnce(undefined);

    render(<RegisterForm />);

    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() =>
      expect(mockRegister).toHaveBeenCalledWith({
        email: "hello@example.com",
        password: "password123",
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith(
        "Account successfully created. You may now log in."
      )
    );

    expect(toastError).not.toHaveBeenCalled();

    await waitFor(() => expect(mockRouterPush).toHaveBeenCalledWith("/login"), {
      timeout: 2000,
    });
  });

  it("surfaces an error when the email is already registered", async () => {
    const user = userEvent.setup();

    mockRegister.mockRejectedValueOnce(
      new ConnectError("already exists", Code.AlreadyExists)
    );

    render(<RegisterForm />);

    await user.type(screen.getByLabelText(/email/i), "taken@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(mockRegister).toHaveBeenCalled());
    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while creating your account."
      )
    );
    expect(toastSuccess).not.toHaveBeenCalled();
    expect(mockRouterPush).not.toHaveBeenCalled();
  });

  it("prevents submission when passwords do not match", async () => {
    const user = userEvent.setup();

    render(<RegisterForm />);

    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    const confirmInput = screen.getByLabelText(/confirm password/i);
    await user.type(confirmInput, "different123");

    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() =>
      expect(confirmInput).toHaveAttribute("aria-invalid", "true")
    );
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it("notifies when an unexpected error occurs", async () => {
    const user = userEvent.setup();

    mockRegister.mockRejectedValueOnce(new Error("network error"));

    render(<RegisterForm />);

    await user.type(screen.getByLabelText(/email/i), "oops@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith(
        "Error occurred while creating your account."
      )
    );
    expect(toastSuccess).not.toHaveBeenCalled();
    expect(mockRouterPush).not.toHaveBeenCalled();
  });
});
