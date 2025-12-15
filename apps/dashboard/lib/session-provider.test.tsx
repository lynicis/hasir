import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { useRouter, usePathname } from "next/navigation";
import { toast } from "sonner";

import { SessionProvider, useSession } from "./session-provider";

vi.mock("next/navigation", () => ({
  useRouter: vi.fn(),
  usePathname: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
  },
}));

global.fetch = vi.fn();

describe("SessionProvider", () => {
  const mockPush = vi.fn();
  const mockRouter = {
    push: mockPush,
  } as unknown;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useRouter).mockReturnValue(
      mockRouter as ReturnType<typeof useRouter>
    );
    vi.mocked(usePathname).mockReturnValue("/dashboard");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should provide loading state initially", () => {
    vi.mocked(global.fetch).mockImplementation(() => new Promise(() => {}));

    const TestComponent = () => {
      const { loading } = useSession();
      return <div>{loading ? "Loading" : "Not loading"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    expect(screen.getByText("Loading")).toBeInTheDocument();
  });

  it("should fetch and set session on mount", async () => {
    const mockSession = {
      user: { id: "123", email: "test@example.com" },
      accessToken: "token",
    };

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => mockSession,
    } as Response);

    const TestComponent = () => {
      const { session, loading } = useSession();
      if (loading) return <div>Loading</div>;
      return <div>Email: {session?.user?.email}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(screen.getByText("Email: test@example.com")).toBeInTheDocument();
    });

    expect(global.fetch).toHaveBeenCalledWith("/api/auth/session");
  });

  it("should set session to null when fetch returns non-ok", async () => {
    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: false,
      status: 401,
    } as Response);

    const TestComponent = () => {
      const { session, loading } = useSession();
      if (loading) return <div>Loading</div>;
      return <div>{session ? "Has session" : "No session"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(screen.getByText("No session")).toBeInTheDocument();
    });
  });

  it("should redirect to login when session fails on protected route", async () => {
    vi.mocked(usePathname).mockReturnValue("/dashboard");

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: false,
      status: 401,
    } as Response);

    const TestComponent = () => {
      const { loading } = useSession();
      return <div>{loading ? "Loading" : "Loaded"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("should not redirect to login on root path", async () => {
    vi.mocked(usePathname).mockReturnValue("/");

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: false,
      status: 401,
    } as Response);

    const TestComponent = () => {
      const { loading } = useSession();
      return <div>{loading ? "Loading" : "Loaded"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(screen.getByText("Loaded")).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("should not redirect to login on public paths", async () => {
    vi.mocked(usePathname).mockReturnValue("/login");

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: false,
      status: 401,
    } as Response);

    const TestComponent = () => {
      const { loading } = useSession();
      return <div>{loading ? "Loading" : "Loaded"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(screen.getByText("Loaded")).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("should handle fetch errors gracefully", async () => {
    vi.mocked(global.fetch).mockRejectedValueOnce(new Error("Network error"));

    const TestComponent = () => {
      const { session, loading } = useSession();
      if (loading) return <div>Loading</div>;
      return <div>{session ? "Has session" : "No session"}</div>;
    };

    render(
      <SessionProvider>
        <TestComponent />
      </SessionProvider>
    );

    await waitFor(() => {
      expect(screen.getByText("No session")).toBeInTheDocument();
      expect(toast.error).toHaveBeenCalledWith("Failed to fetch session");
    });
  });

  describe("refreshSession", () => {
    it("should refresh session successfully", async () => {
      const mockSession = {
        user: { id: "123", email: "test@example.com" },
        accessToken: "token",
      };

      vi.mocked(global.fetch)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockSession,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            ...mockSession,
            accessToken: "new-token",
          }),
        } as Response);

      const TestComponent = () => {
        const { session, loading, refreshSession } = useSession();

        if (loading) return <div>Loading</div>;

        return (
          <div>
            <div>Token: {session?.accessToken}</div>
            <button onClick={() => refreshSession()}>Refresh</button>
          </div>
        );
      };

      render(
        <SessionProvider>
          <TestComponent />
        </SessionProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Token: token")).toBeInTheDocument();
      });

      const button = screen.getByText("Refresh");
      button.click();

      await waitFor(() => {
        expect(screen.getByText("Token: new-token")).toBeInTheDocument();
      });
    });

    it("should handle refresh errors gracefully", async () => {
      const mockSession = {
        user: { id: "123", email: "test@example.com" },
        accessToken: "token",
      };

      vi.mocked(global.fetch)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockSession,
        } as Response)
        .mockRejectedValueOnce(new Error("Network error"));

      const TestComponent = () => {
        const { session, loading, refreshSession } = useSession();

        if (loading) return <div>Loading</div>;

        return (
          <div>
            <div>Token: {session?.accessToken || "none"}</div>
            <button onClick={() => refreshSession()}>Refresh</button>
          </div>
        );
      };

      render(
        <SessionProvider>
          <TestComponent />
        </SessionProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Token: token")).toBeInTheDocument();
      });

      const button = screen.getByText("Refresh");
      button.click();

      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith("Failed to refresh session");
      });

      await waitFor(() => {
        expect(screen.getByText("Token: none")).toBeInTheDocument();
      });
    });

    it("should set session to null when refresh returns non-ok", async () => {
      const mockSession = {
        user: { id: "123", email: "test@example.com" },
        accessToken: "token",
      };

      vi.mocked(global.fetch)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockSession,
        } as Response)
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
        } as Response);

      const TestComponent = () => {
        const { session, loading, refreshSession } = useSession();

        if (loading) return <div>Loading</div>;

        return (
          <div>
            <div>{session ? "Has session" : "No session"}</div>
            <button onClick={() => refreshSession()}>Refresh</button>
          </div>
        );
      };

      render(
        <SessionProvider>
          <TestComponent />
        </SessionProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Has session")).toBeInTheDocument();
      });

      const button = screen.getByText("Refresh");
      button.click();

      await waitFor(() => {
        expect(screen.getByText("No session")).toBeInTheDocument();
      });
    });
  });

  describe("useSession hook", () => {
    it("should use default context when used outside provider", () => {
      const TestComponent = () => {
        const { session, loading } = useSession();
        return (
          <div>
            <div>Session: {session ? "exists" : "null"}</div>
            <div>Loading: {loading ? "yes" : "no"}</div>
          </div>
        );
      };

      render(<TestComponent />);

      expect(screen.getByText("Session: null")).toBeInTheDocument();
      expect(screen.getByText("Loading: yes")).toBeInTheDocument();
    });
  });
});
