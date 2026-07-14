import { NextRequest, NextResponse } from "next/server";

import { getSession, isExpiredMillis } from "@/lib/session";

type RouteParams = {
  organizationId: string;
  repositoryId: string;
  commitHash: string;
};

async function validateSession() {
  const session = await getSession();

  if (!session.user) {
    return { error: "Unauthenticated", status: 401 };
  }

  if (isExpiredMillis(session.expiresAt)) {
    session.destroy();
    return { error: "Session expired", status: 401 };
  }

  if (!session.accessToken) {
    return { error: "No access token", status: 401 };
  }

  return { session };
}

function getBackendUrl(): string | null {
  const baseUrl = process.env.NEXT_PUBLIC_API_URL ?? "";
  return baseUrl.replace(/\/$/, "") || null;
}

async function fetchBackendDocs(url: string, accessToken: string) {
  const response = await fetch(url, {
    headers: {
      Accept: "text/markdown, text/plain, */*",
      Authorization: `Bearer ${accessToken}`,
    },
    cache: "no-store",
  });

  if (!response.ok) {
    let details: string | undefined;
    try {
      const contentType = response.headers.get("content-type");
      if (contentType?.includes("application/json")) {
        const json = await response.json();
        details = json.error ?? json.message ?? JSON.stringify(json);
      } else {
        details = await response.text();
      }
    } catch {}

    return { error: details ?? `Backend error (${response.status})`, status: response.status };
  }

  return { content: await response.text() };
}

export async function GET(
  _request: NextRequest,
  context: { params: Promise<RouteParams> }
) {
  const { organizationId, repositoryId, commitHash } = await context.params;

  if (!organizationId || !repositoryId || !commitHash) {
    return NextResponse.json(
      { error: "Missing required parameters" },
      { status: 400 }
    );
  }

  const authResult = await validateSession();
  if ("error" in authResult) {
    return NextResponse.json(
      { error: authResult.error },
      { status: authResult.status }
    );
  }

  const backendUrl = getBackendUrl();
  if (!backendUrl) {
    return NextResponse.json(
      { error: "API URL not configured" },
      { status: 500 }
    );
  }

  const docsUrl = `${backendUrl}/docs/${organizationId}/${repositoryId}/${commitHash}`;

  try {
    const result = await fetchBackendDocs(docsUrl, authResult.session.accessToken!);

    if ("error" in result) {
      return NextResponse.json(
        { error: result.error },
        { status: result.status }
      );
    }

    return new NextResponse(result.content, {
      headers: { "Content-Type": "text/markdown; charset=utf-8" },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    const isNetworkError = error instanceof TypeError && message.includes("fetch");

    return NextResponse.json(
      { error: isNetworkError ? "Failed to connect to backend" : message },
      { status: isNetworkError ? 502 : 500 }
    );
  }
}
