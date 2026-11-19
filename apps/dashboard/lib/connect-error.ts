import { Code, ConnectError } from "@connectrpc/connect";
import { toast } from "sonner";

export type ConnectErrorMessages = Partial<
  Record<Code, string | ((error: ConnectError) => string)>
> & {
  fallback: string;
};

export function getConnectErrorMessage(
  error: unknown,
  messages: ConnectErrorMessages
): string {
  if (error instanceof ConnectError) {
    const handler = messages[error.code];
    if (typeof handler === "function") {
      return handler(error);
    }
    if (typeof handler === "string") {
      return handler;
    }
  }
  return messages.fallback;
}

export function handleConnectErrorToast(
  error: unknown,
  messages: ConnectErrorMessages
): void {
  toast.error(getConnectErrorMessage(error, messages));
}
