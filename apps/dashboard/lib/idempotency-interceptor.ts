import type { Interceptor } from "@connectrpc/connect";

import { MethodOptions_IdempotencyLevel } from "@bufbuild/protobuf/wkt";

export const idempotencyInterceptor: Interceptor = (next) => async (req) => {
  if (req.method.idempotency === MethodOptions_IdempotencyLevel.IDEMPOTENT) {
    req.header.set("Idempotency-Key", crypto.randomUUID());
  }
  return await next(req);
};
