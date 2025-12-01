import { UserService } from '@buf/hasir_hasir.bufbuild_es/user/v1/user_pb';
import { createConnectTransport } from '@connectrpc/connect-web';
import { NextRequest, NextResponse } from 'next/server';
import { createClient } from '@connectrpc/connect';
import { decodeJwt, JWTPayload } from "jose";

import { saveSession } from '@/lib/session';

const transport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  useBinaryFormat: true,
});

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { email, password } = body;

    const client = createClient(UserService, transport);
    const response = await client.login({
      email,
      password,
    });

    const accessTokenPayload = decodeJwt(response.accessToken) as JWTPayload;
    const userInfo = {
      id: accessTokenPayload.sub || '',
      email: accessTokenPayload.email || email,
    };

    const refreshTokenPayload = decodeJwt(response.refreshToken) as JWTPayload;

    const accessTokenExpiryMs = (accessTokenPayload.exp ?? 0) * 1000;
    const refreshTokenExpiryMs = (refreshTokenPayload.exp ?? 0) * 1000;

    await saveSession({
      user: userInfo,
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
      expiresAt: refreshTokenExpiryMs,
      refreshAt: accessTokenExpiryMs,
    });

    return NextResponse.json({
      user: userInfo,
      accessToken: response.accessToken,
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Login failed';
    return NextResponse.json(
      { error: errorMessage },
      { status: 401 }
    );
  }
}
