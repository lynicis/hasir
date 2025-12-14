import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";

import { customRetry } from "./query-retry";

describe("customRetry", () => {
  it("should return false when error is null", () => {
    expect(customRetry(0, null as unknown as Error)).toBe(false);
  });

  it("should return false when error is undefined", () => {
    expect(customRetry(0, undefined as unknown as Error)).toBe(false);
  });

  it("should return false for NotFound errors", () => {
    const notFoundError = new ConnectError("Not found", Code.NotFound);
    expect(customRetry(0, notFoundError)).toBe(false);
    expect(customRetry(1, notFoundError)).toBe(false);
    expect(customRetry(2, notFoundError)).toBe(false);
  });

  it("should return false when failureCount is 3 or more", () => {
    const error = new Error("Some error");
    expect(customRetry(3, error)).toBe(false);
    expect(customRetry(4, error)).toBe(false);
    expect(customRetry(10, error)).toBe(false);
  });

  it("should return true when failureCount is less than 3 and error is not NotFound", () => {
    const error = new Error("Some error");
    expect(customRetry(0, error)).toBe(true);
    expect(customRetry(1, error)).toBe(true);
    expect(customRetry(2, error)).toBe(true);
  });

  it("should return false when failureCount is 2 and error is NotFound", () => {
    const notFoundError = new ConnectError("Not found", Code.NotFound);
    expect(customRetry(2, notFoundError)).toBe(false);
  });

  it("should return false when failureCount is 3 and error is not NotFound", () => {
    const error = new Error("Some error");
    expect(customRetry(3, error)).toBe(false);
  });
});