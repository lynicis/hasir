import { NextResponse } from 'next/server';

import { getSession } from '@/lib/session';

export async function GET() {
  const session = await getSession();

  if (!session.user) {
    return NextResponse.json({ user: null }, { status: 401 });
  }

  if (session.expiresAt && session.expiresAt < Date.now()) {
    session.destroy();
    return NextResponse.json({ user: null }, { status: 401 });
  }

  return NextResponse.json({
    user: session.user,
    accessToken: session.accessToken,
  });
}