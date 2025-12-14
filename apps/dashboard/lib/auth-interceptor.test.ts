import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";

import { authInterceptor } from "./auth-interceptor";

const mockLocation = {
  pathname: "/",
};

const mockReplaceState = vi.fn();

global.fetch = vi.fn();

type MockRequest = {
  url?: string;
  header: {
    set: ReturnType<typeof vi.fn>;
  };
};

describe("authInterceptor", () => {
  const mockNext = vi.fn();
  const mockRequest: MockRequest = {
    url: "https://api.example.com/service/method",
    header: {
      set: vi.fn(),
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, "location", {
      value: mockLocation,
      writable: true,
      configurable: true,
    });
    Object.defineProperty(window, "history", {
      value: {
        replaceState: mockReplaceState,
      },
      writable: true,
      configurable: true,
    });
    vi.mocked(global.fetch).mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("public methods", () => {
    it("should allow public methods without authentication", async () => {
      mockRequest.url = "https://api.example.com/auth/login";
      mockLocation.pathname = "/dashboard";

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalledWith(mockRequest);
    });

    it("should allow login method", async () => {
      mockRequest.url = "https://api.example.com/auth/login";
      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });

    it("should allow register method", async () => {
      mockRequest.url = "https://api.example.com/auth/register";
      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });

    it("should allow forgotpassword method", async () => {
      mockRequest.url = "https://api.example.com/auth/forgotpassword";
      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });

    it("should allow resetpassword method", async () => {
      mockRequest.url = "https://api.example.com/auth/resetpassword";
      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });
  });

  describe("public pages", () => {
    it("should allow requests from root path", async () => {
      mockLocation.pathname = "/";
      mockRequest.url = "https://api.example.com/any/method";

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });

    it("should allow requests from login page", async () => {
      mockLocation.pathname = "/login";
      mockRequest.url = "https://api.example.com/any/method";

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });

    it("should allow requests from register page", async () => {
      mockLocation.pathname = "/register";
      mockRequest.url = "https://api.example.com/any/method";

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalled();
    });
  });

  describe("authenticated requests", () => {
    it("should add Authorization header when session has accessToken", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          accessToken: "test-token",
        }),
      } as Response);

      mockNext.mockResolvedValue({});

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).toHaveBeenCalledWith("/api/auth/session");
      expect(mockRequest.header.set).toHaveBeenCalledWith(
        "Authorization",
        "Bearer test-token"
      );
      expect(mockNext).toHaveBeenCalledWith(mockRequest);
    });

    it("should not add Authorization header when session has no accessToken", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      } as Response);

      mockNext.mockResolvedValue({});

      const interceptor = authInterceptor(mockNext);
      await interceptor(mockRequest as never);

      expect(global.fetch).toHaveBeenCalledWith("/api/auth/session");
      expect(mockRequest.header.set).not.toHaveBeenCalled();
      expect(mockNext).toHaveBeenCalledWith(mockRequest);
    });
  });

  describe("unauthorized requests", () => {
    it("should redirect to login when session returns 401", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      vi.mocked(global.fetch)
        .mockResolvedValueOnce({
          status: 401,
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
        } as Response);

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow(
        ConnectError
      );

      expect(global.fetch).toHaveBeenCalledWith("/api/auth/session");
      expect(global.fetch).toHaveBeenCalledWith("/api/auth/logout", {
        method: "POST",
      });
      expect(mockReplaceState).toHaveBeenCalledWith(null, "", "/login");
      expect(mockNext).not.toHaveBeenCalled();
    });

    it("should redirect to login when request throws Unauthenticated error", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      const unauthenticatedError = new ConnectError(
        "Unauthenticated",
        Code.Unauthenticated
      );

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          accessToken: "test-token",
        }),
      } as Response);

      mockNext.mockRejectedValueOnce(unauthenticatedError);

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow(
        ConnectError
      );

      expect(global.fetch).toHaveBeenCalledWith("/api/auth/logout", {
        method: "POST",
      });
      expect(mockReplaceState).toHaveBeenCalledWith(null, "", "/login");
    });

    it("should not redirect for public methods even if error occurs", async () => {
      mockRequest.url = "https://api.example.com/auth/login";
      mockLocation.pathname = "/login";

      const error = new ConnectError("Some error", Code.Internal);
      mockNext.mockRejectedValueOnce(error);

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow(ConnectError);

      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockReplaceState).not.toHaveBeenCalled();
    });

    it("should not redirect for other error types", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      const otherError = new ConnectError("Not found", Code.NotFound);

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          accessToken: "test-token",
        }),
      } as Response);

      mockNext.mockRejectedValueOnce(otherError);

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow(ConnectError);

      expect(mockReplaceState).not.toHaveBeenCalled();
    });
  });

  describe("error handling", () => {
    it("should propagate fetch errors", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      vi.mocked(global.fetch).mockRejectedValueOnce(new Error("Network error"));

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow("Network error");
      expect(mockNext).not.toHaveBeenCalled();
    });

    it("should handle logout fetch errors gracefully", async () => {
      mockLocation.pathname = "/dashboard";
      mockRequest.url = "https://api.example.com/protected/method";

      vi.mocked(global.fetch)
        .mockResolvedValueOnce({
          status: 401,
          ok: false,
        } as Response)
        .mockRejectedValueOnce(new Error("Logout failed"));

      const interceptor = authInterceptor(mockNext);

      await expect(interceptor(mockRequest as never)).rejects.toThrow();

      expect(mockReplaceState).toHaveBeenCalled();
    });
  });
});