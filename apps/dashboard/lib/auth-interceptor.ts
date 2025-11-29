import type { Interceptor } from "@connectrpc/connect";

import { RedirectType, redirect } from "next/navigation";

export const authInterceptor: Interceptor = (next) => async (req) => {
  const sessionResponse = await fetch('/api/auth/session');
  if (sessionResponse.ok) {
    const session = await sessionResponse.json();

    if (session.accessToken) {
      req.header.set("Authorization", `Bearer ${session.accessToken}`);
    }
  }
  
  if (sessionResponse.status === 401) {
    redirect("/login", RedirectType.replace)
  }

  return await next(req);
};