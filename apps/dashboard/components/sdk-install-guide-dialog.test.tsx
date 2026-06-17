import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SdkInstallGuideDialog } from "./sdk-install-guide-dialog";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe("SdkInstallGuideDialog", () => {
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

  it("renders the trigger button", () => {
    render(<SdkInstallGuideDialog {...defaultProps} />);
    expect(
      screen.getByRole("button", { name: "View SDK installation guide" })
    ).toBeInTheDocument();
  });

  it("opens the dialog and shows default Go SDK content when clicked", async () => {
    const user = userEvent.setup();
    render(<SdkInstallGuideDialog {...defaultProps} />);

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    await user.click(trigger);

    expect(screen.getByText("SDK Installation Guide")).toBeInTheDocument();
    expect(screen.getByText("Go SDK")).toBeInTheDocument();
    expect(screen.getByText("JavaScript / TypeScript")).toBeInTheDocument();

    // Go SDK content should be visible by default
    expect(screen.getByText(/1. Configure Go Environment/i)).toBeInTheDocument();
    expect(screen.getByText(/2. Get the package/i)).toBeInTheDocument();
    expect(screen.getByText(/3. Go Module Replacement/i)).toBeInTheDocument();

    // Verify Go env content
    expect(screen.getByText(/go env -w GOINSECURE=localhost/i)).toBeInTheDocument();
    expect(screen.getByText(/go get localhost\/sdk\/org-123\/repo-456\/commit-abc\/go-connectrpc/i)).toBeInTheDocument();
    expect(screen.getByText(/replace localhost\/sdk\/org-123\/repo-456\/commit-abc\/go-connectrpc => hasir v0.0.0-commit-abc/i)).toBeInTheDocument();
  });

  it("switches to JS SDK tab and shows npm install commands", async () => {
    const user = userEvent.setup();
    render(<SdkInstallGuideDialog {...defaultProps} />);

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    await user.click(trigger);

    const jsTab = screen.getByRole("tab", { name: "JavaScript / TypeScript" });
    await user.click(jsTab);

    expect(screen.getByText(/Install via HTTP/i)).toBeInTheDocument();
    expect(screen.getByText(/Install via SSH/i)).toBeInTheDocument();

    expect(
      screen.getByText(/npm install git\+http:\/\/localhost:8080\/sdk\/org-123\/repo-456\/commit-abc\/js-connectrpc\//i)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/npm install git\+ssh:\/\/git@localhost:2222\/sdk\/org-123\/repo-456\/commit-abc\/js-connectrpc\//i)
    ).toBeInTheDocument();
  });

  it("copies Go env commands to clipboard", () => {
    render(<SdkInstallGuideDialog {...defaultProps} />);

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    fireEvent.click(trigger);

    const copyEnvButton = screen.getByRole("button", {
      name: "Copy Go env commands",
    });
    fireEvent.click(copyEnvButton);

    expect(mockWriteText).toHaveBeenCalledWith(
      "go env -w GOINSECURE=localhost\ngo env -w GONOSUMDB=localhost"
    );
  });

  it("copies Go get command to clipboard", () => {
    render(<SdkInstallGuideDialog {...defaultProps} />);

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    fireEvent.click(trigger);

    const copyGetButton = screen.getByRole("button", {
      name: "Copy go get command",
    });
    fireEvent.click(copyGetButton);

    expect(mockWriteText).toHaveBeenCalledWith(
      "go get localhost/sdk/org-123/repo-456/commit-abc/go-connectrpc"
    );
  });

  it("copies Go replace command to clipboard", () => {
    render(<SdkInstallGuideDialog {...defaultProps} />);

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    fireEvent.click(trigger);

    const copyReplaceButton = screen.getByRole("button", {
      name: "Copy Go replace command",
    });
    fireEvent.click(copyReplaceButton);

    expect(mockWriteText).toHaveBeenCalledWith(
      "replace localhost/sdk/org-123/repo-456/commit-abc/go-connectrpc => hasir v0.0.0-commit-abc"
    );
  });

  it("copies NPM HTTP and SSH install commands to clipboard", async () => {
    const user = userEvent.setup();
    render(<SdkInstallGuideDialog {...defaultProps} />);

    // Re-apply mock immediately after userEvent.setup() to override its stub
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: mockWriteText,
      },
      writable: true,
      configurable: true,
    });

    const trigger = screen.getByRole("button", {
      name: "View SDK installation guide",
    });
    await user.click(trigger);

    const jsTab = screen.getByRole("tab", { name: "JavaScript / TypeScript" });
    await user.click(jsTab);

    const copyHttpButton = screen.getByRole("button", {
      name: "Copy NPM HTTP install command",
    });
    fireEvent.click(copyHttpButton);

    expect(mockWriteText).toHaveBeenCalledWith(
      "npm install git+http://localhost:8080/sdk/org-123/repo-456/commit-abc/js-connectrpc/"
    );

    const copySshButton = screen.getByRole("button", {
      name: "Copy NPM SSH install command",
    });
    fireEvent.click(copySshButton);

    expect(mockWriteText).toHaveBeenLastCalledWith(
      "npm install git+ssh://git@localhost:2222/sdk/org-123/repo-456/commit-abc/js-connectrpc/"
    );
  });
});
