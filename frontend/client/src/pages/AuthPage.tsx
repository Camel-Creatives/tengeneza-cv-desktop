"use client";

/** Nairobi Blue Ledger: authentication is focused, calm, and explicit about the next step. */
import { BrandMark } from "@/components/BrandMark";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Eye,
  EyeOff,
} from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import { api } from "@/lib/api";
import { session } from "@/lib/session";
import { toast } from "sonner";
import { createCamelAuthorizePath, startCamelSocialLogin } from "@/lib/auth";
import Link from "next/link";
import { useRouter } from "next/navigation";

function Field({
  label,
  htmlFor,
  children,
}: {
  label: string;
  htmlFor: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={htmlFor} className="font-bold text-[#171717]">{label}</Label>
      {children}
    </div>
  );
}

export function AuthPage({ mode }: { mode: "login" | "signup" }) {
  const signup = mode === "signup";
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [providers, setProviders] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [working, setWorking] = useState(false);
  const [activeProvider, setActiveProvider] = useState("");
  const camelAccountsUrl = (process.env.NEXT_PUBLIC_CAMEL_ACCOUNTS_URL || "https://accounts.camelcreatives.com").replace(/\/$/, "");

  useEffect(() => {
    let mounted = true;
    fetch(`${camelAccountsUrl}/auth/social/providers`, { headers: { accept: "application/json" } })
      .then(async response => {
        const payload = await response.json();
        const available = payload.data?.providers || payload.providers || [];
        if (mounted && response.ok) setProviders(available.filter((provider: string) => provider === "google"));
      })
      .catch(() => { if (mounted) setProviders([]); });
    return () => { mounted = false; };
  }, [camelAccountsUrl]);

  async function submitAccount(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setWorking(true);
    try {
      const authorizePath = await createCamelAuthorizePath("/workspace");
      if (!authorizePath) throw new Error("Could not start secure sign-in. Please try again.");
      const form = document.createElement("form");
      form.method = "POST";
      form.action = `${camelAccountsUrl}/auth/${signup ? "register" : "login"}/continue`;
      form.hidden = true;
      const fields = signup
        ? { email, password, next: authorizePath }
        : { identifier: email, password, next: authorizePath };
      for (const [name, value] of Object.entries(fields)) {
        const field = document.createElement("input");
        field.type = "hidden";
        field.name = name;
        field.value = value;
        form.appendChild(field);
      }
      document.body.appendChild(form);
      form.submit();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not complete sign-in.");
      setWorking(false);
    }
  }

  async function continueWithSocial(provider: string) {
    setError("");
    setActiveProvider(provider);
    try {
      await startCamelSocialLogin(provider, "/workspace");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not start social sign-in.");
      setActiveProvider("");
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#F7F7F7] px-5 py-10">
      <div className="w-full max-w-[430px]">
        <div className="mb-10 flex justify-center">
          <Link href="/">
            <BrandMark className="auth-brand" />
          </Link>
        </div>
        <div className="rounded-lg border border-[#E5E7EB] bg-white p-7 shadow-sm sm:p-9">
          <div className="mb-8 text-center">
            <h1 className="text-2xl font-extrabold tracking-tight text-[#171717]">
              {signup ? "Create your workspace" : "Welcome back"}
            </h1>
            <p className="mt-2 text-sm leading-6 text-[#737373]">
              {signup
                ? "Build a professional CV with your CamelAccounts identity."
                : "Sign in securely to continue to your CV workspace."}
            </p>
          </div>
          <div className="space-y-3">
            {providers.map(provider => (
              <button key={provider} type="button" onClick={() => void continueWithSocial(provider)} disabled={working || !!activeProvider} className="flex h-11 w-full items-center justify-center gap-3 rounded-lg border border-[#D1D5DB] bg-white text-sm font-semibold text-[#171717] shadow-sm transition hover:border-[#9CA3AF] hover:bg-[#FAFAFA] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#004AAD] disabled:opacity-60">
                {provider === "google" ? <GoogleMark /> : <GitHubMark />}
                {activeProvider === provider ? "Connecting…" : `Continue with ${provider === "google" ? "Google" : "GitHub"}`}
              </button>
            ))}
          </div>
          {providers.length > 0 && <div className="my-5 flex items-center gap-3 text-xs text-[#737373]"><span className="h-px flex-1 bg-[#E5E7EB]" /><span>or continue with email</span><span className="h-px flex-1 bg-[#E5E7EB]" /></div>}
          <form className="space-y-4" onSubmit={submitAccount}>
            <Field label="Email address" htmlFor="account-email">
              <Input id="account-email" type="email" autoComplete="email" required value={email} onChange={event => setEmail(event.target.value)} placeholder="you@example.com" className="h-12 border-[#AEBFD2] bg-[#F8FAFC] px-4 text-[15px] text-[#171717] placeholder:text-[#667085] shadow-[inset_0_1px_2px_rgba(16,42,86,0.04)] hover:border-[#899CB2] focus-visible:border-[#004AAD] focus-visible:ring-[#004AAD]/30" />
            </Field>
            <Field label="Password" htmlFor="account-password">
              <div className="relative">
                <Input id="account-password" type={showPassword ? "text" : "password"} autoComplete={signup ? "new-password" : "current-password"} minLength={signup ? 8 : undefined} required value={password} onChange={event => setPassword(event.target.value)} placeholder={signup ? "At least 8 characters" : "Enter your password"} className="h-12 border-[#AEBFD2] bg-[#F8FAFC] px-4 pr-11 text-[15px] text-[#171717] placeholder:text-[#667085] shadow-[inset_0_1px_2px_rgba(16,42,86,0.04)] hover:border-[#899CB2] focus-visible:border-[#004AAD] focus-visible:ring-[#004AAD]/30" />
                <button type="button" onClick={() => setShowPassword(value => !value)} aria-label={showPassword ? "Hide password" : "Show password"} className="absolute inset-y-0 right-0 flex w-11 items-center justify-center text-[#737373] hover:text-[#171717]">{showPassword ? <Eye size={17} /> : <EyeOff size={17} />}</button>
              </div>
            </Field>
            {!signup && <div className="-mt-1 text-right"><a href={`${camelAccountsUrl}/forgot-password`} className="text-xs font-semibold text-[#004AAD] hover:underline">Forgot password?</a></div>}
            {error && <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm leading-5 text-red-700">{error}</p>}
            <Button type="submit" disabled={working || !!activeProvider} className="h-11 w-full rounded-lg bg-[#004AAD] text-sm font-bold text-white shadow-sm hover:bg-[#003B8E]">
              {working ? "Please wait…" : <>{signup ? "Create account" : "Sign in"}<ArrowRight size={16} className="ml-2" /></>}
            </Button>
          </form>
          <p className="mt-7 text-center text-xs leading-5 text-[#525252]">
            Securely managed by Camel Accounts. Your Tengeneza CV workspace uses your verified account identity.
          </p>
        </div>
        <p className="mt-5 text-center text-sm text-[#525252]">
          {signup ? "Already have an account?" : "New to Tengeneza CV?"} {" "}
          <Link href={signup ? "/login" : "/signup"} className="font-bold text-[#004AAD] hover:underline">{signup ? "Sign in" : "Create an account"}</Link>
        </p>
        <Link
          href="/"
          className="mt-6 flex items-center justify-center gap-2 text-sm font-semibold text-[#737373] transition hover:text-[#004AAD]"
        >
          <ArrowLeft size={15} /> Back to Tengeneza CV
        </Link>
      </div>
    </div>
  );
}

function GoogleMark() {
  return <svg aria-hidden="true" viewBox="0 0 18 18" className="h-[18px] w-[18px]"><path fill="#4285F4" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62Z"/><path fill="#34A853" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.91-2.26c-.81.54-1.84.86-3.05.86-2.34 0-4.33-1.58-5.04-3.71H.96v2.33A9 9 0 0 0 9 18Z"/><path fill="#FBBC05" d="M3.96 10.71a5.41 5.41 0 0 1 0-3.42V4.96H.96a9 9 0 0 0 0 8.08l3-2.33Z"/><path fill="#EA4335" d="M9 3.58c1.32 0 2.51.45 3.44 1.35l2.58-2.59A9 9 0 0 0 .96 4.96l3 2.33C4.67 5.16 6.66 3.58 9 3.58Z"/></svg>;
}

function GitHubMark() {
  return <svg aria-hidden="true" viewBox="0 0 24 24" className="h-[18px] w-[18px] fill-[#181717]"><path d="M12 .3a12 12 0 0 0-3.79 23.39c.6.11.82-.26.82-.58v-2.03c-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.33-1.76-1.33-1.76-1.09-.75.08-.73.08-.73 1.2.08 1.84 1.24 1.84 1.24 1.07 1.83 2.81 1.3 3.5 1 .1-.78.42-1.31.76-1.61-2.66-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.13-.3-.54-1.52.11-3.18 0 0 1.01-.32 3.3 1.23a11.5 11.5 0 0 1 6.01 0c2.29-1.55 3.29-1.23 3.29-1.23.66 1.66.25 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.63-5.49 5.92.43.38.82 1.11.82 2.24v3.32c0 .32.21.7.82.58A12 12 0 0 0 12 .3Z"/></svg>;
}

export function VerifyOtpPage() {
  const router = useRouter();
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [working, setWorking] = useState(false);
  const [resending, setResending] = useState(false);
  const [devOtp, setDevOtp] = useState("");
  const [destination, setDestination] = useState("");

  useEffect(() => {
    if (typeof window !== "undefined") {
      setDevOtp(sessionStorage.getItem("tengenezacv_development_otp") || "");
      setDestination(
        sessionStorage.getItem("tengenezacv_destination") ||
          "your email address"
      );
    }
  }, []);

  async function submit(e: FormEvent) {
    e.preventDefault();
    const v =
      typeof window !== "undefined"
        ? sessionStorage.getItem("tengenezacv_verification_id")
        : null;
    if (!v)
      return setError(
        "Your verification session is missing. Please sign up again."
      );
    setWorking(true);
    try {
      const d = await api.verifyOtp({ verification_id: v, code });
      if (typeof window !== "undefined") {
        sessionStorage.removeItem("tengenezacv_development_otp");
      }
      session.set({
        token: d.access_token,
        refreshToken: d.refresh_token,
        user: d.user,
      });
      router.push("/workspace");
    } catch (x) {
      setError(x instanceof Error ? x.message : "Code could not be verified.");
    } finally {
      setWorking(false);
    }
  }

  async function resend() {
    const v =
      typeof window !== "undefined"
        ? sessionStorage.getItem("tengenezacv_verification_id")
        : null;
    if (!v)
      return setError(
        "Your verification session is missing. Please sign up again."
      );
    setResending(true);
    setError("");
    try {
      const d = await api.resendOtp({ verification_id: v });
      if (typeof window !== "undefined") {
        sessionStorage.setItem(
          "tengenezacv_verification_id",
          d.verification_id
        );
        if (d.development_otp) {
          sessionStorage.setItem(
            "tengenezacv_development_otp",
            d.development_otp
          );
          setDevOtp(d.development_otp);
        } else {
          sessionStorage.removeItem("tengenezacv_development_otp");
          setDevOtp("");
        }
      }
      toast.success("A new code is on its way.");
    } catch (x) {
      setError(x instanceof Error ? x.message : "Could not resend the code.");
    } finally {
      setResending(false);
    }
  }

  return (
    <div className="min-h-screen bg-[#F7F7F7] p-7">
      <div className="mx-auto max-w-md">
        <Link href="/">
          <BrandMark className="auth-brand" />
        </Link>
        <div className="mt-16 rounded-lg border border-[#E5E7EB] bg-white p-8 shadow-xl">
          <CheckCircle2 className="text-[#004AAD]" />
          <p className="ledger-label mt-6">
            <span /> VERIFY YOUR CONTACT
          </p>
          <h1 className="mt-3 font-serif text-4xl text-[#171717]">
            Enter your six-digit code.
          </h1>
          <p className="mt-3 leading-7 text-[#525252]">
            We emailed a verification code to{" "}
            <span className="font-bold text-[#171717]">{destination}</span>.
            Enter it below to continue.
          </p>
          <form className="mt-8 space-y-5" onSubmit={submit}>
            {devOtp && (
              <p className="rounded-md border border-dashed border-[#D4D4D4] bg-[#F7F7F7] p-3 text-xs font-bold text-[#525252]">
                Development mode: your verification code is{" "}
                <b className="font-mono text-[#004AAD]">{devOtp}</b>
              </p>
            )}
            <Input
              required
              inputMode="numeric"
              maxLength={6}
              value={code}
              onChange={e => setCode(e.target.value.replace(/\D/g, ""))}
              className="h-14 text-center font-mono text-xl tracking-[.5em]"
              placeholder="000000"
            />
            {error && <p className="text-sm font-bold text-red-600">{error}</p>}
            <Button
              disabled={working || code.length !== 6}
              className="w-full bg-[#004AAD] text-white hover:bg-[#003B8E]"
            >
              {working ? "Verifying…" : "Verify and enter workspace"}
            </Button>
            <button
              type="button"
              onClick={resend}
              disabled={resending}
              className="w-full text-center text-sm font-bold text-[#525252] transition-colors hover:text-[#004AAD] disabled:opacity-60"
            >
              {resending ? "Sending…" : "Didn't get it? Resend the code"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
