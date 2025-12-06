import { getIronSession, IronSession, SessionOptions } from 'iron-session';
import { cookies } from 'next/headers';
import { DateTime } from "luxon";

export type SessionData = Partial<{
  user: {
    id: string;
    email: string;
  };
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  refreshAt: number;
}>;

const sessionOptions: SessionOptions = {
  password: process.env.SESSION_SECRET || "",
  cookieName: 'hasir-session',
  ttl: 60 * 60 * 24 * 7,
  cookieOptions: {
    secure: process.env.NODE_ENV === "production",
    httpOnly: true,
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7,
  },
};

export async function getSession() {
  const cookieStore = await cookies();
  return getIronSession<SessionData>(cookieStore, sessionOptions);
}

export async function saveSession(data: SessionData) {
  const session = await getSession();
  Object.assign(session, data);
  await session.save();
  return session;
}

export async function destroySession() {
  const session = await getSession();
  session.destroy();
}

export async function refreshSession(session: IronSession<SessionData>) {
  await session.save();
}

export function isExpiredSeconds(exp?: number): boolean {
  if (!exp) return true;
  return DateTime.utc() > DateTime.fromSeconds(exp);
}

export function isExpiredMillis(timestamp?: number): boolean {
  if (!timestamp) return true;
  return DateTime.utc() > DateTime.fromMillis(timestamp);
}

