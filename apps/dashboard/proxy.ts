import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password'];

const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";

function addSecurityHeaders(response: NextResponse): NextResponse {
  response.headers.set(
    "Content-Security-Policy",
    [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
      "style-src 'self' 'unsafe-inline'",
      "img-src 'self' data: blob:",
      "font-src 'self' data:",
      `connect-src 'self' ${apiUrl} https: wss:`,
      "form-action 'self'",
      "frame-ancestors 'none'",
      "base-uri 'self'",
    ].join("; "),
  );
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
  return response;
}

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (
    pathname === '/' ||
    publicPaths.some((path) => pathname.startsWith(path)) ||
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api/auth') ||
    pathname.includes('.')
  ) {
    return addSecurityHeaders(NextResponse.next());
  }

  const sessionCookie = request.cookies.get('hasir-session');

  if (!sessionCookie) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return addSecurityHeaders(NextResponse.redirect(loginUrl));
  }

  return addSecurityHeaders(NextResponse.next());
}

export const config = {
  matcher: [
    '/((?!api/auth|_next/static|_next/image|favicon.ico|.*\\..*).*)',
  ],
};
