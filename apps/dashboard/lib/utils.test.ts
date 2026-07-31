import { Code, ConnectError } from "@connectrpc/connect";

import { isNotFoundError, isUnauthenticatedError } from "./utils";

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
