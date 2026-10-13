'use client';

/** Nairobi Blue Ledger: template gallery where clicking a template opens its detail page. */
import { AppShell } from "@/components/AppShell";
import { CvPreview } from "@/components/CvPreview";
import { Button } from "@/components/ui/button";
import { api, ExampleSummary, Template } from "@/lib/api";
import { useLanguage } from "@/contexts/LanguageContext";
import { getToken, useAuthGuard } from "@/lib/auth";
import { LayoutTemplate, Loader2, Sparkles, ArrowRight } from "lucide-react";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

export function TemplatesPage() {
  const router = useRouter();
  const { t: tr, locale } = useLanguage();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [examples, setExamples] = useState<ExampleSummary[]>([]);
  const [docsBySlug, setDocsBySlug] = useState<
    Record<string, { title: string; template_id: string; content: any }>
  >({});
  const [loading, setLoading] = useState(true);
  const [using, setUsing] = useState<string | null>(null);
  const token = getToken();
  useAuthGuard();

  useEffect(() => {
    Promise.all([api.templatesPublic(), api.examples()])
      .then(([t, e]) => {
        setTemplates(t);
        setExamples(e);
        Promise.all(
          e.map(x =>
            api
              .example(x.slug)
              .then(d => [x.slug, d] as const)
              .catch(() => null)
          )
        ).then(entries => {
          const map: typeof docsBySlug = {};
          for (const entry of entries) if (entry) map[entry[0]] = entry[1];
          setDocsBySlug(map);
        });
      })
      .catch(err =>
        toast.error(
          err instanceof Error ? err.message : "Could not load templates."
        )
      )
      .finally(() => setLoading(false));
  }, []);

  async function useTemplate(template_id: string, name?: string) {
    setUsing(template_id);
    try {
      const cv = await api.createCV(token, {
        title: `Untitled CV — ${name || template_id}`,
        template_id,
        document_style: "resume",
        target_region: "eac",
        language: locale,
      });
      toast.success("CV created!");
      router.push(`/editor/${cv.id}`);
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Could not create CV with this template."
      );
    } finally {
      setUsing(null);
    }
  }

  return (
    <AppShell>
      <main className="mx-auto max-w-7xl px-5 pb-20 pt-8 lg:px-8">
        <div className="flex flex-wrap items-center justify-between gap-4 rounded-lg border border-[#E5E7EB] bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#EAF2FF] text-[#004AAD]">
              <Sparkles size={20} />
            </div>
            <div>
              <b className="block text-base font-extrabold text-[#171717]">
                {tr("Looking for Example Documents?")}
              </b>
              <p className="text-xs text-[#737373]">
                {tr("Explore 24 live API sample documents with filled career content.")}
              </p>
            </div>
          </div>
          <Link
            href="/examples"
            className="flex items-center rounded-lg bg-[#004AAD] px-4 py-2 text-sm font-bold text-white hover:bg-[#003B8E]"
          >
            {tr("View Example Documents")}
            <ArrowRight className="ml-1.5" size={16} />
          </Link>
        </div>

        <section className="mt-8">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-extrabold text-[#171717]">
              <LayoutTemplate className="mr-2 inline text-[#004AAD]" size={22} />
              {tr("All templates")} ({templates.length})
            </h2>
            <span className="rounded-full bg-[#EAF7EE] px-3 py-1 text-xs font-extrabold text-[#17705B]">
              {tr("ALL UNLOCKED · FREE EXPORTS")}
            </span>
          </div>

          {loading ? (
            <div className="mt-16 flex justify-center">
              <Loader2 className="animate-spin text-[#004AAD]" size={32} />
            </div>
          ) : (
            <div className="mt-8 grid gap-7 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3">
              {templates.map(t => {
                const slug = examples.find(x => x.template_id === t.id)?.slug;
                const doc = slug ? docsBySlug[slug] : null;
                return (
                  <article
                    key={t.id}
                    className="group overflow-hidden rounded-lg border border-[#E5E7EB] bg-white shadow-xs transition duration-200 hover:-translate-y-0.5 hover:border-[#004AAD]/40 hover:shadow-md flex flex-col justify-between"
                  >
                    <button
                      onClick={() => router.push(`/templates/${t.id}`)}
                      className="template-thumb block w-full text-left cursor-pointer relative"
                      title={tr("View {name} template details & live preview").replace("{name}", t.name)}
                    >
                      {doc ? (
                        <CvPreview
                          content={doc.content}
                          title={doc.title}
                          template={t}
                        />
                      ) : (
                        <div
                          className="h-48 w-full"
                          style={{
                            background: `linear-gradient(120deg, ${t.accent_color_hex}, ${t.accent_color_hex}CC)`,
                          }}
                        />
                      )}
                    </button>
                    <div className="p-4 border-t border-[#EFF5FD] bg-white">
                      <button
                        onClick={() => router.push(`/templates/${t.id}`)}
                        className="text-left w-full group-hover:text-[#004AAD] transition"
                      >
                        <div className="flex items-center justify-between">
                          <h3 className="font-extrabold text-[#171717] text-lg">{t.name}</h3>
                          <span
                            className="inline-block h-3.5 w-3.5 rounded-full border border-white shadow-xs"
                            style={{ background: t.accent_color_hex || "#004AAD" }}
                            title={tr("Accent color: {color}").replace("{color}", t.accent_color_hex)}
                          />
                        </div>
                        <p className="mt-0.5 text-xs capitalize text-[#737373]">
                          {tr(t.family.replace("_", " "))} {tr("Family")}
                          {t.supports_photo ? ` · ${tr("photo-ready")}` : ""}
                        </p>
                      </button>
                      <div className="mt-3.5 flex gap-2">
                        <Button
                          onClick={() => router.push(`/templates/${t.id}`)}
                          size="sm"
                          variant="outline"
                          className="w-1/2 border-[#E5E7EB] text-[#171717] hover:bg-[#F5F5F5] font-bold text-xs"
                        >
                          {tr("Preview")}
                        </Button>
                        <Button
                          onClick={() => useTemplate(t.id, t.name)}
                          disabled={using === t.id}
                          size="sm"
                          className="w-1/2 bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold text-xs shadow-xs"
                        >
                          {using === t.id ? (
                            <Loader2 className="animate-spin" size={14} />
                          ) : (
                            tr("Use design")
                          )}
                        </Button>
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </section>
      </main>
    </AppShell>
  );
}
