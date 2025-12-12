import { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import RepositorySdkPreferencesContent from "./repository-sdk-preferences-content";
import { RepositoryContext } from "./repository-context";

vi.mock("next/navigation", () => ({
  useParams: () => ({ repositoryId: "repo-123" }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateSdkPreferences: vi.fn().mockResolvedValue({}),
  }),
}));

const mockRepository: Repository = {
  $typeName: "registry.v1.Repository",
  id: "repo-123",
  name: "test-repo",
  sdkPreferences: [],
  organizationId: "org-123",
  visibility: 1,
  createdAt: undefined,
  updatedAt: undefined,
} as Repository;

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

    const switches = screen.getAllByRole("switch");

  
    const goSwitch = switches[0];
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    expect(protocolBuffersSwitch).not.toBeDisabled();
  });

  it("throws error when used outside RepositoryLayout", () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => {
      render(<RepositorySdkPreferencesContent />);
    }).toThrow("SdkPreferencesPage must be used within RepositoryLayout");

    consoleError.mockRestore();
  });

  it("disables all toggles when Reset to Defaults button is clicked", async () => {
    const user = userEvent.setup();

    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>
    );

    const switches = screen.getAllByRole("switch");

    const goSwitch = switches[0];
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    await user.click(protocolBuffersSwitch);

    expect(goSwitch).toBeChecked();
    expect(protocolBuffersSwitch).toBeChecked();

    const resetButton = screen.getByRole("button", { name: /reset to defaults/i });
    await user.click(resetButton);

    expect(goSwitch).not.toBeChecked();
    expect(protocolBuffersSwitch).not.toBeChecked();
  });
});
