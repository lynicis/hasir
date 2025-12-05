import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TransportProvider } from "@connectrpc/connect-query";
import userEvent from "@testing-library/user-event";

import { SshApiKeyPanel } from "./ssh-api-key-panel";

type MockKey = {
  id: string;
  name: string;
};

const { toastSuccess, toastError } = vi.hoisted(() => ({
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccess,
    error: toastError,
  },
}));

// Mock the query hooks
const mockSshKeys = vi.hoisted(() => ({
  data: { keys: [] as MockKey[], totalPage: 1 },
  isLoading: false,
  refetch: vi.fn(),
}));

const mockApiKeys = vi.hoisted(() => ({
  data: { keys: [] as MockKey[], totalPage: 1 },
  isLoading: false,
  refetch: vi.fn(),
}));

vi.mock("@connectrpc/connect-query", async () => {
  const actual = await vi.importActual("@connectrpc/connect-query");
  return {
    ...actual,
    useQuery: vi.fn((queryFn) => {
      // Determine which query is being called based on the query function
      if (
        queryFn.name === "getSshKeys" ||
        queryFn.toString().includes("SshKey")
      ) {
        return mockSshKeys;
      }
      if (
        queryFn.name === "getApiKeys" ||
        queryFn.toString().includes("ApiKey")
      ) {
        return mockApiKeys;
      }
      return mockSshKeys;
    }),
  };
});

// Mock the user client
const mockUserClient = {
  createSshKey: vi.fn(),
  revokeSshKey: vi.fn(),
  createApiKey: vi.fn(),
  revokeApiKey: vi.fn(),
};

vi.mock("@/lib/use-client", () => ({
  useClient: () => mockUserClient,
}));

