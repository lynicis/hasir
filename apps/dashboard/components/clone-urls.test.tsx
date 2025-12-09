import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import userEvent from "@testing-library/user-event";

import { CloneUrls } from "./clone-urls";

const writeTextMock = vi.fn();
Object.defineProperty(navigator, "clipboard", {
  value: {
    writeText: writeTextMock,
  },
  writable: true,
  configurable: true,
});

describe("CloneUrls", () => {
  const defaultProps = {
    repositoryName: "test-repo",
    repositoryId: "123e4567-e89b-12d3-a456-426614174000",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("displays HTTPS URL by default using repository ID", () => {
    render(<CloneUrls {...defaultProps} />);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(`http://localhost:8080/git/${defaultProps.repositoryId}.git`);
  });

  it("switches to SSH URL using repository name", async () => {
    const user = userEvent.setup();
    render(<CloneUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(`git@localhost:${defaultProps.repositoryName}.git`);
  });

  it("copies URL to clipboard", async () => {
    const writeTextSpy = vi.fn();
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: writeTextSpy,
      },
      writable: true,
      configurable: true,
    });

    render(<CloneUrls {...defaultProps} />);

    const copyButton = screen.getByRole("button", { name: "Copy to clipboard" });
    fireEvent.click(copyButton);

    expect(writeTextSpy).toHaveBeenCalledWith(
      `http://localhost:8080/git/${defaultProps.repositoryId}.git`
    );
  });
});
