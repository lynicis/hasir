import { describe, it, expect, vi, beforeEach } from "vitest";

import { GET } from "./route";

const mockGetSession = vi.fn();
const mockRefreshSession = vi.fn();

vi.mock("@/lib/session", () => ({
  getSession: () => mockGetSession(),
  refreshSession: (...args: unknown[]) => mockRefreshSession(...args),
}));

vi.mock("@buf/hasir_hasir.bufbuild_es/user/v1/user_pb", () => ({
  UserService: {},
}));

vi.mock("@connectrpc/connect-web", () => ({
  createConnectTransport: vi.fn(() => ({})),
}));

const mockRenewTokens = vi.fn();

vi.mock("@connectrpc/connect", () => ({
  createClient: vi.fn(() => ({
    renewTokens: (...args: unknown[]) => mockRenewTokens(...args),
  })),
}));

const mockDecodeJwt = vi.fn();

vi.mock("jose", () => ({
  decodeJwt: (token: string) => mockDecodeJwt(token),
}));

describe("GET /api/auth/session", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns 401 when no user in session", async () => {
    mockGetSession.mockResolvedValueOnce({
      user: undefined,
      destroy: vi.fn(),
    });

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(401);
    expect(json.user).toBeNull();
  });

  it("destroys session and returns 401 when session is expired", async () => {
    const destroy = vi.fn();
    mockGetSession.mockResolvedValueOnce({
      user: { id: "id", email: "e" },
      expiresAt: Date.now() - 1000,
      destroy,
    });

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(401);
    expect(json.user).toBeNull();
    expect(destroy).toHaveBeenCalled();
  });

  it("returns session user and accessToken when valid and no refresh needed", async () => {
    mockGetSession.mockResolvedValueOnce({
      user: { id: "id", email: "e" },
      accessToken: "access-token",
      expiresAt: Date.now() + 1000,
      refreshAt: Date.now() + 1000,
    });

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.user).toEqual({ id: "id", email: "e" });
    expect(json.accessToken).toBe("access-token");
  });

  it("renews access token when refreshAt has passed and session valid", async () => {
    const session = {
      user: { id: "id", email: "e" },
      accessToken: "old-access-token",
      refreshToken: "refresh-token",
      expiresAt: Date.now() + 10_000,
      refreshAt: Date.now() - 1000,
    };

    mockGetSession.mockResolvedValueOnce(session);

    mockRenewTokens.mockResolvedValueOnce({
      accessToken: "new-access-token",
    });

    mockDecodeJwt.mockReturnValueOnce({
      exp: Math.floor((Date.now() + 20_000) / 1000),
    });

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(mockRenewTokens).toHaveBeenCalledWith({
      refreshToken: "refresh-token",
    });
    expect(session.accessToken).toBe("new-access-token");
    expect(session.refreshAt).toBeGreaterThan(Date.now());
    expect(mockRefreshSession).toHaveBeenCalledWith(session);
    expect(json.user).toEqual({ id: "id", email: "e" });
    expect(json.accessToken).toBe("new-access-token");
  });

  it("destroys session and returns 401 when refresh token missing but refresh needed", async () => {
    const destroy = vi.fn();
    mockGetSession.mockResolvedValueOnce({
      user: { id: "id", email: "e" },
      accessToken: "old-access-token",
      expiresAt: Date.now() + 10_000,
      refreshAt: Date.now() - 1000,
      refreshToken: undefined,
      destroy,
    });

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(401);
    expect(json.user).toBeNull();
    expect(destroy).toHaveBeenCalled();
  });

  it("destroys session and returns 401 when renewTokens throws", async () => {
    const destroy = vi.fn();
    mockGetSession.mockResolvedValueOnce({
      user: { id: "id", email: "e" },
      accessToken: "old-access-token",
      expiresAt: Date.now() + 10_000,
      refreshAt: Date.now() - 1000,
      refreshToken: "refresh-token",
      destroy,
    });

    mockRenewTokens.mockRejectedValueOnce(new Error("failed"));

    const response = await GET();
    const json = await response.json();

    expect(response.status).toBe(401);
    expect(json.user).toBeNull();
    expect(destroy).toHaveBeenCalled();
  });
});
