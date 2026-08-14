import type { Interceptor } from "@connectrpc/connect";

import { getCsrfToken } from "./csrf";

export const csrfInterceptor: Interceptor = (next) => async (req) => {
  const token = getCsrfToken();
  if (token) {
    req.header.set("X-CSRF-Token", token);
  }
  return await next(req);
};
