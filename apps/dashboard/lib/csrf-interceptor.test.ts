import { csrfInterceptor } from "./csrf-interceptor";
import { getCsrfToken } from "./csrf";

vi.mock("./csrf", () => ({
  getCsrfToken: vi.fn(),
}));

const mockedGetCsrfToken = getCsrfToken as unknown as Mock;

describe("csrfInterceptor", () => {
  beforeEach(() => {
    mockedGetCsrfToken.mockReset();
  });

  it("should set X-CSRF-Token header when token exists", async () => {
    mockedGetCsrfToken.mockReturnValue("test-csrf-token");

    const mockNext = vi.fn().mockResolvedValue({ header: new Headers() });
    const intercepted = csrfInterceptor(mockNext);

    const mockReq = {
      header: new Headers(),
    };

    await intercepted(mockReq as never);

    expect(mockReq.header.get("X-CSRF-Token")).toBe("test-csrf-token");
    expect(mockNext).toHaveBeenCalledWith(mockReq);
  });

  it("should not set header when token is empty", async () => {
    mockedGetCsrfToken.mockReturnValue("");

    const mockNext = vi.fn().mockResolvedValue({ header: new Headers() });
    const intercepted = csrfInterceptor(mockNext);

    const mockReq = {
      header: new Headers(),
    };

    await intercepted(mockReq as never);

    expect(mockReq.header.has("X-CSRF-Token")).toBe(false);
    expect(mockNext).toHaveBeenCalledWith(mockReq);
  });

  it("should pass through to next interceptor", async () => {
    mockedGetCsrfToken.mockReturnValue("token");

    const expectedResponse = { header: new Headers(), message: "ok" };
    const mockNext = vi.fn().mockResolvedValue(expectedResponse);
    const intercepted = csrfInterceptor(mockNext);

    const mockReq = {
      header: new Headers(),
    };

    const result = await intercepted(mockReq as never);

    expect((result as unknown as { message: string }).message).toBe("ok");
  });
});
