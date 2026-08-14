import { describe, it, expect } from 'bun:test';
import { NextRequest } from 'next/server';

import { proxy } from './proxy';

describe('proxy', () => {
  it('should allow access to public paths without authentication', async () => {
    const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password'];

    for (const path of publicPaths) {
      const request = new NextRequest(new URL(path, 'http://localhost:3000'));
      const response = await proxy(request);

      expect(response.status).not.toBe(307);
    }
  });

  it('should redirect unauthenticated users to login with redirect parameter', async () => {
    const inviteUrl = 'http://localhost:3000/invite/abc123';
    const request = new NextRequest(new URL(inviteUrl));
    const response = await proxy(request);

    expect(response.status).toBe(307);
    const redirectUrl = new URL(response.headers.get('location') || '');
    expect(redirectUrl.pathname).toBe('/login');
    expect(redirectUrl.searchParams.get('redirect')).toBe('/invite/abc123');
  });

  it('should redirect unauthenticated users accessing dashboard', async () => {
    const dashboardUrl = 'http://localhost:3000/dashboard';
    const request = new NextRequest(new URL(dashboardUrl));
    const response = await proxy(request);

    expect(response.status).toBe(307);
    const redirectUrl = new URL(response.headers.get('location') || '');
    expect(redirectUrl.pathname).toBe('/login');
    expect(redirectUrl.searchParams.get('redirect')).toBe('/dashboard');
  });

  it('should allow authenticated users to access protected routes', async () => {
    const inviteUrl = 'http://localhost:3000/invite/abc123';
    const request = new NextRequest(new URL(inviteUrl));
    request.cookies.set('hasir-session', 'mock-session-token');

    const response = await proxy(request);

    expect(response.status).not.toBe(307);
  });

  it('should allow access to API routes without authentication', async () => {
    const apiUrl = 'http://localhost:3000/api/auth/session';
    const request = new NextRequest(new URL(apiUrl));
    const response = await proxy(request);

    expect(response.status).not.toBe(307);
  });

  it('should allow access to static files without authentication', async () => {
    const staticUrl = 'http://localhost:3000/_next/static/file.js';
    const request = new NextRequest(new URL(staticUrl));
    const response = await proxy(request);

    expect(response.status).not.toBe(307);
  });

  it('should allow access to root path without authentication', async () => {
    const rootUrl = 'http://localhost:3000/';
    const request = new NextRequest(new URL(rootUrl));
    const response = await proxy(request);

    expect(response.status).not.toBe(307);
  });

  it('should bypass authentication and security headers for __nextjs dev endpoints', async () => {
    const devUrl = 'http://localhost:3000/__nextjs_original-stack-frames';
    const request = new NextRequest(new URL(devUrl));
    const response = await proxy(request);

    expect(response.status).not.toBe(307);
    expect(response.headers.get('Content-Security-Policy')).toBeNull();
  });

  it('should set hasir-csrf cookie on GET requests if not present', async () => {
    const rootUrl = 'http://localhost:3000/';
    const request = new NextRequest(new URL(rootUrl));
    const response = await proxy(request);

    const setCookie = response.headers.get('set-cookie');
    expect(setCookie).toContain('hasir-csrf=');
  });

  it('should return 403 for POST requests without a CSRF token', async () => {
    const apiUrl = 'http://localhost:3000/api/auth/login';
    const request = new NextRequest(new URL(apiUrl), { method: 'POST' });
    
    // Cookie exists but header is missing
    request.cookies.set('hasir-csrf', 'valid-token');
    
    const response = await proxy(request);
    
    expect(response.status).toBe(403);
    const body = await response.json();
    expect(body.error).toBe('Invalid CSRF token');
  });

  it('should return 403 for POST requests with a mismatched CSRF token', async () => {
    const apiUrl = 'http://localhost:3000/api/auth/login';
    const request = new NextRequest(new URL(apiUrl), { method: 'POST' });
    
    request.cookies.set('hasir-csrf', 'valid-token');
    request.headers.set('x-csrf-token', 'invalid-token');
    
    const response = await proxy(request);
    
    expect(response.status).toBe(403);
  });

  it('should allow POST requests with matching CSRF token', async () => {
    const apiUrl = 'http://localhost:3000/api/auth/login';
    const request = new NextRequest(new URL(apiUrl), { method: 'POST' });
    
    const token = 'valid-token';
    request.cookies.set('hasir-csrf', token);
    request.headers.set('x-csrf-token', token);
    
    const response = await proxy(request);
    
    expect(response.status).not.toBe(403);
  });
});
