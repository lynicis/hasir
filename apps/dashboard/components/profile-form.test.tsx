import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { useUserStore } from "@/stores/user-store-provider";

import { ProfileForm } from "./profile-form";

vi.mock("@/stores/user-store-provider", () => ({
  useUserStore: vi.fn(),
}));

const mockedUseUserStore = vi.mocked(useUserStore);

describe("ProfileForm", () => {
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    mockedUseUserStore.mockReturnValue({
      email: "test@example.com",
    } as never);
    mockOnSubmit.mockReset();
  });

  it("renders all form fields", () => {
    render(<ProfileForm onSubmit={mockOnSubmit} />);

    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("New Password")).toBeInTheDocument();
    expect(screen.getByLabelText("Confirm New Password")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /save changes/i })
    ).toBeInTheDocument();
  });

  it("displays user email from store", () => {
    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText("Email") as HTMLInputElement;
    expect(emailInput.value).toBe("test@example.com");
  });

  it("submits form with valid email only", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText("Email");
    await user.clear(emailInput);
    await user.type(emailInput, "newemail@example.com");
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalled();
      const callArgs = mockOnSubmit.mock.calls[0][0];
      expect(callArgs).toMatchObject({
        email: "newemail@example.com",
        password: "",
        confirmPassword: "",
      });
    });
  });

  it("submits form with valid password only", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    await user.type(screen.getByLabelText("New Password"), "newpassword123");
    await user.type(
      screen.getByLabelText("Confirm New Password"),
      "newpassword123"
    );
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalled();
      const callArgs = mockOnSubmit.mock.calls[0][0];
      expect(callArgs).toMatchObject({
        password: "newpassword123",
        confirmPassword: "newpassword123",
      });
    });
  });

  it("submits form with both email and password", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText("Email");
    await user.clear(emailInput);
    await user.type(emailInput, "updated@example.com");
    await user.type(screen.getByLabelText("New Password"), "newpass123");
    await user.type(
      screen.getByLabelText("Confirm New Password"),
      "newpass123"
    );
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalled();
      const callArgs = mockOnSubmit.mock.calls[0][0];
      expect(callArgs).toMatchObject({
        email: "updated@example.com",
        password: "newpass123",
        confirmPassword: "newpass123",
      });
    });
  });

  it("prevents submission when email is invalid", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText("Email");
    // Select all and replace with invalid email
    await user.click(emailInput);
    await user.keyboard("{Control>}a{/Control}");
    await user.type(emailInput, "invalid-email");
    // Trigger blur to validate
    await user.tab();
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(
      () => {
        expect(emailInput).toHaveAttribute("aria-invalid", "true");
      },
      { timeout: 3000 }
    );
    // Wait a bit to ensure onSubmit is not called
    await new Promise((resolve) => setTimeout(resolve, 200));
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it("prevents submission when password is too short", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const passwordInput = screen.getByLabelText("New Password");
    await user.type(passwordInput, "short");
    await user.type(screen.getByLabelText("Confirm New Password"), "short");
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(passwordInput).toHaveAttribute("aria-invalid", "true");
    });
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it("prevents submission when password is too long", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    const passwordInput = screen.getByLabelText("New Password");
    await user.type(passwordInput, "a".repeat(21));
    await user.type(
      screen.getByLabelText("Confirm New Password"),
      "a".repeat(21)
    );
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(passwordInput).toHaveAttribute("aria-invalid", "true");
    });
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it("prevents submission when passwords do not match", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    await user.type(screen.getByLabelText("New Password"), "password123");
    await user.type(
      screen.getByLabelText("Confirm New Password"),
      "different123"
    );
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      const confirmPasswordInput = screen.getByLabelText(
        "Confirm New Password"
      );
      expect(confirmPasswordInput).toHaveAttribute("aria-invalid", "true");
    });
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it("allows submission with empty optional fields", async () => {
    const user = userEvent.setup();

    render(<ProfileForm onSubmit={mockOnSubmit} />);

    // Form already has email from store, so we can submit with empty password fields
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(
      () => {
        expect(mockOnSubmit).toHaveBeenCalled();
        const callArgs = mockOnSubmit.mock.calls[0][0];
        expect(callArgs).toMatchObject({
          email: "test@example.com",
          password: "",
          confirmPassword: "",
        });
      },
      { timeout: 3000 }
    );
  });

  it("updates email field when user store email changes", () => {
    const { rerender } = render(<ProfileForm onSubmit={mockOnSubmit} />);

    mockedUseUserStore.mockReturnValue({
      email: "newemail@example.com",
    } as never);

    rerender(<ProfileForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText("Email") as HTMLInputElement;
    expect(emailInput.value).toBe("newemail@example.com");
  });
});
