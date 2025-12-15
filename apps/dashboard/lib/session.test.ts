import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { getIronSession, type IronSession } from "iron-session";
import { cookies } from "next/headers";
import { DateTime } from "luxon";

import {
  getSession,
  saveSession,
  destroySession,
  refreshSession,
  isExpiredSeconds,
  isExpiredMillis,
  type SessionData,
} from "./session";

vi.mock("next/headers", () => ({
  cookies: vi.fn(),
}));

vi.mock("iron-session", () => ({
  getIronSession: vi.fn(),
}));

describe("session", () => {
  const mockCookieStore = {} as unknown;
  const mockIronSession = {
    save: vi.fn().mockResolvedValue(undefined),
    destroy: vi.fn().mockResolvedValue(undefined),
  } as unknown as IronSession<SessionData>;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(cookies).mockResolvedValue(mockCookieStore as Awaited<ReturnType<typeof cookies>>);
    vi.mocked(getIronSession).mockResolvedValue(mockIronSession);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("getSession", () => {
    it("should get session from cookies", async () => {
      const session = await getSession();
      expect(cookies).toHaveBeenCalled();
      expect(getIronSession).toHaveBeenCalledWith(
        mockCookieStore,
        expect.objectContaining({
          password: expect.any(String),
          cookieName: "hasir-session",
        })
      );
      expect(session).toBe(mockIronSession);
    });
  });

  describe("saveSession", () => {
    it("should save session data", async () => {
      const data: SessionData = {
        user: { id: "123", email: "test@example.com" },
        accessToken: "token",
      };

      const session = await saveSession(data);

      expect(cookies).toHaveBeenCalled();
      expect(getIronSession).toHaveBeenCalled();
      expect(mockIronSession.save).toHaveBeenCalled();
      expect(session).toBe(mockIronSession);
      expect(mockIronSession.user).toEqual(data.user);
      expect(mockIronSession.accessToken).toBe(data.accessToken);
    });
  });

  describe("destroySession", () => {
    it("should destroy session", async () => {
      await destroySession();

      expect(cookies).toHaveBeenCalled();
      expect(getIronSession).toHaveBeenCalled();
      expect(mockIronSession.destroy).toHaveBeenCalled();
    });
  });

  describe("refreshSession", () => {
    it("should refresh session", async () => {
      await refreshSession(mockIronSession);

      expect(mockIronSession.save).toHaveBeenCalled();
    });
  });

  describe("isExpiredSeconds", () => {
    it("should return true when exp is undefined", () => {
      expect(isExpiredSeconds(undefined)).toBe(true);
    });

    it("should return true when exp is in the past", () => {
      const pastTime = DateTime.utc().minus({ seconds: 100 }).toSeconds();
      expect(isExpiredSeconds(pastTime)).toBe(true);
    });

    it("should return false when exp is in the future", () => {
      const futureTime = DateTime.utc().plus({ seconds: 100 }).toSeconds();
      expect(isExpiredSeconds(futureTime)).toBe(false);
    });

    it("should return true when exp is exactly now", () => {
      const now = DateTime.utc().toSeconds();
      const slightlyPast = now - 1;

      expect(isExpiredSeconds(slightlyPast)).toBe(true);
    });
  });

  describe("isExpiredMillis", () => {
    it("should return true when timestamp is undefined", () => {
      expect(isExpiredMillis(undefined)).toBe(true);
    });

    it("should return true when timestamp is in the past", () => {
      const pastTime = DateTime.utc().minus({ milliseconds: 100 }).toMillis();
      expect(isExpiredMillis(pastTime)).toBe(true);
    });

    it("should return false when timestamp is in the future", () => {
      const futureTime = DateTime.utc()
        .plus({ milliseconds: 100 })
        .toMillis();
      expect(isExpiredMillis(futureTime)).toBe(false);
    });

    it("should return true when timestamp is exactly now", () => {
      const now = DateTime.utc().toMillis();
      const slightlyPast = now - 1;
      
      expect(isExpiredMillis(slightlyPast)).toBe(true);
    });
  });
});