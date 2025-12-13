import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SdkUrls } from "./sdk-urls";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const mockUseQuery = vi.fn();

vi.mock("@connectrpc/connect-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

vi.mock("@/lib/query-retry", () => ({
  customRetry: vi.fn(),
}));

describe("SdkUrls", () => {
  const defaultProps = {
    repositoryId: "repo-456",
  };

  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    vi.clearAllMocks();
    process.env.NEXT_PUBLIC_API_URL = "http://localhost:8080";
    mockUseQuery.mockReturnValue({
      data: { id: "org-123" },
      isLoading: false,
      error: null,
    });
  });

  afterEach(() => {
    process.env.NEXT_PUBLIC_API_URL = originalEnv;
  });

  it("displays HTTPS SDK URL by default with go-protobuf", () => {
    render(<SdkUrls {...defaultProps} />);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `http://localhost:8080/sdk/org-123/${defaultProps.repositoryId}/go-protobuf/`
    );
  });

  it("switches to SSH URL", async () => {
    const user = userEvent.setup();
    render(<SdkUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `ssh://git@localhost:8080/sdk/org-123/${defaultProps.repositoryId}/go-protobuf/`
    );
  });

  it("switches SDK type", async () => {
    const user = userEvent.setup();
    render(<SdkUrls {...defaultProps} />);

    const sdkTrigger = screen.getByRole("button", {
      name: /Go \/ Protocol Buffers/,
    });
    await user.click(sdkTrigger);

    const jsOption = await screen.findByRole("menuitem", {
      name: /JS \/ @bufbuild\/es/,
    });
    await user.click(jsOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `http://localhost:8080/sdk/org-123/${defaultProps.repositoryId}/js-bufbuild-es/`
    );
  });

  it("copies SDK URL to clipboard", async () => {
    const writeTextSpy = vi
      .spyOn(navigator.clipboard, "writeText")
      .mockResolvedValue(undefined);

    const user = userEvent.setup();
    render(<SdkUrls {...defaultProps} />);

    const copyButton = screen.getByRole("button", {
      name: "Copy SDK URL to clipboard",
    });
    await user.click(copyButton);

    expect(writeTextSpy).toHaveBeenCalledWith(
      `http://localhost:8080/sdk/org-123/${defaultProps.repositoryId}/go-protobuf/`
    );

    writeTextSpy.mockRestore();
  });

  it("handles SSH URL without port", async () => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.example.com";

    const user = userEvent.setup();
    render(<SdkUrls {...defaultProps} />);

    const trigger = screen.getByRole("button", { name: "HTTPS" });
    await user.click(trigger);

    const sshOption = await screen.findByRole("menuitem", { name: "SSH" });
    await user.click(sshOption);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `ssh://git@api.example.com/sdk/org-123/${defaultProps.repositoryId}/go-protobuf/`
    );
  });

  it("displays SDK Download URL label", () => {
    render(<SdkUrls {...defaultProps} />);

    expect(screen.getByText("SDK Download URL")).toBeInTheDocument();
  });

  it("calls getOrganizationId with correct repositoryId", () => {
    render(<SdkUrls {...defaultProps} />);

    expect(mockUseQuery).toHaveBeenCalledWith(
      expect.anything(),
      { repositoryId: defaultProps.repositoryId },
      expect.anything()
    );
  });

  it("shows loading skeleton when organization ID is loading", () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    render(<SdkUrls {...defaultProps} />);

    expect(screen.getByText("SDK Download URL")).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("shows loading skeleton when organization ID is not available", () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    render(<SdkUrls {...defaultProps} />);

    expect(screen.getByText("SDK Download URL")).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("handles all SDK types correctly", async () => {
    const user = userEvent.setup();
    render(<SdkUrls {...defaultProps} />);

    const sdkTrigger = screen.getByRole("button", {
      name: /Go \/ Protocol Buffers/,
    });
    await user.click(sdkTrigger);

    const sdkTypes = [
      { name: /Go \/ Connect-RPC/, expected: "go-connectrpc" },
      { name: /Go \/ gRPC/, expected: "go-grpc" },
      { name: /JS \/ @bufbuild\/es/, expected: "js-bufbuild-es" },
      { name: /JS \/ protocolbuffers/, expected: "js-protobuf" },
      { name: /JS \/ @connectrpc/, expected: "js-connectrpc" },
    ];

    for (const { name, expected } of sdkTypes) {
      const option = await screen.findByRole("menuitem", { name });
      await user.click(option);

      const input = screen.getByRole("textbox");
      expect(input).toHaveValue(
        `http://localhost:8080/sdk/org-123/${defaultProps.repositoryId}/${expected}/`
      );

      // Reopen dropdown for next selection
      await user.click(sdkTrigger);
    }
  });
});
