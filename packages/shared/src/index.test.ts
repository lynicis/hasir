import { describe, expect, it } from "bun:test";

describe("shared", () => {
  it("exports expected modules", async () => {
    const shared = await import("./index");
    expect(shared).toBeDefined();
  });
});
