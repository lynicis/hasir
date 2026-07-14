import { type DescService } from "@bufbuild/protobuf";
import { renderHook } from "@testing-library/react";
import { type Client } from "@connectrpc/connect";

const mockClient = {} as Client<DescService>;
const mockCreateClient = vi.fn(() => mockClient);

vi.mock("@connectrpc/connect", () => ({
  createClient: mockCreateClient,
}));

vi.mock("./auth-interceptor", () => ({
  authInterceptor: vi.fn(),
}));

import { useClient } from "./use-client.impl";

describe("useClient", () => {
  const mockService = {} as DescService;

  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateClient.mockReturnValue(mockClient);
  });

  it("should create a client with the service", () => {
    const { result } = renderHook(() => useClient(mockService));

    expect(mockCreateClient).toHaveBeenCalledWith(
      mockService,
      expect.anything()
    );
    expect(result.current).toBe(mockClient);
  });

  it("should memoize the client", () => {
    const { result, rerender } = renderHook(() => useClient(mockService));

    const firstClient = result.current;

    rerender();

    expect(result.current).toBe(firstClient);
    expect(mockCreateClient).toHaveBeenCalledTimes(1);
  });

  it("should create a new client when service changes", () => {
    const mockService1 = {} as DescService;
    const mockService2 = {} as DescService;
    const mockClient1 = {} as Client<DescService>;
    const mockClient2 = {} as Client<DescService>;

    mockCreateClient
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
    expect(mockCreateClient).toHaveBeenCalledTimes(2);
  });
});
