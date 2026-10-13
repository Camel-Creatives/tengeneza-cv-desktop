"use client";

import { CamelAuthProvider, CamelLoginButton, useCamelAuth } from "camel-accounts";
import "camel-accounts/style.css";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { API_BASE_URL, type User } from "@/lib/api";
import { session } from "@/lib/session";

function CamelLoginContent() {
  const { login, authenticated, loading, user, getAccessToken } = useCamelAuth();
  const router = useRouter();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!authenticated || !user) return;
    const token = getAccessToken();
    if (!token) return;
    fetch(`${API_BASE_URL}/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async response => {
        const payload = await response.json();
        if (!response.ok) throw new Error(payload.message || "Could not load your account.");
        const appUser = (payload.data || payload.user || {}) as User;
        const camelUser = user as Record<string, any>;
        const camelEmail = camelUser.email || camelUser.email_address || camelUser.emailAddress || camelUser.profile?.email || "";
        // Camel is the source of identity truth. Do not let an existing but
        // empty API row mask verified name/email claims.
        const mergedUser: User = {
          ...appUser,
          id: appUser.id || user.sub,
          full_name: appUser.full_name || user.name || "",
          email: appUser.email || camelEmail,
          phone: appUser.phone || (user as { phone?: string }).phone || "",
        };
        if ((!appUser.full_name && mergedUser.full_name) || (!appUser.email && mergedUser.email) || (!appUser.phone && mergedUser.phone)) {
          await fetch(`${API_BASE_URL}/me`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
            body: JSON.stringify({ full_name: mergedUser.full_name }),
          });
        }
        session.set({ token, user: mergedUser });
        window.camel?.identify(String(mergedUser.id));
        router.replace("/workspace");
      })
      .catch(reason => setError(reason instanceof Error ? reason.message : "Sign-in failed."));
  }, [authenticated, getAccessToken, router, user]);

  if (error) return <p role="alert" className="rounded-md bg-red-50 p-3 text-center text-sm font-semibold text-red-700">{error}</p>;
  if (loading || authenticated) return <p className="text-center text-sm font-semibold text-[#525252]">{authenticated ? "Opening your workspace…" : "Checking your sign-in…"}</p>;
  return <CamelLoginButton onClick={login} label="Continue with CamelAccounts" />;
}

export function CamelPopupLogin() {
  return (
    <CamelAuthProvider
      config={{
        issuer: process.env.NEXT_PUBLIC_CAMEL_ACCOUNTS_URL || "https://accounts.camelcreatives.com",
        clientId: process.env.NEXT_PUBLIC_CAMEL_ACCOUNTS_CLIENT_ID || "tengenezacv-web",
        scopes: ["openid", "profile", "email"],
      }}
    >
      <CamelLoginContent />
    </CamelAuthProvider>
  );
}
