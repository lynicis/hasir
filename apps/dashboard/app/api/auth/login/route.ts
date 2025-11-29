import { UserService } from '@buf/hasir_hasir.bufbuild_es/user/v1/user_pb';
import { createConnectTransport } from '@connectrpc/connect-web';
import { NextRequest, NextResponse } from 'next/server';
import { createClient } from '@connectrpc/connect';
import { decodeJwt } from "jose";

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

    const payload = decodeJwt(response.accessToken);
    const userInfo = {
      id: payload.sub || '',
      email: payload.email || email,
    };

    await saveSession({
      user: userInfo,
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
      expiresAt: Date.now() + (7 * 24 * 60 * 60 * 1000),
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