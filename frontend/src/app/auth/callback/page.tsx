"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { session } from "@/lib/session";
import { API_BASE_URL } from "@/lib/api";
import { BrandMark } from "@/components/BrandMark";

function AuthCallbackContent() {
  const params = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState("");

  useEffect(() => {
    const code = params.get("code");
    const state = params.get("state");
    const expectedState = sessionStorage.getItem("tengenezacv_oauth_state");
    const verifier = sessionStorage.getItem("tengenezacv_pkce_verifier");
    const returnTo =
      sessionStorage.getItem("tengenezacv_return_to") || "/workspace";
    if (!code || !state || state !== expectedState || !verifier) {
      setError(
        "This sign-in link is invalid or has expired. Please try again."
      );
      return;
    }
    const redirectUri = `${window.location.origin}/auth/callback`;
    fetch("/api/auth/callback", {
      method: "POST",
      headers: {
        "content-type": "application/json",
        accept: "application/json",
      },
      body: JSON.stringify({
        code,
        code_verifier: verifier,
        redirect_uri: redirectUri,
      }),
    })
      .then(async response => {
        const payload = await response.json();
        if (!response.ok || !payload.access_token)
          throw new Error(
            payload.error_description ||
              payload.message ||
              "CamelAccounts could not complete sign-in."
          );
        const meResponse = await fetch(`${API_BASE_URL}/me`, {
          headers: { Authorization: `Bearer ${payload.access_token}` },
        });
        const mePayload = await meResponse.json();
        const user = mePayload.data || mePayload.user;
        session.set({
          token: payload.access_token,
          refreshToken: payload.refresh_token,
          user: user || { id: "camel-account", full_name: "", email: "" },
        });
        if (user?.id != null) {
          window.camel?.identify(String(user.id));
        }
        sessionStorage.removeItem("tengenezacv_pkce_verifier");
        sessionStorage.removeItem("tengenezacv_oauth_state");
        router.replace(returnTo.startsWith("/") ? returnTo : "/workspace");
      })
      .catch(reason =>
        setError(reason instanceof Error ? reason.message : "Sign-in failed.")
      );
  }, [params, router]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#F7F7F7] px-6">
      <div className="max-w-md rounded-2xl border border-[#E5E7EB] bg-white p-8 text-center shadow-xl">
        <BrandMark compact className="mb-6 justify-center" />
        <div className="mx-auto mb-5 h-10 w-10 animate-spin rounded-full border-4 border-[#E5E7EB] border-t-[#004AAD]" />
        <h1 className="text-xl font-extrabold text-[#171717]">
          {error ? "Sign-in needs attention" : "Securing your workspace"}
        </h1>
        <p className="mt-3 text-sm leading-6 text-[#525252]">
          {error || "Finishing your CamelAccounts sign-in…"}
        </p>
        {error && (
          <button
            className="mt-6 rounded-lg bg-[#004AAD] px-5 py-3 text-sm font-bold text-white"
            onClick={() => router.replace("/login")}
          >
            Return to sign in
          </button>
        )}
      </div>
    </main>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <main className="flex min-h-screen items-center justify-center bg-[#F7F7F7] px-6">
          <p className="text-sm font-semibold text-[#525252]">
            Securing your workspace…
          </p>
        </main>
      }
    >
      <AuthCallbackContent />
    </Suspense>
  );
}
