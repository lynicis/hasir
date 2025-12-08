import { describe, it, expect, vi, beforeEach } from "vitest";
import { NextRequest } from "next/server";

import { GET, POST } from "./route";

process.env.NEXT_PUBLIC_API_URL = "http://localhost:8080";

const mockGetSession = vi.fn();
const mockRefreshSession = vi.fn();
const mockSaveSession = vi.fn();

vi.mock("@/lib/session", () => ({
  getSession: () => mockGetSession(),
  isExpiredSeconds: vi.fn((exp?: number) => {
    if (!exp) return true;
    return Math.floor(Date.now() / 1000) > exp;
  }),
  isExpiredMillis: vi.fn((timestamp?: number) => {
    if (!timestamp) return true;
    return Date.now() > timestamp;
  }),
  refreshSession: (...args: unknown[]) => mockRefreshSession(...args),
  saveSession: (...args: unknown[]) => mockSaveSession(...args),
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
      refreshAt: 0,
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
      refreshAt: 0,
      destroy: vi.fn(),
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
      refreshAt: 0,
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
      refreshAt: 0,
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

describe("POST /api/auth/session", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const createValidTokens = () => {
    const futureTime = Math.floor(Date.now() / 1000) + 3600;
    const validJwtToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSIsImV4cCI6MTczMzU0NDYxMiwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo4MDgwIn0.signature";
    const validRefreshToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzM2MzEwMTIsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCJ9.signature";

    return {
      accessToken: validJwtToken,
      refreshToken: validRefreshToken,
      accessPayload: {
        sub: "user-123",
        email: "user@example.com",
        exp: futureTime,
        iss: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
      },
      refreshPayload: {
        exp: futureTime + 86400,
        iss: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
      },
    };
  };

  it("saves session and returns success when valid tokens provided", async () => {
    const { accessToken, refreshToken, accessPayload, refreshPayload } = createValidTokens();

    mockDecodeJwt
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce(refreshPayload);

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.success).toBe(true);
    expect(mockSaveSession).toHaveBeenCalledWith({
      user: {
        id: "user-123",
        email: "user@example.com",
      },
      accessToken,
      refreshToken,
      refreshAt: accessPayload.exp * 1000,
      expiresAt: refreshPayload.exp * 1000,
    });
  });

  it("returns 400 when request body is invalid", async () => {
    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        invalidField: "value",
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("Invalid request body");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when access token is not a valid JWT", async () => {
    const { refreshToken } = createValidTokens();
    const invalidJwt = "invalid-jwt";

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken: invalidJwt,
        newTokens: {
          accessToken: invalidJwt,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("Invalid request body");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when access token is expired", async () => {
    const { accessToken, refreshToken, accessPayload } = createValidTokens();
    const pastTime = Math.floor(Date.now() / 1000) - 3600;

    mockDecodeJwt
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce({
        sub: "user-123",
        email: "user@example.com",
        exp: pastTime,
        iss: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
      });

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("New access token expired");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when access token has invalid issuer", async () => {
    const { accessToken, refreshToken, accessPayload } = createValidTokens();
    const futureTime = Math.floor(Date.now() / 1000) + 3600;

    mockDecodeJwt
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce({
        sub: "user-123",
        email: "user@example.com",
        exp: futureTime,
        iss: "https://malicious.com",
      });

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("Invalid new access token issuer");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when refresh token is not a valid JWT", async () => {
    const { accessToken } = createValidTokens();
    const invalidJwt = "invalid-jwt";

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken: invalidJwt,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("Invalid request body");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when refresh token is expired", async () => {
    const { accessToken, refreshToken, accessPayload } = createValidTokens();
    const pastTime = Math.floor(Date.now() / 1000) - 3600;

    mockDecodeJwt
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce({
        exp: pastTime,
        iss: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
      });

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("New refresh token expired");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });

  it("returns 400 when refresh token has invalid issuer", async () => {
    const { accessToken, refreshToken, accessPayload } = createValidTokens();
    const futureTime = Math.floor(Date.now() / 1000) + 3600;

    mockDecodeJwt
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce({
        exp: futureTime,
        iss: "https://malicious.com",
      });

    const request = new NextRequest("http://localhost/api/auth/session", {
      method: "POST",
      body: JSON.stringify({
        accessToken,
        newTokens: {
          accessToken,
          refreshToken,
        },
      }),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json.error).toBe("Invalid new refresh token issuer");
    expect(mockSaveSession).not.toHaveBeenCalled();
  });
});
