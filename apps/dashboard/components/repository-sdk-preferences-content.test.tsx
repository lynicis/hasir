import {
  type Repository,
  SDK,
} from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import * as connectQueryActual from "@connectrpc/connect-query";
import { TransportProvider } from "@connectrpc/connect-query";
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

vi.mock("@connectrpc/connect-query", () => ({
  ...connectQueryActual,
  useQuery: vi.fn(() => ({
    data: { id: "org-123" },
    isLoading: false,
  })),
}));

const mockRepository: Repository = {
  $typeName: "registry.v1.Repository",
  id: "repo-123",
  name: "test-repo",
  sdkPreferences: [],
  organizationId: "org-123",
  visibility: 1,
  managedByBuf: false,
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

    expect(screen.getAllByText("Protocol Buffers")[0]).toBeInTheDocument();
    expect(screen.getAllByText("Connect-RPC")[0]).toBeInTheDocument();
    expect(screen.getAllByText("gRPC")[0]).toBeInTheDocument();
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

    const goSwitch = switches[0]!;
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText(/protocol buffers/i);
    expect(protocolBuffersSwitch).not.toBeDisabled();
  });

  it("throws error when used outside RepositoryLayout", () => {
    const originalError = console.error;
    console.error = vi.fn();

    expect(() => {
      render(<RepositorySdkPreferencesContent />);
    }).toThrow("SdkPreferencesPage must be used within RepositoryLayout");

    console.error = originalError;
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

    const goSwitch = switches[0]!;
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
    const goSwitch = switches[0]!;
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

    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0]!;
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();

    await user.click(saveButton);
    await screen.findByRole("button", { name: /save configuration/i });

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

    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0]!;
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();

    const resetButton = screen.getByRole("button", {
      name: /reset to defaults/i,
    });
    await user.click(resetButton);

    expect(saveButton).toBeDisabled();
  });

  it("keeps Save Configuration button enabled when changes exist after reset", async () => {
    const user = userEvent.setup();

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
    expect(saveButton).toBeDisabled();

    const switches = screen.getAllByRole("switch");
    const goSwitch = switches[0]!;
    await user.click(goSwitch);

    expect(saveButton).not.toBeDisabled();

    const resetButton = screen.getByRole("button", {
      name: /reset to defaults/i,
    });
    await user.click(resetButton);

    expect(saveButton).toBeDisabled();
  });

  describe("when repository is managed by Buf", () => {
    const managedRepository: Repository = {
      ...mockRepository,
      managedByBuf: true,
    } as Repository;

    const managedContextValue = {
      repository: managedRepository,
      isLoading: false,
      error: null,
    };

    it("shows the managed by Buf alert banner", () => {
      render(
        <RepositoryContext.Provider value={managedContextValue}>
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Managed by Buf")).toBeInTheDocument();
      expect(
        screen.getByText(
          /sdk generation preferences are configured externally and cannot be modified here/i
        )
      ).toBeInTheDocument();
    });

    it("disables all language toggle switches", () => {
      render(
        <RepositoryContext.Provider value={managedContextValue}>
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      const switches = screen.getAllByRole("switch");

      for (const switchEl of switches) {
        expect(switchEl).toBeDisabled();
      }
    });

    it("does not render Save Configuration or Reset to Defaults buttons", () => {
      render(
        <RepositoryContext.Provider value={managedContextValue}>
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      expect(
        screen.queryByRole("button", { name: /save configuration/i })
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /reset to defaults/i })
      ).not.toBeInTheDocument();
    });

    it("still renders the Download SDK section", () => {
      render(
        <RepositoryContext.Provider value={managedContextValue}>
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      expect(screen.getByText("Download SDK")).toBeInTheDocument();
    });

    it("displays existing SDK preferences as read-only", () => {
      const managedWithPreferences: Repository = {
        ...mockRepository,
        managedByBuf: true,
        sdkPreferences: [
          { sdk: SDK.SDK_GO_PROTOBUF, status: true },
          { sdk: SDK.SDK_GO_CONNECTRPC, status: true },
        ],
      } as Repository;

      render(
        <RepositoryContext.Provider
          value={{
            repository: managedWithPreferences,
            isLoading: false,
            error: null,
          }}
        >
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      const switches = screen.getAllByRole("switch");
      const goSwitch = switches[0]!;

      expect(goSwitch).toBeChecked();
      expect(goSwitch).toBeDisabled();
    });

    it("does not show alert banner when managedByBuf is false", () => {
      render(
        <RepositoryContext.Provider value={contextValue}>
          <RepositorySdkPreferencesContent />
        </RepositoryContext.Provider>,
        { wrapper }
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByText("Managed by Buf")).not.toBeInTheDocument();
    });
  });
});
