import type { ReactNode } from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { TransportProvider } from "@connectrpc/connect-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { RepositoryContext } from "@/components/repository-context";

import SdkPreferencesPage from "./page";

const mockParams = { repositoryId: "test-repo-id" };
const mockMutateAsync = vi.fn();
const mockRepositoryData = {
  id: "test-repo-id",
  name: "Test Repository",
  visibility: 0,
  sdkPreferences: [],
  $typeName: "registry.v1.Repository" as const,
};

vi.mock("next/navigation", () => ({
  useParams: () => mockParams,
}));

vi.mock("@connectrpc/connect-query", async () => {
  const actual = await vi.importActual("@connectrpc/connect-query");
  return {
    ...actual,
    useMutation: () => ({
      mutateAsync: mockMutateAsync,
      isPending: false,
    }),
    useQuery: vi.fn(() => ({
      data: { id: "org-123" },
      isLoading: false,
    })),
  };
});

vi.mock("@/lib/use-client", () => ({
  useClient: () => ({
    updateSdkPreferences: vi.fn().mockResolvedValue({}),
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

function RepositoryContextWrapper({
  children,
  repository = mockRepositoryData,
  isLoading = false,
  error = null,
}: {
  children: ReactNode;
  repository?: typeof mockRepositoryData;
  isLoading?: boolean;
  error?: unknown;
}) {
  return (
    <RepositoryContext.Provider value={{ repository, isLoading, error }}>
      {children}
    </RepositoryContext.Provider>
  );
}

describe("SdkPreferencesPage", () => {
  let queryClient: QueryClient;
  const transport = createConnectTransport({
    baseUrl: "http://localhost:3000",
  });

  const wrapper = ({ children }: { children: ReactNode }) => (
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
  });

  function renderWithContext(
    repository = mockRepositoryData,
    isLoading = false,
    error = null
  ) {
    return render(
      <RepositoryContextWrapper
        repository={repository}
        isLoading={isLoading}
        error={error}
      >
        <SdkPreferencesPage />
      </RepositoryContextWrapper>,
      { wrapper }
    );
  }

  it("renders SDK preferences page with all SDK options", () => {
    renderWithContext();

    expect(screen.getByText("SDK Generation Preferences")).toBeInTheDocument();
    expect(
      screen.getByText("Configure which SDKs to generate for this repository")
    ).toBeInTheDocument();

    expect(screen.getByText("Go")).toBeInTheDocument();
    expect(screen.getByText("JavaScript/Typescript")).toBeInTheDocument();

    expect(screen.getByText("Protocol Buffers")).toBeInTheDocument();
    expect(screen.getByText("Connect-RPC")).toBeInTheDocument();
    expect(screen.getByText("gRPC")).toBeInTheDocument();

    expect(screen.getByText("@bufbuild/es")).toBeInTheDocument();
    expect(screen.getByText("protocolbuffers/js")).toBeInTheDocument();
    expect(screen.getByText("@connectrpc/connectrpc")).toBeInTheDocument();
  });

  it("enables language when toggling main switch", async () => {
    const user = userEvent.setup();
    renderWithContext();

    const allSwitches = screen.getAllByRole("switch");
    const goSwitch = allSwitches[0];
    await user.click(goSwitch);

    const protocolBuffersSwitch = screen.getByLabelText("Protocol Buffers");
    expect(protocolBuffersSwitch).not.toBeDisabled();
  });

  it("disables sub-options when language is disabled", async () => {
    const user = userEvent.setup();
    renderWithContext();

    const allSwitches = screen.getAllByRole("switch");
    const goSwitch = allSwitches[0];

    const protocolBuffersSwitch = screen.getByLabelText("Protocol Buffers");
    expect(protocolBuffersSwitch).toBeDisabled();

    await user.click(goSwitch);
    expect(protocolBuffersSwitch).not.toBeDisabled();

    await user.click(goSwitch);
    expect(protocolBuffersSwitch).toBeDisabled();
  });

  it("enables language when a sub-option is toggled on", async () => {
    const user = userEvent.setup();
    renderWithContext();

    const protocolBuffersLabel = screen.getByText("Protocol Buffers");
    const protocolBuffersSwitch = protocolBuffersLabel
      .closest("div")
      ?.querySelector("button") as HTMLElement;

    expect(protocolBuffersSwitch).toBeDisabled();

    const goSwitch = screen.getAllByRole("switch")[0];
    await user.click(goSwitch);

    await user.click(protocolBuffersSwitch);

    const goSwitchChecked = goSwitch.getAttribute("data-state");
    expect(goSwitchChecked).toBe("checked");
  });

  it("shows save and reset buttons", () => {
    renderWithContext();

    expect(
      screen.getByRole("button", { name: /save configuration/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /reset to defaults/i })
    ).toBeInTheDocument();
  });

  it("displays SDK download section", () => {
    renderWithContext();

    expect(screen.getByText("Download SDK")).toBeInTheDocument();
    expect(
      screen.getByText(/Copy the SDK URL to install the generated SDK package/i)
    ).toBeInTheDocument();
  });

  it("resets configuration when Reset button is clicked", async () => {
    const user = userEvent.setup();
    renderWithContext();

    const goSwitch = screen.getAllByRole("switch")[0];
    await user.click(goSwitch);

    expect(goSwitch.getAttribute("data-state")).toBe("checked");

    const resetButton = screen.getByRole("button", {
      name: /reset to/i,
    });
    await user.click(resetButton);

    expect(goSwitch.getAttribute("data-state")).toBe("unchecked");
  });

  it("shows all preferences disabled when SDK preferences array is empty", () => {
    renderWithContext();

    const allSwitches = screen.getAllByRole("switch");

    const goSwitch = allSwitches[0];
    const jsSwitch = allSwitches[1];

    expect(goSwitch.getAttribute("data-state")).toBe("unchecked");
    expect(jsSwitch.getAttribute("data-state")).toBe("unchecked");

    const protocolBuffersSwitch = screen.getByLabelText("Protocol Buffers");
    expect(protocolBuffersSwitch.getAttribute("data-state")).toBe("unchecked");
  });

  it("shows all preferences disabled when SDK preferences is undefined", () => {
    const mockData = {
      id: "test-repo-id",
      name: "Test Repository",
      visibility: 0,
      $typeName: "registry.v1.Repository" as const,
      // sdkPreferences is undefined
    };

    renderWithContext(mockData as typeof mockRepositoryData);

    const allSwitches = screen.getAllByRole("switch");
    const goSwitch = allSwitches[0];
    const jsSwitch = allSwitches[1];

    expect(goSwitch.getAttribute("data-state")).toBe("unchecked");
    expect(jsSwitch.getAttribute("data-state")).toBe("unchecked");
  });
});
