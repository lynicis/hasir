import { describe, it, expect } from 'vitest';
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
});
