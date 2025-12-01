import { UserService } from '@buf/hasir_hasir.bufbuild_es/user/v1/user_pb';
import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { decodeJwt, JWTPayload } from 'jose';
import { NextResponse } from 'next/server';

import { getSession, refreshSession } from '@/lib/session';

const transport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  useBinaryFormat: true,
});

export async function GET() {
  const session = await getSession();

  if (!session.user) {
    return NextResponse.json({ user: null }, { status: 401 });
  }

  if (session.expiresAt && session.expiresAt < Date.now()) {
    session.destroy();
    return NextResponse.json({ user: null }, { status: 401 });
  }

  if (session.refreshAt && session.refreshAt < Date.now()) {
    if (!session.refreshToken) {
      session.destroy();
      return NextResponse.json({ user: null }, { status: 401 });
    }

    try {
      const client = createClient(UserService, transport);
      const response = await client.renewTokens({
        refreshToken: session.refreshToken,
      });

      const newAccessPayload = decodeJwt(response.accessToken) as JWTPayload;
      const newAccessExpiryMs = (newAccessPayload.exp ?? 0) * 1000;

      session.accessToken = response.accessToken;
      session.refreshAt = newAccessExpiryMs;

      await refreshSession(session);
    } catch {
      session.destroy();
      return NextResponse.json({ user: null }, { status: 401 });
    }
  }

  return NextResponse.json({
    user: session.user,
    accessToken: session.accessToken,
  });
}
