'use client';

import { PublicHeader } from "@/components/AppShell";
import { BrandMark } from "@/components/BrandMark";
import { Button } from "@/components/ui/button";
import { api, type Plan } from "@/lib/api";
import { ArrowRight, Coins, Download, Loader2, Mic2, Sparkles } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

const money = (amount?: number) => new Intl.NumberFormat("en-TZ", {
  style: "currency", currency: "TZS", maximumFractionDigits: 0,
}).format(amount ?? 0);
const number = (value?: number) => new Intl.NumberFormat("en-TZ").format(value ?? 0);

export function PricingPage() {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.plans().then(setPlans).catch(e => setError(e instanceof Error ? e.message : "Unable to load plans."))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-[#F7F7F7]">
      <PublicHeader />
      <main className="mx-auto max-w-7xl px-5 pb-20 lg:px-8">
        <section className="border-b border-[#E5E7EB] pb-10 pt-12">
          <p className="ledger-label"><span /> CLEAR, LOCAL PRICING</p>
          <h1 className="mt-4 max-w-3xl text-4xl font-semibold tracking-tight text-[#171717] sm:text-5xl">
            Build a CV that moves your career forward.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[#525252]">
            Create and improve professional CVs for Tanzania and the wider region. Plans show the export, AI, and voice balances provided by the CV service.
          </p>
          <Link href="/workspace"><Button size="lg" className="mt-7 bg-[#004AAD] text-white hover:bg-[#003B8E]">Open your workspace <ArrowRight className="ml-2" size={18} /></Button></Link>
        </section>

        <section className="mt-10" aria-live="polite">
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div><p className="ledger-label"><span /> TENGENEZA CV PLANS</p><h2 className="mt-2 text-2xl font-semibold text-[#171717]">Choose the balance that fits your work</h2></div>
            <p className="text-sm text-[#737373]">Prices in Tanzanian shillings (TZS)</p>
          </div>
          {loading ? <div className="flex min-h-52 items-center justify-center text-[#004AAD]"><Loader2 className="animate-spin" aria-label="Loading plans" /></div> : error ? <div role="alert" className="mt-6 rounded-lg border border-red-200 bg-white p-5 text-sm text-red-800">{error}</div> : (
            <div className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {plans.map(plan => <article key={plan.code} className="flex flex-col rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
                <div className="flex items-start justify-between gap-3"><div><p className="text-xs font-bold uppercase tracking-[.12em] text-[#004AAD]">{plan.subscription ? "Subscription" : plan.code === "free" ? "Start here" : "One-time pack"}</p><h3 className="mt-2 text-xl font-semibold text-[#171717]">{plan.name}</h3></div><span className="rounded-full bg-[#EAF2FF] p-2 text-[#004AAD]"><Coins size={18} /></span></div>
                <p className="mt-5 text-3xl font-semibold tracking-tight text-[#171717]">{plan.price_tzs === 0 ? "Free" : money(plan.price_tzs)}{plan.subscription && <span className="ml-1 text-sm font-medium text-[#737373]">/ month</span>}</p>
                <div className="my-5 h-px bg-[#E5E7EB]" />
                <ul className="flex-1 space-y-3 text-sm text-[#525252]">
                  <li className="flex gap-2"><Download size={16} className="mt-0.5 shrink-0 text-[#004AAD]" />{plan.export_credits ? `${number(plan.export_credits)} document exports` : "No exports included"}</li>
                  <li className="flex gap-2"><Sparkles size={16} className="mt-0.5 shrink-0 text-[#004AAD]" />{plan.ai_tokens ? `${number(plan.ai_tokens)} AI tokens` : "No AI tokens included"}</li>
                  <li className="flex gap-2"><Mic2 size={16} className="mt-0.5 shrink-0 text-[#004AAD]" />{plan.voice_seconds ? `${number(plan.voice_seconds)} voice seconds` : "No voice seconds included"}</li>
                </ul>
                <p className="mt-5 rounded-md bg-[#F7F7F7] p-3 text-xs leading-5 text-[#737373]">{plan.description || (plan.subscription ? "Recurring plan balance." : "Plan balance details from the CV service.")}</p>
              </article>)}
            </div>
          )}
          <p className="mt-5 rounded-lg border border-[#E5E7EB] bg-white px-4 py-3 text-sm leading-6 text-[#525252]">Plan balances are shown for transparency. Purchasing or changing plans isn’t available in this app yet.</p>
        </section>
      </main>
      <footer className="border-t border-[#E5E7EB] bg-white py-8 text-center text-sm text-[#737373]"><BrandMark compact /></footer>
    </div>
  );
}
