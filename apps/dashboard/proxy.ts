import { NextResponse } from "next/server";
import { NextRequest } from "next/server";

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const accessToken = request.cookies.get("accessToken")?.value;
  const refreshToken = request.cookies.get("refreshToken")?.value;


  if (pathname === "/" && accessToken && refreshToken) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }


  const isAuthenticatedRoute =
    pathname.startsWith("/dashboard") || pathname.startsWith("/profile");

  if (isAuthenticatedRoute && (!accessToken || !refreshToken)) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/dashboard/:path*", "/profile/:path*"],
};

