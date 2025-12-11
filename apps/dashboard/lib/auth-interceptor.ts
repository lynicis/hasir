import { type Interceptor, Code, ConnectError } from "@connectrpc/connect";

const publicMethods = ['login', 'register', 'forgotpassword', 'resetpassword'];

function isPublicPage(): boolean {
  const pathname = window.location.pathname;
  return pathname === '/' ||
    pathname === '/login' ||
    pathname === '/register';
}

function isPublicMethod(req: { url?: string }): boolean {
  const url = req.url?.toLowerCase() || '';
  if (!url) return false;

  const urlPath = url.split('?')[0];
  const parts = urlPath.split('/').filter(Boolean);
  const methodPart = parts[parts.length - 1] || '';

  return publicMethods.some(method =>
    methodPart === method ||
    methodPart.endsWith(method) ||
    url.includes(`/${method}/`) ||
    url.includes(`/${method}`)
  );
}

async function destroySession(): Promise<void> {
  try {
    await fetch('/api/auth/logout', { method: 'POST' });
  } catch (error) {
    console.warn('Failed to destroy session:', error);
  }
}

async function redirectToLogin(): Promise<void> {
  await destroySession();
  window.history.replaceState(null, "", "/login");
}

export const authInterceptor: Interceptor = (next) => async (req) => {
  const isPublic = isPublicMethod(req) || isPublicPage();

  if (isPublic) {
    return await next(req);
  }

  const sessionResponse = await fetch('/api/auth/session');

  if (sessionResponse.status === 401) {
    await redirectToLogin();
    throw new ConnectError("Unauthenticated", Code.Unauthenticated);
  }

  if (sessionResponse.ok) {
    const session = await sessionResponse.json();
    if (session.accessToken) {
      req.header.set("Authorization", `Bearer ${session.accessToken}`);
    }
  }

  try {
    return await next(req);
  } catch (error) {
    if (
      error instanceof ConnectError &&
      error.code === Code.Unauthenticated &&
      !isPublic
    ) {
      await redirectToLogin();
    }

    throw error;
  }
};
