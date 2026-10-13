'use client';

/** Dedicated Template Detail page showing clean preview and template options. */
import { AppShell } from "@/components/AppShell";
import { CvPreview } from "@/components/CvPreview";
import { Button } from "@/components/ui/button";
import { api, ExampleSummary, Template } from "@/lib/api";
import { useLanguage } from "@/contexts/LanguageContext";
import { getToken, useAuthGuard } from "@/lib/auth";
import { ArrowLeft, Check, Copy, LayoutTemplate, Loader2, Plus } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";

export function TemplateDetailPage() {
  const router = useRouter();
  const { t, locale } = useLanguage();
  const params = useParams();
  const templateId = (params?.id as string) || "";

  const [template, setTemplate] = useState<Template | null>(null);
  const [examples, setExamples] = useState<ExampleSummary[]>([]);
  const [exampleDoc, setExampleDoc] = useState<{
    title: string;
    template_id: string;
    content: any;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [using, setUsing] = useState(false);
  const [mobilePreviewZoom, setMobilePreviewZoom] = useState<number | null>(null);
  const previewFrameRef = useRef<HTMLDivElement>(null);
  const token = getToken();
  useAuthGuard();

  useEffect(() => {
    if (!templateId) return;
    Promise.all([api.templatesPublic(), api.examples()])
      .then(([tList, eList]) => {
        const found = tList.find(t => t.id === templateId);
        setTemplate(found || null);
        setExamples(eList);

        const match = eList.find(x => x.template_id === templateId) || eList[0];
        if (match) {
          api
            .example(match.slug)
            .then(doc => setExampleDoc(doc))
            .catch(() => {});
        }
      })
      .catch(err =>
        toast.error(
          err instanceof Error ? err.message : "Could not load template details."
        )
      )
      .finally(() => setLoading(false));
  }, [templateId]);

  useEffect(() => {
    if (loading) return;
    const frame = previewFrameRef.current;
    if (!frame || typeof ResizeObserver === "undefined") return;

    const updatePreviewZoom = () => {
      if (window.innerWidth > 767) {
        setMobilePreviewZoom(null);
        return;
      }

      const styles = window.getComputedStyle(frame);
      const horizontalPadding = parseFloat(styles.paddingLeft) + parseFloat(styles.paddingRight);
      const availableWidth = frame.clientWidth - horizontalPadding;
      setMobilePreviewZoom(Math.min(1, Math.max(0.38, availableWidth / 680)));
    };

    const observer = new ResizeObserver(updatePreviewZoom);
    observer.observe(frame);
    updatePreviewZoom();
    return () => observer.disconnect();
  }, [loading]);

  async function startWithTemplate(withExample: boolean = false) {
    if (!template) return;
    setUsing(true);
    try {
      const cv = await api.createCV(token, {
        title: withExample && exampleDoc
          ? `${exampleDoc.title} CV`
          : `Untitled CV — ${template.name}`,
        template_id: template.id,
        document_style: "resume",
        target_region: "eac",
        language: locale,
        content: withExample && exampleDoc ? exampleDoc.content : undefined,
      });
      toast.success("CV created!");
      router.push(`/editor/${cv.id}`);
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Could not create CV with this template."
      );
    } finally {
      setUsing(false);
    }
  }

  if (loading) {
    return (
      <AppShell>
        <div className="flex h-96 items-center justify-center">
          <Loader2 className="animate-spin text-[#004AAD]" size={32} />
        </div>
      </AppShell>
    );
  }

  if (!template) {
    return (
      <AppShell>
        <div className="mx-auto max-w-4xl px-5 py-16 text-center">
          <h1 className="text-2xl font-extrabold text-[#171717]">
            {t("Template not found")}
          </h1>
          <Link href="/templates" className="mt-4 inline-block text-sm font-bold text-[#004AAD]">
            {t("← Back to all templates")}
          </Link>
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <main className="mx-auto max-w-7xl px-4 pb-20 pt-5 sm:px-5 sm:pt-8 lg:px-8">
        <div className="mb-5 flex flex-col items-stretch gap-4 border-b border-[#E5E7EB] pb-5 sm:mb-6 sm:flex-row sm:items-center sm:justify-between">
          <Link
            href="/templates"
            className="flex items-center text-sm font-bold text-[#737373] hover:text-[#171717]"
          >
            <ArrowLeft className="mr-1.5" size={16} />
            {t("Back to Templates")}
          </Link>
          <div className="grid grid-cols-1 gap-2 sm:flex sm:gap-3">
            <Button
              variant="outline"
              onClick={() => startWithTemplate(false)}
              disabled={using}
              className="w-full border-[#D4D4D4] text-[#004AAD] sm:w-auto"
            >
              <Plus className="mr-2" size={16} />
              {t("Start clean document")}
            </Button>
            <Button
              onClick={() => startWithTemplate(true)}
              disabled={using}
              className="w-full bg-[#004AAD] text-white hover:bg-[#003B8E] sm:w-auto"
            >
              {using ? (
                <Loader2 className="mr-2 animate-spin" size={16} />
              ) : (
                <Copy className="mr-2" size={16} />
              )}
              {t("Use template with sample data")}
            </Button>
          </div>
        </div>

        <div className="grid min-w-0 gap-5 lg:grid-cols-[300px_minmax(0,1fr)] lg:gap-8">
          <aside className="order-2 h-fit space-y-5 rounded-lg border border-[#E5E7EB] bg-white p-5 shadow-sm sm:p-6 lg:order-1 lg:space-y-6">
            <div className="flex items-center gap-3">
              <span
                className="inline-block h-10 w-10 rounded-lg shadow-sm"
                style={{ background: template.accent_color_hex }}
              />
              <div>
                <h1 className="text-xl font-extrabold text-[#171717]">
                  {template.name}
                </h1>
                <p className="text-xs capitalize text-[#737373]">
                  {t(template.family.replace("_", " "))} {t("Family")}
                </p>
              </div>
            </div>

            <div className="space-y-3 pt-2 text-sm text-[#525252]">
              <div className="flex items-center gap-2">
                <Check size={16} className="text-[#004AAD]" />
                <span>{t("ATS-Optimized Formatting")}</span>
              </div>
              <div className="flex items-center gap-2">
                <Check size={16} className="text-[#004AAD]" />
                <span>{template.supports_photo ? t("Photo & portrait layout") : t("Clean text layout")}</span>
              </div>
              <div className="flex items-center gap-2">
                <Check size={16} className="text-[#004AAD]" />
                <span>{t("Export clean PDF & DOCX")}</span>
              </div>
            </div>

            <div className="border-t border-[#E5E7EB] pt-5">
              <Button
                onClick={() => startWithTemplate(true)}
                disabled={using}
                className="w-full bg-[#004AAD] text-white hover:bg-[#003B8E]"
              >
                {t("Start with {name}").replace("{name}", template.name)}
              </Button>
            </div>
          </aside>

          <div ref={previewFrameRef} className="template-detail-preview order-1 min-w-0 overflow-hidden rounded-lg border border-[#E5E7EB] bg-white p-3 shadow-sm sm:p-5 lg:order-2 lg:p-6">
            <div className="mb-4 flex flex-col gap-2 border-b border-[#E5E7EB] pb-4 sm:flex-row sm:items-center sm:justify-between">
              <span className="text-xs font-bold tracking-wider text-[#737373] uppercase">
                {t("Rendered Template Preview")}
              </span>
              <span className="w-fit rounded-full bg-[#EAF7EE] px-3 py-1 text-[10px] font-extrabold text-[#17705B] sm:text-xs">
                {t("PDF/DOCX RENDER MATCH")}
              </span>
            </div>
            {exampleDoc ? (
              <div className="template-detail-paper" style={mobilePreviewZoom === null ? undefined : { zoom: mobilePreviewZoom }}>
                <CvPreview
                  content={exampleDoc.content}
                  title={exampleDoc.title}
                  template={template}
                />
              </div>
            ) : (
              <div className="flex h-96 items-center justify-center">
                <Loader2 className="animate-spin text-[#004AAD]" size={28} />
              </div>
            )}
          </div>
        </div>
      </main>
    </AppShell>
  );
}
