import { NextRequest, NextResponse } from "next/server";

const camelAccountsUrl = (process.env.CAMEL_ACCOUNTS_URL || process.env.NEXT_PUBLIC_CAMEL_ACCOUNTS_URL || "https://accounts.camelcreatives.com").replace(/\/$/, "");
const clientId = process.env.CAMEL_ACCOUNTS_CLIENT_ID || process.env.NEXT_PUBLIC_CAMEL_ACCOUNTS_CLIENT_ID || "tengenezacv-web";

export async function POST(request: NextRequest) {
  const body = await request.json().catch(() => null) as { code?: string; code_verifier?: string } | null;
  if (!body?.code || !body.code_verifier) return NextResponse.json({ message: "Missing OAuth authorization data." }, { status: 400 });
  const redirectUri = `${request.nextUrl.origin}/auth/callback`;
  try {
    const response = await fetch(`${camelAccountsUrl}/oauth/token`, {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded", accept: "application/json" },
      body: new URLSearchParams({ grant_type: "authorization_code", client_id: clientId, redirect_uri: redirectUri, code: body.code, code_verifier: body.code_verifier }),
      cache: "no-store",
    });
    const payload = await response.json();
    if (!response.ok || !payload.access_token) return NextResponse.json({ message: payload.error_description || payload.message || "CamelAccounts token exchange failed." }, { status: response.status || 502 });
    return NextResponse.json({ access_token: payload.access_token, refresh_token: payload.refresh_token, expires_in: payload.expires_in });
  } catch {
    return NextResponse.json({ message: "Authentication service is temporarily unavailable." }, { status: 503 });
  }
}