describe("SshApiKeyPanel", () => {
  let clipboardWriteTextSpy: ReturnType<typeof vi.fn>;
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
    toastSuccess.mockReset();
    toastError.mockReset();

    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    // Reset mock data
    mockSshKeys.data = { keys: [] as MockKey[], totalPage: 1 };
    mockSshKeys.isLoading = false;
    mockSshKeys.refetch = vi.fn();

    mockApiKeys.data = { keys: [] as MockKey[], totalPage: 1 };
    mockApiKeys.isLoading = false;
    mockApiKeys.refetch = vi.fn();

    // Reset mock clients
    mockUserClient.createSshKey.mockReset();
    mockUserClient.revokeSshKey.mockReset();
    mockUserClient.createApiKey.mockReset();
    mockUserClient.revokeApiKey.mockReset();

    // Mock clipboard API
    clipboardWriteTextSpy = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: {
        writeText: clipboardWriteTextSpy,
      },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("Initial Rendering", () => {
    it("renders SSH and API sections with proper headings", () => {
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/ssh & api access/i)).toBeInTheDocument();
      expect(
        screen.getByRole("heading", { name: /ssh public keys/i })
      ).toBeInTheDocument();
      expect(
        screen.getByRole("heading", { name: /^api keys$/i, level: 3 })
      ).toBeInTheDocument();
    });

    it("shows empty state for SSH keys", () => {
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/no ssh keys yet/i)).toBeInTheDocument();
      expect(
        screen.getByText(
          /add your first ssh public key to enable secure authentication/i
        )
      ).toBeInTheDocument();
    });

    it("shows empty state for API keys", () => {
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/no api keys yet/i)).toBeInTheDocument();
      expect(
        screen.getByText(
          /generate your first api key to get started with programmatic access/i
        )
      ).toBeInTheDocument();
    });

    it("renders SSH key form with public key input", () => {
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByLabelText(/^public key$/i)).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /add ssh key/i })
      ).toBeInTheDocument();
    });

    it("disables add button when SSH key input is empty", () => {
      render(<SshApiKeyPanel />, { wrapper });

      const addButton = screen.getByRole("button", { name: /add ssh key/i });
      expect(addButton).toBeDisabled();
    });

    it("shows loading state for SSH keys", () => {
      mockSshKeys.isLoading = true;
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/loading ssh keys/i)).toBeInTheDocument();
    });

    it("shows loading state for API keys", () => {
      mockApiKeys.isLoading = true;
      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/loading api keys/i)).toBeInTheDocument();
    });
  });

  describe("SSH Key Management", () => {
    it("adds a new SSH key", async () => {
      const user = userEvent.setup({ delay: null });

      mockUserClient.createSshKey.mockResolvedValueOnce({});
      mockSshKeys.refetch.mockResolvedValueOnce({
        data: {
          keys: [{ id: "ssh-1", name: "SSH Key 1" }],
        },
      });

      render(<SshApiKeyPanel />, { wrapper });

      const nameInput = screen.getByLabelText(/^key name$/i, {
        selector: "#ssh-key-name",
      });
      const keyInput = screen.getByLabelText(/^public key$/i);
      const addButton = screen.getByRole("button", { name: /add ssh key/i });

      await user.type(nameInput, "SSH Key 1");
      await user.type(
        keyInput,
        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGq user@example.com"
      );

      await waitFor(() => {
        expect(addButton).toBeEnabled();
      });

      await user.click(addButton);

      await waitFor(() => {
        expect(mockUserClient.createSshKey).toHaveBeenCalledWith({
          name: "SSH Key 1",
          publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGq user@example.com",
        });
      });

      await waitFor(() => {
        expect(mockSshKeys.refetch).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(toastSuccess).toHaveBeenCalledWith(
          "SSH key added successfully",
          expect.any(Object)
        );
      });
    });

    it("displays SSH keys from query", () => {
      mockSshKeys.data = {
        keys: [
          { id: "ssh-1", name: "Work Laptop" },
          { id: "ssh-2", name: "Home Desktop" },
        ],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText("Work Laptop")).toBeInTheDocument();
      expect(screen.getByText("Home Desktop")).toBeInTheDocument();
      expect(screen.getByText(/key id: ssh-1/i)).toBeInTheDocument();
      expect(screen.getByText(/key id: ssh-2/i)).toBeInTheDocument();
    });

    it("deletes SSH key", async () => {
      const user = userEvent.setup({ delay: null });

      mockSshKeys.data = {
        keys: [{ id: "ssh-1", name: "Work Laptop" }],
        totalPage: 1,
      };

      mockUserClient.revokeSshKey.mockResolvedValueOnce({});
      mockSshKeys.refetch.mockResolvedValueOnce({
        data: { keys: [], totalPage: 1 },
      });

      render(<SshApiKeyPanel />, { wrapper });

      const deleteButton = screen.getByRole("button", {
        name: /delete ssh key/i,
      });

      await user.click(deleteButton);

      await waitFor(() => {
        expect(mockUserClient.revokeSshKey).toHaveBeenCalledWith({
          id: "ssh-1",
        });
      });

      await waitFor(() => {
        expect(mockSshKeys.refetch).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(toastSuccess).toHaveBeenCalledWith(
          "SSH key deleted",
          expect.any(Object)
        );
      });
    });

    it("clears form after adding SSH key", async () => {
      const user = userEvent.setup({ delay: null });

      mockUserClient.createSshKey.mockResolvedValueOnce({});
      mockSshKeys.refetch.mockResolvedValueOnce({
        data: { keys: [], totalPage: 1 },
      });

      render(<SshApiKeyPanel />, { wrapper });

      const nameInput = screen.getByLabelText(/^key name$/i, {
        selector: "#ssh-key-name",
      });
      const keyInput = screen.getByLabelText(/^public key$/i);

      await user.type(nameInput, "Test Key");
      await user.type(keyInput, "ssh-ed25519 AAAAC3 user@example.com");

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /add ssh key/i })
        ).toBeEnabled();
      });

      await user.click(screen.getByRole("button", { name: /add ssh key/i }));

      await waitFor(() => {
        expect(nameInput).toHaveValue("");
        expect(keyInput).toHaveValue("");
      });
    });
  });

  describe("API Key Management", () => {
    it("generates a new API key", async () => {
      const user = userEvent.setup({ delay: null });

      mockUserClient.createApiKey.mockResolvedValueOnce({
        key: "api_key_12345678",
      });

      mockApiKeys.refetch.mockResolvedValueOnce({
        data: {
          keys: [{ id: "api-1", name: "API Key 1" }],
        },
      });

      render(<SshApiKeyPanel />, { wrapper });

      const nameInput = screen.getByLabelText(/key name/i, {
        selector: "#api-key-name",
      });
      const generateButton = screen.getByRole("button", {
        name: /generate key/i,
      });

      await user.type(nameInput, "Production API");
      await user.click(generateButton);

      await waitFor(() => {
        expect(mockUserClient.createApiKey).toHaveBeenCalledWith({
          name: "Production API",
        });
      });

      await waitFor(() => {
        expect(mockApiKeys.refetch).toHaveBeenCalled();
      });
    });

    it("displays API keys from query", () => {
      mockApiKeys.data = {
        keys: [
          { id: "api-1", name: "Production API" },
          { id: "api-2", name: "Development Key" },
        ],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText("Production API")).toBeInTheDocument();
      expect(screen.getByText("Development Key")).toBeInTheDocument();
    });

    it("deletes API key", async () => {
      const user = userEvent.setup({ delay: null });

      mockApiKeys.data = {
        keys: [{ id: "api-1", name: "Production API" }],
        totalPage: 1,
      };

      mockUserClient.revokeApiKey.mockResolvedValueOnce({});
      mockApiKeys.refetch.mockResolvedValueOnce({
        data: { keys: [], totalPage: 1 },
      });

      render(<SshApiKeyPanel />, { wrapper });

      const deleteButton = screen.getByRole("button", {
        name: /delete api key/i,
      });

      await user.click(deleteButton);

      await waitFor(() => {
        expect(mockUserClient.revokeApiKey).toHaveBeenCalledWith({
          id: "api-1",
        });
      });

      await waitFor(() => {
        expect(mockApiKeys.refetch).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(toastSuccess).toHaveBeenCalledWith(
          "API key deleted",
          expect.any(Object)
        );
      });
    });

    it("shows full key for newly created API keys with copy button", async () => {
      const user = userEvent.setup({ delay: null });

      mockUserClient.createApiKey.mockResolvedValueOnce({
        key: "full-api-key-token-123",
      });

      mockApiKeys.refetch.mockImplementationOnce(async () => {
        mockApiKeys.data = {
          keys: [{ id: "api-1", name: "Production API" }],
          totalPage: 1,
        };
        return {
          data: mockApiKeys.data,
        };
      });

      render(<SshApiKeyPanel />, { wrapper });

      const keyNameInput = screen.getByLabelText(/^key name$/i, {
        selector: "#api-key-name",
      });
      const generateButton = screen.getByRole("button", {
        name: /generate key/i,
      });

      await user.type(keyNameInput, "Production API");
      await user.click(generateButton);

      // Wait for the newly created key to appear with the full key visible
      await waitFor(() => {
        expect(screen.getByText("Production API")).toBeInTheDocument();
        expect(screen.getByText("full-api-key-token-123")).toBeInTheDocument();
      });

      const copyButton = await screen.findByRole("button", {
        name: /copy api key/i,
      });
      expect(copyButton).toBeInTheDocument();

      expect(screen.getByText("New")).toBeInTheDocument();
    });

    it("hides API key after it is no longer newly created", async () => {
      mockApiKeys.data = {
        keys: [{ id: "api-1", name: "Production API" }],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Verify the API key name is displayed
      expect(screen.getByText("Production API")).toBeInTheDocument();

      // Verify the key is hidden with a security message
      expect(
        screen.getByText("Key is hidden for security")
      ).toBeInTheDocument();

      // Verify there is no copy button for existing keys
      expect(
        screen.queryByRole("button", { name: /copy api key/i })
      ).not.toBeInTheDocument();
    });
  });

  describe("Form Validation", () => {
    it("enables add button only when SSH key input has content", async () => {
      const user = userEvent.setup({ delay: null });

      render(<SshApiKeyPanel />, { wrapper });

      const nameInput = screen.getByLabelText(/^key name$/i, {
        selector: "#ssh-key-name",
      });
      const keyInput = screen.getByLabelText(/^public key$/i);
      const addButton = screen.getByRole("button", { name: /add ssh key/i });

      expect(addButton).toBeDisabled();

      await user.type(nameInput, "Test Key");
      await user.type(keyInput, "ssh-ed25519 AAAAC3");

      await waitFor(() => {
        expect(addButton).toBeEnabled();
      });
    });

    it("does not enable button with only whitespace in SSH key input", async () => {
      const user = userEvent.setup({ delay: null });

      render(<SshApiKeyPanel />, { wrapper });

      const keyInput = screen.getByLabelText(/^public key$/i);
      const addButton = screen.getByRole("button", { name: /add ssh key/i });

      await user.type(keyInput, "   ");

      await waitFor(() => {
        expect(addButton).toBeDisabled();
      });
    });
  });

  describe("Pagination", () => {
    it("does not show pagination when there is only one page", () => {
      mockSshKeys.data = {
        keys: [{ id: "ssh-1", name: "Key 1" }],
        totalPage: 1,
      };
      mockApiKeys.data = {
        keys: [{ id: "api-1", name: "Key 1" }],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Pagination should not be visible when totalPage is 1
      const paginationButtons = screen.queryAllByRole("button", {
        name: /^[0-9]+$/,
      });
      expect(paginationButtons.length).toBe(0);
    });

    it("shows pagination for SSH keys when there are multiple pages", () => {
      mockSshKeys.data = {
        keys: Array.from({ length: 10 }, (_, i) => ({
          id: `ssh-${i}`,
          name: `SSH Key ${i}`,
        })),
        totalPage: 3,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Should show page 1 button (current page)
      expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    });

    it("shows pagination for API keys when there are multiple pages", () => {
      mockApiKeys.data = {
        keys: Array.from({ length: 10 }, (_, i) => ({
          id: `api-${i}`,
          name: `API Key ${i}`,
        })),
        totalPage: 3,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Should show page 1 button (current page)
      expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    });

    it("hides pagination when loading SSH keys", () => {
      mockSshKeys.isLoading = true;
      mockSshKeys.data = {
        keys: [],
        totalPage: 3,
      };

      render(<SshApiKeyPanel />, { wrapper });

      const paginationButtons = screen.queryAllByRole("button", {
        name: /^[0-9]+$/,
      });
      expect(paginationButtons.length).toBe(0);
    });

    it("hides pagination when loading API keys", () => {
      mockApiKeys.isLoading = true;
      mockApiKeys.data = {
        keys: [],
        totalPage: 3,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Pagination should not be visible when loading
      const paginationButtons = screen.queryAllByRole("button", {
        name: /^[0-9]+$/,
      });
      expect(paginationButtons.length).toBe(0);
    });

    it("hides pagination when there are no SSH keys", () => {
      mockSshKeys.data = {
        keys: [],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      // Pagination should not be visible when there are no items
      expect(screen.getByText(/no ssh keys yet/i)).toBeInTheDocument();
      const paginationButtons = screen.queryAllByRole("button", {
        name: /^[0-9]+$/,
      });
      expect(paginationButtons.length).toBe(0);
    });

    it("hides pagination when there are no API keys", () => {
      mockApiKeys.data = {
        keys: [],
        totalPage: 1,
      };

      render(<SshApiKeyPanel />, { wrapper });

      expect(screen.getByText(/no api keys yet/i)).toBeInTheDocument();
      const paginationButtons = screen.queryAllByRole("button", {
        name: /^[0-9]+$/,
      });
      expect(paginationButtons.length).toBe(0);
    });
  });
});
