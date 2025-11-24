import { NextRequest } from "next/server";

import { NextResponse } from "next/server";
import { decodeJwt } from "jose";

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname === "/") {
    const accessToken = request.cookies.get("accessToken")?.value;

    if (accessToken) {
      try {
        const payload = decodeJwt(accessToken);
        const isAuthenticated = Boolean(payload.sub && payload.email);

        if (isAuthenticated) {
          return NextResponse.redirect(new URL("/dashboard", request.url));
        }
      } catch {}
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: "/",
};

