/* eslint-disable @typescript-eslint/no-empty-object-type, @typescript-eslint/no-explicit-any */
/// <reference types="bun-types/test-globals" />


import type { TestingLibraryMatchers } from "@testing-library/jest-dom/matchers";

import { Mock as BunMock, mock as bunMock } from "bun:test";

declare module "bun:test" {
  interface Matchers<T = any> extends TestingLibraryMatchers<typeof expect.stringContaining, T> {}
  interface AsymmetricMatchers extends TestingLibraryMatchers {}
}

declare global {
  type Mock<T extends (...args: any[]) => any = (...args: any[]) => any> = BunMock<T>;
  const mock: typeof bunMock;
}
