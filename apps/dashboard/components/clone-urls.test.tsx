import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, afterEach, it, vi, expect } from "bun:test";
import userEvent from "@testing-library/user-event";

import { CloneUrls } from "./clone-urls";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe("CloneUrls", () => {
  const defaultProps = {
    repositoryId: "123e4567-e89b-12d3-a456-426614174000",
  };

  const originalClipboard = navigator.clipboard;
  let mockWriteText: ReturnType<typeof mock>;

  beforeEach(() => {
    vi.clearAllMocks();
    delete process.env.NEXT_PUBLIC_API_URL;

    mockWriteText = vi.fn(() => Promise.resolve());
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: mockWriteText,
      },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    Object.defineProperty(navigator, "clipboard", {
      value: originalClipboard,
      writable: true,
      configurable: true,
    });
  });

  it("displays HTTPS URL by default using repository ID", () => {
    render(<CloneUrls {...defaultProps} />);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `http://localhost:8080/git/${defaultProps.repositoryId}.git`
    );
  });

  it("switches to SSH URL using repository ID", async () => {
    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `ssh://git@localhost:8080/${defaultProps.repositoryId}.git`
    );
  });

  it("copies URL to clipboard", async () => {
    render(<CloneUrls {...defaultProps} />);

    const copyButton = screen.getByRole("button", {
      name: "Copy to clipboard",
    });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith(
        `http://localhost:8080/git/${defaultProps.repositoryId}.git`
      );
    });
  });

  it("handles SSH URL without port", async () => {
    const originalEnv = process.env.NEXT_PUBLIC_API_URL;
    process.env.NEXT_PUBLIC_API_URL = "https://example.com";

    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `ssh://git@example.com/${defaultProps.repositoryId}.git`
    );

    process.env.NEXT_PUBLIC_API_URL = originalEnv;
  });

  it("does not show SSH guide link when HTTPS is selected", () => {
    render(<CloneUrls {...defaultProps} />);

    const link = screen.queryByRole("link", { name: /how to configure ssh/i });
    expect(link).not.toBeInTheDocument();
  });

  it("shows SSH guide link when SSH protocol is selected", async () => {
    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const link = screen.getByRole("link", { name: /how to configure ssh/i });
    expect(link).toBeInTheDocument();
  });

  it("SSH guide link points to /docs/ssh-configuration", async () => {
    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const link = screen.getByRole("link", { name: /how to configure ssh/i });
    expect(link).toHaveAttribute("href", "/docs/ssh-configuration");
  });
});
