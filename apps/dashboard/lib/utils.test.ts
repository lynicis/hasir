import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";

import { cn, isNotFoundError, isUnauthenticatedError } from "./utils";

describe("cn", () => {
  it("should merge class names correctly", () => {
    expect(cn("foo", "bar")).toBe("foo bar");
  });

  it("should handle conditional classes", () => {
    expect(cn("foo", false && "bar", "baz")).toBe("foo baz");
  });

  it("should merge Tailwind classes correctly", () => {
    expect(cn("px-2 py-1", "px-4")).toBe("py-1 px-4");
  });

  it("should handle empty inputs", () => {
    expect(cn()).toBe("");
  });

  it("should handle undefined and null", () => {
    expect(cn("foo", undefined, null, "bar")).toBe("foo bar");
  });
});

describe("isNotFoundError", () => {
  it("should return true for NotFound ConnectError", () => {
    const error = new ConnectError("Not found", Code.NotFound);
    expect(isNotFoundError(error)).toBe(true);
  });

  it("should return false for other ConnectError codes", () => {
    const error = new ConnectError("Unauthenticated", Code.Unauthenticated);
    expect(isNotFoundError(error)).toBe(false);
  });

  it("should return false for non-ConnectError errors", () => {
    const error = new Error("Some error");
    expect(isNotFoundError(error)).toBe(false);
  });

  it("should return false for null", () => {
    expect(isNotFoundError(null)).toBe(false);
  });

  it("should return false for undefined", () => {
    expect(isNotFoundError(undefined)).toBe(false);
  });
});

describe("isUnauthenticatedError", () => {
  it("should return true for Unauthenticated ConnectError", () => {
    const error = new ConnectError("Unauthenticated", Code.Unauthenticated);
    expect(isUnauthenticatedError(error)).toBe(true);
  });

  it("should return false for other ConnectError codes", () => {
    const error = new ConnectError("Not found", Code.NotFound);
    expect(isUnauthenticatedError(error)).toBe(false);
  });

  it("should return false for non-ConnectError errors", () => {
    const error = new Error("Some error");
    expect(isUnauthenticatedError(error)).toBe(false);
  });

  it("should return false for null", () => {
    expect(isUnauthenticatedError(null)).toBe(false);
  });

  it("should return false for undefined", () => {
    expect(isUnauthenticatedError(undefined)).toBe(false);
  });
});