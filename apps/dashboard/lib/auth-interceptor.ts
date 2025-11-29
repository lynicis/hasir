import type { Interceptor } from "@connectrpc/connect";

export const authInterceptor: Interceptor = (next) => async (req) => {
  const sessionResponse = await fetch('/api/auth/session');
  if (sessionResponse.ok) {
    const session = await sessionResponse.json();

    if (session.accessToken) {
      req.header.set("Authorization", `Bearer ${session.accessToken}`);
    }
  }

  return await next(req);
};