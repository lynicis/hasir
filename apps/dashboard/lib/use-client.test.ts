import { type Client, createClient } from "@connectrpc/connect";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { type DescService } from "@bufbuild/protobuf";
import { renderHook } from "@testing-library/react";

const mockTransport = vi.hoisted(() => ({} as unknown));
const mockClient = vi.hoisted(() => ({} as Client<DescService>));

vi.mock("@connectrpc/connect-web", () => ({
  createConnectTransport: vi.fn(() => mockTransport) as unknown,
}));

vi.mock("@connectrpc/connect", () => ({
  createClient: vi.fn(() => mockClient),
}));

vi.mock("./auth-interceptor", () => ({
  authInterceptor: vi.fn(),
}));

import { useClient } from "./use-client";

describe("useClient", () => {
  const mockService = {} as DescService;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(createClient).mockReturnValue(mockClient);
  });

  it("should create a client with the service", () => {
    const { result } = renderHook(() => useClient(mockService));

    expect(createClient).toHaveBeenCalledWith(mockService, mockTransport);
    expect(result.current).toBe(mockClient);
  });

  it("should memoize the client", () => {
    const { result, rerender } = renderHook(() => useClient(mockService));

    const firstClient = result.current;

    rerender();

    expect(result.current).toBe(firstClient);
    expect(createClient).toHaveBeenCalledTimes(1);
  });

  it("should create a new client when service changes", () => {
    const mockService1 = {} as DescService;
    const mockService2 = {} as DescService;
    const mockClient1 = {} as Client<DescService>;
    const mockClient2 = {} as Client<DescService>;

    vi.mocked(createClient)
      .mockReturnValueOnce(mockClient1)
      .mockReturnValueOnce(mockClient2);

    const { result, rerender } = renderHook(
      ({ service }) => useClient(service),
      {
        initialProps: { service: mockService1 },
      }
    );

    expect(result.current).toBe(mockClient1);

    rerender({ service: mockService2 });

    expect(result.current).toBe(mockClient2);
    expect(createClient).toHaveBeenCalledTimes(2);
  });
});