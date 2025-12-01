import { describe, it, expect, vi, beforeEach } from "vitest";
import { NextRequest } from "next/server";

import { POST } from "./route";

type SessionInput = {
  user: {
    id: string;
    email: string;
  };
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  refreshAt: number;
};

const mockSaveSession = vi.fn<(input: SessionInput) => void>();

vi.mock("@/lib/session", () => ({
  saveSession: (input: SessionInput) => mockSaveSession(input),
}));

vi.mock("@buf/hasir_hasir.bufbuild_es/user/v1/user_pb", () => ({
  UserService: {},
}));

vi.mock("@connectrpc/connect-web", () => ({
  createConnectTransport: vi.fn(() => ({})),
}));

type LoginInput = {
  email: string;
  password: string;
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const mockLogin = vi.fn(async (_input: LoginInput) => ({
  accessToken: "access-token",
  refreshToken: "refresh-token",
}));

vi.mock("@connectrpc/connect", () => ({
  createClient: vi.fn(() => ({
    login: (input: LoginInput) => mockLogin(input),
  })),
}));

const mockDecodeJwt = vi.fn((token: string) => {
  if (token === "access-token") {
    return { sub: "user-id", email: "user@example.com", exp: 1111 };
  }
  return { exp: 2222 };
});

vi.mock("jose", () => ({
  decodeJwt: (token: string) => mockDecodeJwt(token),
}));

describe("POST /api/auth/login", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("creates a session and returns user and access token on success", async () => {
    const body = { email: "user@example.com", password: "password" };

    const request = new NextRequest("http://localhost/api/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(200);
    expect(json.user).toEqual({
      id: "user-id",
      email: "user@example.com",
    });
    expect(json.accessToken).toBe("access-token");
    expect(mockLogin).toHaveBeenCalledWith({
      email: "user@example.com",
      password: "password",
    });
    expect(mockSaveSession).toHaveBeenCalledWith({
      user: {
        id: "user-id",
        email: "user@example.com",
      },
      accessToken: "access-token",
      refreshToken: "refresh-token",
      expiresAt: 2222 * 1000,
      refreshAt: 1111 * 1000,
    });
  });

  it("returns 401 with error message when login fails", async () => {
    mockLogin.mockRejectedValueOnce(new Error("Invalid credentials"));

    const body = { email: "user@example.com", password: "wrong" };

    const request = new NextRequest("http://localhost/api/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    });

    const response = await POST(request);
    const json = await response.json();

    expect(response.status).toBe(401);
    expect(json.error).toBe("Invalid credentials");
  });
});
