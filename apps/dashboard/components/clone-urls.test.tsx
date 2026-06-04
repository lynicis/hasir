import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
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

  beforeEach(() => {
    vi.clearAllMocks();
    delete process.env.NEXT_PUBLIC_API_URL;

    if (!navigator.clipboard) {
      Object.defineProperty(navigator, "clipboard", {
        value: {
          writeText: vi.fn().mockResolvedValue(undefined),
        },
        writable: true,
        configurable: true,
      });
    }
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
    const writeTextSpy = vi
      .spyOn(navigator.clipboard, "writeText")
      .mockResolvedValue(undefined);

    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const copyButton = screen.getByRole("button", {
      name: "Copy to clipboard",
    });
    await user.click(copyButton);

    await waitFor(() => {
      expect(writeTextSpy).toHaveBeenCalledWith(
        `http://localhost:8080/git/${defaultProps.repositoryId}.git`
      );
    });

    writeTextSpy.mockRestore();
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
