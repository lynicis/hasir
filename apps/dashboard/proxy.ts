import type { NextRequest } from 'next/server';

import { NextResponse } from 'next/server';

const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password'];

const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";

const csrfMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];

function addSecurityHeaders(request: NextRequest, response: NextResponse): NextResponse {
  // Generate a random nonce for this request
  const nonce = Buffer.from(crypto.randomUUID()).toString('base64');

  // Set the nonce in the request headers so Next.js can read it for Script components
  request.headers.set('x-nonce', nonce);

  // In development, Next.js Fast Refresh requires unsafe-eval.
  // In production, we strictly enforce the nonce.
  const isDev = process.env.NODE_ENV !== 'production';
  const scriptSrc = isDev
    ? `'self' 'unsafe-inline' 'unsafe-eval'`
    : `'self' 'nonce-${nonce}' 'strict-dynamic'`;

  const csp = [
    "default-src 'self'",
    `script-src ${scriptSrc}`,
    "style-src 'self' 'unsafe-inline'", // Tailwind/Next.js styles often need this
    "img-src 'self' data: blob:",
    "font-src 'self' data:",
    `connect-src 'self' ${apiUrl} https: wss:`,
    "form-action 'self'",
    "frame-ancestors 'none'",
    "base-uri 'self'",
  ].join("; ");

  response.headers.set("Content-Security-Policy", csp);
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");

  // Forward the updated request headers
  return response;
}

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname.startsWith('/__nextjs')) {
    return NextResponse.next();
  }

  const isStaticFile = pathname.startsWith('/_next') || pathname.includes('.');

  // CSRF Protection Validation
  let csrfToken = request.cookies.get('hasir-csrf')?.value;

  if (!isStaticFile && csrfMethods.includes(request.method)) {
    const headerToken = request.headers.get('x-csrf-token');
    if (!csrfToken || !headerToken || csrfToken !== headerToken) {
      return new NextResponse(JSON.stringify({ error: 'Invalid CSRF token' }), {
        status: 403,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  // Pass request down so it contains the injected x-nonce header
  let response = NextResponse.next({
    request: {
      headers: request.headers,
    }
  });

  // Redirect logic for unauthenticated users
  if (
    !isStaticFile &&
    pathname !== '/' &&
    !pathname.startsWith('/api/auth') &&
    !publicPaths.some((path) => pathname.startsWith(path))
  ) {
    const sessionCookie = request.cookies.get('hasir-session');
    if (!sessionCookie) {
      const loginUrl = new URL('/login', request.url);
      loginUrl.searchParams.set('redirect', pathname);
      response = NextResponse.redirect(loginUrl, {
        headers: request.headers // Preserve nonce header on redirect
      });
    }
  }

  // Set CSRF token on responses if it doesn't exist
  if (!isStaticFile && !csrfToken) {
    csrfToken = crypto.randomUUID();
    const isProd = process.env.NODE_ENV === 'production';
    response.cookies.set('hasir-csrf', csrfToken, {
      path: '/',
      httpOnly: false, // Needs to be readable by JS fetch
      secure: isProd,
      sameSite: 'lax',
    });
  }

  if (!isStaticFile) {
    response = addSecurityHeaders(request, response);
  }

  return response;
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|__nextjs|.*\\..*).*)',
  ],
};
