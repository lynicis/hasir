import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import RepositorySdkPreferencesContent from "./repository-sdk-preferences-content";
import { RepositoryContext } from "./repository-context";

vi.mock("next/navigation", () => ({
  useParams: () => ({ repositoryId: "repo-123" }),
}));

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateSdkPreferences: vi.fn().mockResolvedValue({}),
  }),
}));

const mockRepository = {
  id: "repo-123",
  name: "test-repo",
  sdkPreferences: [],
} as const;

const contextValue = {
  repository: mockRepository,
  isLoading: false,
  error: null,
};

describe("RepositorySdkPreferencesContent", () => {
  it("renders the page heading and description", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(
      screen.getByRole("heading", { name: /sdk generation preferences/i })
    ).toBeInTheDocument();
    expect(
      screen.getByText(/configure which sdks to generate for this repository/i)
    ).toBeInTheDocument();
  });

  it("renders language cards for Go and JavaScript", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("Go")).toBeInTheDocument();
    expect(screen.getByText("JavaScript/Typescript")).toBeInTheDocument();
  });

  it("renders Go SDK options", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("Protocol Buffers")).toBeInTheDocument();
    expect(screen.getByText("Connect-RPC")).toBeInTheDocument();
    expect(screen.getByText("gRPC")).toBeInTheDocument();
  });

  it("renders JavaScript SDK options", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("@bufbuild/es")).toBeInTheDocument();
    expect(screen.getByText("protocolbuffers/js")).toBeInTheDocument();
    expect(screen.getByText("@connectrpc/connectrpc")).toBeInTheDocument();
  });

  it("renders Save Configuration and Reset buttons", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(
      screen.getByRole("button", { name: /save configuration/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /reset to defaults/i })
    ).toBeInTheDocument();
  });

  it("displays the available SDK targets information", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText(/available sdk targets:/i)).toBeInTheDocument();
  });

  it("shows skeleton loading state when repository is loading", () => {
    const loadingContextValue = {
      repository: undefined,
      isLoading: true,
      error: null,
    };

    const { container } = render(
      <RepositoryContext.Provider value={loadingContextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("SDK Generation Preferences")).toBeInTheDocument();

    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("enables sub-options when language is toggled on", async () => {
    const user = userEvent.setup();

    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    // Get all switches (one for Go, one for JavaScript, plus sub-options)
    const switches = screen.getAllByRole("switch");

    // The first switch should be for the Go language toggle
    const goSwitch = switches[0];
    await user.click(goSwitch);

    // After enabling Go, the sub-option switches should be enabled
    // We can verify the state by checking if switches are not disabled
    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    expect(protocolBuffersSwitch).not.toBeDisabled();
  });

  it("throws error when used outside RepositoryLayout", () => {
    // Suppress console.error for this test
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => {
      render(<RepositorySdkPreferencesContent />);
    }).toThrow("SdkPreferencesPage must be used within RepositoryLayout");

    consoleError.mockRestore();
  });
});
