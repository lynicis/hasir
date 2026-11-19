import { Code, ConnectError } from "@connectrpc/connect";
import { toast } from "sonner";

import {
  getConnectErrorMessage,
  handleConnectErrorToast,
} from "./connect-error";

describe("connect-error", () => {
  it("maps matching ConnectError code to string message", () => {
    const err = new ConnectError("denied", Code.PermissionDenied);
    expect(
      getConnectErrorMessage(err, {
        [Code.PermissionDenied]: "No access.",
        fallback: "Generic failure.",
      })
    ).toBe("No access.");
  });

  it("supports function mapper receiving the ConnectError", () => {
    const err = new ConnectError("bad field", Code.InvalidArgument);
    expect(
      getConnectErrorMessage(err, {
        [Code.InvalidArgument]: (e) => e.rawMessage || "Invalid input.",
        fallback: "Generic failure.",
      })
    ).toBe("bad field");
  });

  it("returns fallback for unmapped ConnectError codes and non-ConnectError values", () => {
    const err = new ConnectError("boom", Code.Internal);
    expect(
      getConnectErrorMessage(err, {
        [Code.NotFound]: "Not found.",
        fallback: "Generic failure.",
      })
    ).toBe("Generic failure.");

    expect(
      getConnectErrorMessage(new Error("network"), {
        [Code.NotFound]: "Not found.",
        fallback: "Generic failure.",
      })
    ).toBe("Generic failure.");
  });

  it("calls toast.error via handleConnectErrorToast", () => {
    const errorSpy = vi.spyOn(toast, "error");
    const err = new ConnectError("missing", Code.NotFound);
    handleConnectErrorToast(err, {
      [Code.NotFound]: "Key not found.",
      fallback: "Generic failure.",
    });

    expect(errorSpy).toHaveBeenCalledWith("Key not found.");
    errorSpy.mockRestore();
  });
});
