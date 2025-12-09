import { UserService } from '@buf/hasir_hasir.bufbuild_es/user/v1/user_pb';
import { createConnectTransport } from '@connectrpc/connect-web';
import { NextRequest, NextResponse } from 'next/server';
import { createClient } from '@connectrpc/connect';
import { decodeJwt, JWTPayload } from 'jose';
import { DateTime } from 'luxon';
import { z } from "zod/v4";

import {
  getSession,
  isExpiredSeconds,
  isExpiredMillis,
  refreshSession,
  saveSession
} from '@/lib/session';

const transport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  useBinaryFormat: true,
});

export async function GET() {
  const session = await getSession();

  if (!session.user) {
    return NextResponse.json({ user: null }, { status: 401 });
  }

  if (isExpiredMillis(session.expiresAt)) {
    session.destroy();
    return NextResponse.json({ user: null }, { status: 401 });
  }

  if (isExpiredMillis(session.refreshAt)) {
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
      const newAccessExpiryMs = DateTime.fromSeconds(newAccessPayload.exp ?? 0).toMillis();

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

const schema = z.object({
    accessToken: z.jwt(),
    newTokens: z.object({
        accessToken: z.jwt(),
        refreshToken: z.jwt(),
    }),
});

export async function POST(request: NextRequest) {
    const requestBody = await request.json();

    const { data: parsedRequestBody, error: parseError } = await schema.safeParseAsync(requestBody);
    if (parseError) {
        return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
    }

    const { accessToken, newTokens } = parsedRequestBody;
    const { accessToken: newAccessToken, refreshToken } = newTokens;

    try {
        decodeJwt(accessToken);
    } catch {
        return NextResponse.json({ error: "Invalid access token" }, { status: 400 });
    }

    let newAccessTokenPayload: JWTPayload;
    try {
        newAccessTokenPayload = decodeJwt<JWTPayload>(newAccessToken);
    } catch {
        return NextResponse.json({ error: "Invalid new access token" }, { status: 400 });
    }

    if (isExpiredSeconds(newAccessTokenPayload.exp)) {
        return NextResponse.json({ error: "New access token expired" }, { status: 400 });
    }

    if (newAccessTokenPayload.iss !== process.env.NEXT_PUBLIC_API_URL) {
        return NextResponse.json({ error: "Invalid new access token issuer" }, { status: 400 });
    }

    const userId = newAccessTokenPayload.sub as string;
    const email = newAccessTokenPayload.email as string;

    let refreshTokenPayload: JWTPayload;
    try {
        refreshTokenPayload = decodeJwt<JWTPayload>(refreshToken);
    } catch {
        return NextResponse.json({ error: "Invalid new refresh token" }, { status: 400 });
    }

    if (isExpiredSeconds(refreshTokenPayload.exp)) {
        return NextResponse.json({ error: "New refresh token expired" }, { status: 400 });
    }

    if (refreshTokenPayload.iss !== process.env.NEXT_PUBLIC_API_URL) {
        return NextResponse.json({ error: "Invalid new refresh token issuer" }, { status: 400 });
    }

    await saveSession({
        user: {
            id: userId,
            email,
        },
        accessToken: newAccessToken,
        refreshToken,
        refreshAt: DateTime.fromSeconds(newAccessTokenPayload.exp ?? 0).toMillis(),
        expiresAt: DateTime.fromSeconds(refreshTokenPayload.exp ?? 0).toMillis(),
    });

    return NextResponse.json({ success: true });
}
