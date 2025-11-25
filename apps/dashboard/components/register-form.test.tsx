import type { Mock } from "vitest";

import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ConnectError, Code } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { RegisterForm } from "./register-form";

// eslint-disable-next-line no-var
var mockRouterPush: ReturnType<typeof vi.fn>;
// eslint-disable-next-line no-var
var toastSuccess: ReturnType<typeof vi.fn>;
// eslint-disable-next-line no-var
var toastError: ReturnType<typeof vi.fn>;

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("next/navigation", () => {
  mockRouterPush = vi.fn();
  return {
    useRouter: () => ({
      push: mockRouterPush,
    }),
  };
});

vi.mock("sonner", () => {
  toastSuccess = vi.fn();
  toastError = vi.fn();
  return {
    toast: {
      success: toastSuccess,
      error: toastError,
    },
  };
});

const mockRegister = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

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

    await user.type(screen.getByLabelText(/username/i), "johnny");
    await user.type(screen.getByLabelText(/email/i), "hello@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");

    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() =>
      expect(mockRegister).toHaveBeenCalledWith({
        username: "johnny",
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

    await user.type(screen.getByLabelText(/username/i), "takenuser");
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

    await user.type(screen.getByLabelText(/username/i), "johnny");
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

    await user.type(screen.getByLabelText(/username/i), "oopsuser");
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
