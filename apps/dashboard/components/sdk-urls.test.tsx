import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SdkUrls } from "./sdk-urls";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe("SdkUrls", () => {
  const defaultProps = {
    organizationId: "org-123",
    repositoryId: "repo-456",
    commitHash: "commit-abc",
  };

  const originalEnv = process.env.NEXT_PUBLIC_API_URL;
  const originalClipboard = navigator.clipboard;
  let mockWriteText: ReturnType<typeof mock>;

  beforeEach(() => {
    vi.clearAllMocks();
    process.env.NEXT_PUBLIC_API_URL = "http://localhost:8080";

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
    process.env.NEXT_PUBLIC_API_URL = originalEnv;
    Object.defineProperty(navigator, "clipboard", {
      value: originalClipboard,
      writable: true,
      configurable: true,
    });
  });

  it("displays HTTPS SDK URL by default with go-protobuf", () => {
    render(<SdkUrls {...defaultProps} />);

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue(
      `http://localhost:8080/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/go-protobuf/`
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
      `ssh://git@localhost:8080/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/go-protobuf/`
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
      `http://localhost:8080/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/js-bufbuild-es/`
    );
  });

  it("copies SDK URL to clipboard", async () => {
    render(<SdkUrls {...defaultProps} />);

    const copyButton = screen.getByRole("button", {
      name: "Copy SDK URL to clipboard",
    });
    fireEvent.click(copyButton);

    expect(mockWriteText).toHaveBeenCalledWith(
      `http://localhost:8080/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/go-protobuf/`
    );
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
      `ssh://git@api.example.com/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/go-protobuf/`
    );
  });

  it("displays SDK Download URL label", () => {
    render(<SdkUrls {...defaultProps} />);

    expect(screen.getByText("SDK Download URL")).toBeInTheDocument();
  });

  it("shows empty URL when organization ID is not provided", () => {
    render(
      <SdkUrls
        organizationId=""
        repositoryId={defaultProps.repositoryId}
        commitHash={defaultProps.commitHash}
      />
    );

    const input = screen.getByRole("textbox");
    expect(input).toHaveValue("http://localhost:8080");
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
        `http://localhost:8080/sdk/${defaultProps.organizationId}/${defaultProps.repositoryId}/${defaultProps.commitHash}/${expected}/`
      );

      await user.click(sdkTrigger);
    }
  });
});
