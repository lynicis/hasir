import {
  Repository,
  SDK,
} from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TransportProvider } from "@connectrpc/connect-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { RepositoryContext } from "@/lib/repository-context";

import RepositorySdkPreferencesContent from "./repository-sdk-preferences-content";

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

const mockUpdateSdkPreferences = vi.fn().mockResolvedValue({});

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateSdkPreferences: mockUpdateSdkPreferences,
  }),
}));

vi.mock("@connectrpc/connect-query", async () => {
  const actual = await vi.importActual("@connectrpc/connect-query");
  return {
    ...actual,
    useQuery: vi.fn(() => ({
      data: { id: "org-123" },
      isLoading: false,
    })),
  };
});

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
  let queryClient: QueryClient;
  const transport = createConnectTransport({
    baseUrl: "http://localhost:3000",
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <TransportProvider transport={transport}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TransportProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    mockUpdateSdkPreferences.mockClear();
    mockUpdateSdkPreferences.mockResolvedValue({});
  });

  it("renders the page heading and description", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
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
      </RepositoryContext.Provider>,
      { wrapper }
    );

    expect(screen.getByText("Go")).toBeInTheDocument();
    expect(screen.getByText("JavaScript/Typescript")).toBeInTheDocument();
  });

  it("renders Go SDK options", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    expect(screen.getByText("Protocol Buffers")).toBeInTheDocument();
    expect(screen.getByText("Connect-RPC")).toBeInTheDocument();
    expect(screen.getByText("gRPC")).toBeInTheDocument();
  });

  it("renders JavaScript SDK options", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    expect(screen.getByText("@bufbuild/es")).toBeInTheDocument();
    expect(screen.getByText("protocolbuffers/js")).toBeInTheDocument();
    expect(screen.getByText("@connectrpc/connectrpc")).toBeInTheDocument();
  });

  it("renders Save Configuration and Reset buttons", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    expect(
      screen.getByRole("button", { name: /save configuration/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /reset to defaults/i })
    ).toBeInTheDocument();
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
      </RepositoryContext.Provider>,
      { wrapper }
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
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const switches = screen.getAllByRole("switch");

    const goSwitch = switches[0];
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    expect(protocolBuffersSwitch).not.toBeDisabled();
  });

  it("throws error when used outside RepositoryLayout", () => {
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

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
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const switches = screen.getAllByRole("switch");

    const goSwitch = switches[0];
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    await user.click(protocolBuffersSwitch);

    expect(goSwitch).toBeChecked();
    expect(protocolBuffersSwitch).toBeChecked();

    const resetButton = screen.getByRole("button", {
      name: /reset to defaults/i,
    });
    await user.click(resetButton);

    expect(goSwitch).not.toBeChecked();
    expect(protocolBuffersSwitch).not.toBeChecked();
  });

  it("disables Save Configuration button when there are no changes", () => {
    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const saveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    expect(saveButton).toBeDisabled();
  });

  it("enables Save Configuration button when changes are made", async () => {
    const user = userEvent.setup();

    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const saveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    expect(saveButton).toBeDisabled();

    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0];
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();
  });

  it("disables Save Configuration button after successful save", async () => {
    const user = userEvent.setup();

    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const saveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    expect(saveButton).toBeDisabled();

    // Make a change
    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0];
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();

    // Save the changes
    await user.click(saveButton);

    // Wait for the save to complete
    await screen.findByRole("button", { name: /save configuration/i });

    // Button should be disabled again after save
    const updatedSaveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    expect(updatedSaveButton).toBeDisabled();
  });

  it("disables Save Configuration button after Reset to Defaults when no initial changes", async () => {
    const user = userEvent.setup();

    render(
      <RepositoryContext.Provider value={contextValue}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const saveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    expect(saveButton).toBeDisabled();

    // Make a change
    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0];
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();

    // Reset to defaults
    const resetButton = screen.getByRole("button", {
      name: /reset to defaults/i,
    });
    await user.click(resetButton);

    // Button should be disabled again after reset
    expect(saveButton).toBeDisabled();
  });

  it("keeps Save Configuration button enabled when changes exist after reset", async () => {
    const user = userEvent.setup();

    // Create a repository with existing SDK preferences
    const repositoryWithPreferences: Repository = {
      ...mockRepository,
      sdkPreferences: [{ sdk: SDK.SDK_GO_PROTOBUF, status: true }],
    } as Repository;

    const contextWithPreferences = {
      repository: repositoryWithPreferences,
      isLoading: false,
      error: null,
    };

    render(
      <RepositoryContext.Provider value={contextWithPreferences}>
        <RepositorySdkPreferencesContent />
      </RepositoryContext.Provider>,
      { wrapper }
    );

    const saveButton = screen.getByRole("button", {
      name: /save configuration/i,
    });
    // Initially disabled because no changes from server state
    expect(saveButton).toBeDisabled();

    // Make a change (toggle off the existing preference)
    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0];
    await user.click(goSwitch);

    // Button should be enabled because we made a change
    expect(saveButton).not.toBeDisabled();

    // Reset to defaults
    const resetButton = screen.getByRole("button", {
      name: /reset to defaults/i,
    });
    await user.click(resetButton);

    // Button should still be enabled because reset changed from server state
    expect(saveButton).not.toBeDisabled();
  });
});
