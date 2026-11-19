import { createConnectTransport } from "@connectrpc/connect-web";

import { idempotencyInterceptor } from "./idempotency-interceptor";
import { csrfInterceptor } from "./csrf-interceptor";
import { authInterceptor } from "./auth-interceptor";

export const connectTransport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "",
  interceptors: [csrfInterceptor, authInterceptor, idempotencyInterceptor],
});
